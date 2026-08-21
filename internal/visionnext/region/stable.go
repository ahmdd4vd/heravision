package region

import (
	"fmt"
	"image"
	"math"
	"sort"

	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/imageview"
	"heravision/internal/visionnext/schema"
)

// StableConfig controls the deterministic multi-scale proposal ablation. The
// base proposer remains available through Propose so B1-old is reproducible.
type StableConfig struct {
	Base               Config
	ScaleFractions     []float64
	MatchIoU           float64
	MinSupport         int
	BoundaryTolerance  float64
	ExtraScaleFraction float64
}

func DefaultStableConfig() StableConfig {
	return StableConfig{
		Base:              DefaultConfig(),
		ScaleFractions:    []float64{0.65, 0.85, 1.0},
		MatchIoU:          0.55,
		MinSupport:        2,
		BoundaryTolerance: 0.40,
	}
}

type stableCandidate struct {
	region     schema.Region
	scaleID    int
	scale      float64
	viewWidth  int
	viewHeight int
}

type stableCluster struct {
	members []stableCandidate
}

// ProposeStable obtains proposals at several bounded resolutions, clusters
// geometrically consistent candidates, and emits only candidates supported by
// multiple scales. Coordinates and areas are returned in the finest view's
// coordinate system. Boundary disagreement prevents weak cross-object merges.
func ProposeStable(img image.Image, maxSide int, cfg StableConfig) ([]schema.Region, error) {
	if img == nil {
		return nil, fmt.Errorf("stable proposal: image is nil")
	}
	if maxSide <= 0 {
		maxSide = 1024
	}
	if cfg.Base.MergeThreshold <= 0 || cfg.Base.MinArea < 1 || cfg.Base.MaxRegions < 1 {
		base := DefaultStableConfig()
		if cfg.Base.MergeThreshold <= 0 {
			cfg.Base.MergeThreshold = base.Base.MergeThreshold
		}
		if cfg.Base.MinArea < 1 {
			cfg.Base.MinArea = base.Base.MinArea
		}
		if cfg.Base.MaxRegions < 1 {
			cfg.Base.MaxRegions = base.Base.MaxRegions
		}
	}
	if len(cfg.ScaleFractions) == 0 {
		cfg.ScaleFractions = DefaultStableConfig().ScaleFractions
	}
	if cfg.ExtraScaleFraction > 1 {
		cfg.ScaleFractions = append(append([]float64(nil), cfg.ScaleFractions...), cfg.ExtraScaleFraction)
	}
	if cfg.MatchIoU <= 0 {
		cfg.MatchIoU = DefaultStableConfig().MatchIoU
	}
	if cfg.MinSupport < 1 {
		cfg.MinSupport = DefaultStableConfig().MinSupport
	}
	if cfg.BoundaryTolerance <= 0 {
		cfg.BoundaryTolerance = DefaultStableConfig().BoundaryTolerance
	}

	baseView, err := imageview.FromImage(img, maxSide)
	if err != nil {
		return nil, err
	}
	candidates := make([]stableCandidate, 0, len(cfg.ScaleFractions)*cfg.Base.MaxRegions)
	seenViews := make(map[[2]int]bool)
	for scaleID, fraction := range cfg.ScaleFractions {
		if fraction <= 0 {
			continue
		}
		side := max(1, int(float64(maxSide)*fraction+0.5))
		view, viewErr := imageview.FromImage(img, side)
		if viewErr != nil {
			return nil, viewErr
		}
		key := [2]int{view.Width, view.Height}
		if seenViews[key] {
			continue
		}
		seenViews[key] = true
		field := evidence.Compute(view)
		for _, proposal := range Propose(field, cfg.Base) {
			candidates = append(candidates, stableCandidate{
				region: proposal, scaleID: scaleID, scale: float64(view.Width) / float64(baseView.Width),
				viewWidth: view.Width, viewHeight: view.Height,
			})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Large candidates seed clusters first, making the matching deterministic
	// while avoiding a small fragmented region becoming the cluster anchor.
	sort.SliceStable(candidates, func(i, j int) bool {
		ai := normalizedArea(candidates[i].region.BBox, candidates[i].viewWidth, candidates[i].viewHeight)
		aj := normalizedArea(candidates[j].region.BBox, candidates[j].viewWidth, candidates[j].viewHeight)
		if ai == aj {
			return candidates[i].scaleID < candidates[j].scaleID
		}
		return ai > aj
	})
	clusters := make([]stableCluster, 0, len(candidates))
	for _, candidate := range candidates {
		bestCluster, bestScore := -1, 0.0
		for i := range clusters {
			anchor := clusters[i].members[0]
			overlap := normalizedIoU(candidate.region.BBox, candidate.viewWidth, candidate.viewHeight, anchor.region.BBox, anchor.viewWidth, anchor.viewHeight)
			boundaryGap := abs(candidate.region.Features.BoundaryStrength - anchor.region.Features.BoundaryStrength)
			if overlap < cfg.MatchIoU || (boundaryGap > cfg.BoundaryTolerance && overlap < 0.80) {
				continue
			}
			score := overlap - 0.10*boundaryGap
			if score > bestScore {
				bestCluster, bestScore = i, score
			}
		}
		if bestCluster < 0 {
			clusters = append(clusters, stableCluster{members: []stableCandidate{candidate}})
		} else {
			clusters[bestCluster].members = append(clusters[bestCluster].members, candidate)
		}
	}

	type output struct {
		region  schema.Region
		support int
		area    float64
	}
	outputs := make([]output, 0, len(clusters))
	for _, cluster := range clusters {
		scaleSet := make(map[int]bool)
		for _, member := range cluster.members {
			scaleSet[member.scaleID] = true
		}
		support := len(scaleSet)
		if support < cfg.MinSupport && len(seenViews) > 1 {
			continue
		}
		representative := cluster.members[0]
		for _, member := range cluster.members[1:] {
			if member.region.Features.BoundaryStrength > representative.region.Features.BoundaryStrength {
				representative = member
			}
		}
		norm := normalizedRect(representative.region.BBox, representative.viewWidth, representative.viewHeight)
		bbox := denormalizedRect(norm, baseView.Width, baseView.Height)
		areaRatio := normalizedArea(representative.region.BBox, representative.viewWidth, representative.viewHeight)
		area := max(1, int(areaRatio*float64(baseView.Width*baseView.Height)+0.5))
		stability := float64(support) / float64(len(seenViews))
		if representative.region.Features.Extra == nil {
			representative.region.Features.Extra = make(map[string]float64)
		}
		representative.region.BBox = bbox
		representative.region.Area = area
		representative.region.Features.AreaRatio = areaRatio
		representative.region.Features.ScaleStability = stability
		representative.region.Features.Extra["scale_support"] = float64(support)
		representative.region.Features.Extra["scale_count"] = float64(len(seenViews))
		representative.region.Evidence = append(representative.region.Evidence,
			schema.EvidenceRef{Kind: "scale-consensus", Stage: "multi-scale-merge", Value: stability, Note: fmt.Sprintf("support=%d/%d", support, len(seenViews))})
		for _, member := range cluster.members {
			representative.region.Evidence = append(representative.region.Evidence, schema.EvidenceRef{
				Kind: "scale-region", Stage: "multi-scale-merge", Scale: member.scale, Value: member.region.Features.BoundaryStrength,
				Note: fmt.Sprintf("candidate=%s", member.region.ID),
			})
		}
		outputs = append(outputs, output{region: representative.region, support: support, area: areaRatio})
	}
	sort.Slice(outputs, func(i, j int) bool {
		if outputs[i].area == outputs[j].area {
			if outputs[i].region.BBox.Y == outputs[j].region.BBox.Y {
				return outputs[i].region.BBox.X < outputs[j].region.BBox.X
			}
			return outputs[i].region.BBox.Y < outputs[j].region.BBox.Y
		}
		return outputs[i].area > outputs[j].area
	})
	if len(outputs) > cfg.Base.MaxRegions {
		outputs = outputs[:cfg.Base.MaxRegions]
	}
	regions := make([]schema.Region, 0, len(outputs))
	for i, item := range outputs {
		item.region.ID = fmt.Sprintf("r-%04d", i+1)
		for j := range item.region.Evidence {
			if item.region.Evidence[j].RegionID != "" {
				item.region.Evidence[j].RegionID = item.region.ID
			}
		}
		regions = append(regions, item.region)
	}
	return regions, nil
}

func normalizedRect(r schema.Rect, width, height int) [4]float64 {
	return [4]float64{float64(r.X) / float64(max(1, width)), float64(r.Y) / float64(max(1, height)), float64(r.W) / float64(max(1, width)), float64(r.H) / float64(max(1, height))}
}

func normalizedArea(r schema.Rect, width, height int) float64 {
	return float64(max(0, r.W)*max(0, r.H)) / float64(max(1, width*height))
}

func normalizedIoU(a schema.Rect, aw, ah int, b schema.Rect, bw, bh int) float64 {
	return rectIoU(normalizedRect(a, aw, ah), normalizedRect(b, bw, bh))
}

func rectIoU(a, b [4]float64) float64 {
	ax2, ay2 := a[0]+a[2], a[1]+a[3]
	bx2, by2 := b[0]+b[2], b[1]+b[3]
	ix1, iy1 := math.Max(a[0], b[0]), math.Max(a[1], b[1])
	ix2, iy2 := math.Min(ax2, bx2), math.Min(ay2, by2)
	if ix2 <= ix1 || iy2 <= iy1 {
		return 0
	}
	inter := (ix2 - ix1) * (iy2 - iy1)
	union := a[2]*a[3] + b[2]*b[3] - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

func denormalizedRect(r [4]float64, width, height int) schema.Rect {
	x := int(math.Floor(r[0]*float64(width) + 0.5))
	y := int(math.Floor(r[1]*float64(height) + 0.5))
	w := max(1, int(math.Floor(r[2]*float64(width)+0.5)))
	h := max(1, int(math.Floor(r[3]*float64(height)+0.5)))
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > width {
		w = max(1, width-x)
	}
	if y+h > height {
		h = max(1, height-y)
	}
	return schema.Rect{X: x, Y: y, W: w, H: h}
}
