package semantic

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"heravision/internal/visionnext/imageview"
	"heravision/internal/visionnext/schema"
)

type Model struct {
	Name         string               `json:"name"`
	FeatureNames []string             `json:"features"`
	Labels       []string             `json:"labels"`
	Weights      map[string][]float64 `json:"weights"`
	Bias         map[string]float64   `json:"bias"`
	MinEvidence  float64              `json:"min_evidence"`
	MinMargin    float64              `json:"min_margin"`
	MinSupport   int                  `json:"min_support,omitempty"`
}

func Load(path string) (Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return Model{}, err
	}
	if len(model.Labels) == 0 || len(model.FeatureNames) == 0 {
		return Model{}, fmt.Errorf("semantic model has no labels or features")
	}
	if model.MinEvidence <= 0 {
		model.MinEvidence = 0.65
	}
	if model.MinMargin <= 0 {
		model.MinMargin = 0.10
	}
	return model, nil
}

func Features(r schema.Region, imageWidth, imageHeight int) []float64 {
	return FeaturesNamed(r, imageWidth, imageHeight, []string{"area_ratio", "aspect_ratio", "compactness", "solidity", "boundary_strength", "scale_stability", "color0", "color1", "color2", "texture0", "x", "y", "w", "h"})
}

func FeaturesNamed(r schema.Region, imageWidth, imageHeight int, names []string) []float64 {
	return featuresNamedWithValues(r, imageWidth, imageHeight, names, nil)
}

// FeaturesNamedWithView adds bounded crop-pixel evidence to geometry features.
// The fixed 8x8 sampling grid bounds CPU cost per region.
func FeaturesNamedWithView(r schema.Region, imageWidth, imageHeight int, view imageview.View, names []string) []float64 {
	return featuresNamedWithValues(r, imageWidth, imageHeight, names, cropValues(r, view))
}

func featuresNamedWithValues(r schema.Region, imageWidth, imageHeight int, names []string, crop map[string]float64) []float64 {
	w, h := float64(max(1, imageWidth)), float64(max(1, imageHeight))
	values := map[string]float64{
		"area_ratio": clamp01(r.Features.AreaRatio), "aspect_ratio": clamp01(r.Features.AspectRatio / 8),
		"compactness": clamp01(r.Features.Compactness), "solidity": clamp01(r.Features.Solidity),
		"boundary_strength": clamp01(r.Features.BoundaryStrength), "scale_stability": clamp01(r.Features.ScaleStability),
		"color0": clamp01(valueAt(r.Features.Color, 0)), "color1": clampSigned(r.Features.Color, 1), "color2": clampSigned(r.Features.Color, 2), "texture0": clamp01(first(r.Features.Texture)),
		"crop_luma_mean": cropValue(crop, "crop_luma_mean"), "crop_luma_std": cropValue(crop, "crop_luma_std"), "crop_chroma_r_mean": cropValue(crop, "crop_chroma_r_mean"), "crop_chroma_b_mean": cropValue(crop, "crop_chroma_b_mean"), "crop_chroma_mag": cropValue(crop, "crop_chroma_mag"), "crop_edge_density": cropValue(crop, "crop_edge_density"), "crop_dark_fraction": cropValue(crop, "crop_dark_fraction"), "crop_bright_fraction": cropValue(crop, "crop_bright_fraction"),
		"ring_luma_mean": cropValue(crop, "ring_luma_mean"), "ring_luma_std": cropValue(crop, "ring_luma_std"), "ring_chroma_mag": cropValue(crop, "ring_chroma_mag"), "object_ring_luma_delta": cropValue(crop, "object_ring_luma_delta"), "object_ring_chroma_delta": cropValue(crop, "object_ring_chroma_delta"),
		"x": clamp01(float64(r.BBox.X) / w), "y": clamp01(float64(r.BBox.Y) / h),
		"w": clamp01(float64(r.BBox.W) / w), "h": clamp01(float64(r.BBox.H) / h),
	}
	out := make([]float64, len(names))
	for i, name := range names {
		out[i] = values[name]
	}
	return out
}

func Infer(regions []schema.Region, imageWidth, imageHeight int, model Model) []schema.Hypothesis {
	return inferWithFeatures(regions, imageWidth, imageHeight, model, func(r schema.Region) []float64 { return FeaturesNamed(r, imageWidth, imageHeight, model.FeatureNames) })
}

// AggregateImageEvidence combines the strongest independent region hypotheses into
// image-level semantic hypotheses. It is intentionally conservative: a single
// small region cannot become an accepted image claim without support from peers.
func AggregateImageEvidence(regions []schema.Region, regionHypotheses []schema.Hypothesis, model Model) []schema.Hypothesis {
	if len(regions) == 0 || len(regionHypotheses) == 0 {
		return nil
	}
	regionByID := make(map[string]schema.Region, len(regions))
	for _, r := range regions {
		regionByID[r.ID] = r
	}
	type scoredRegion struct {
		id        string
		score     float64
		area      float64
		stability float64
	}
	byLabel := make(map[string][]scoredRegion)
	bestByRegion := make(map[string]scoredRegion)
	bestLabelByRegion := make(map[string]string)
	for _, h := range regionHypotheses {
		if len(h.RegionIDs) != 1 || h.Label == "" {
			continue
		}
		id := h.RegionIDs[0]
		r, ok := regionByID[id]
		if !ok {
			continue
		}
		candidate := scoredRegion{id: id, score: clamp01(h.Score), area: clamp01(r.Features.AreaRatio * 4), stability: clamp01(r.Features.ScaleStability)}
		if previous, exists := bestByRegion[id]; !exists || candidate.score > previous.score {
			bestByRegion[id] = candidate
			bestLabelByRegion[id] = h.Label
		}
	}
	for id, item := range bestByRegion {
		byLabel[bestLabelByRegion[id]] = append(byLabel[bestLabelByRegion[id]], item)
	}
	type aggregate struct {
		label   string
		score   float64
		regions []scoredRegion
	}
	aggregates := make([]aggregate, 0, len(byLabel))
	for label, items := range byLabel {
		sort.Slice(items, func(i, j int) bool { return items[i].score > items[j].score })
		if len(items) > 3 {
			items = items[:3]
		}
		if len(items) == 0 {
			continue
		}
		top := items[0].score
		mean := 0.0
		area := 0.0
		stability := 0.0
		for _, item := range items {
			mean += item.score
			area += item.score * item.area
			stability += item.stability
		}
		mean /= float64(len(items))
		area /= float64(len(items))
		stability /= float64(len(items))
		support := clamp01(float64(len(items)) / 3)
		score := clamp01(0.45*top + 0.25*mean + 0.15*area + 0.10*stability + 0.05*support)
		aggregates = append(aggregates, aggregate{label: label, score: score, regions: items})
	}
	sort.Slice(aggregates, func(i, j int) bool { return aggregates[i].score > aggregates[j].score })
	if len(aggregates) == 0 {
		return nil
	}
	margin := aggregates[0].score
	if len(aggregates) > 1 { margin -= aggregates[1].score }
	minSupport := model.MinSupport; if minSupport <= 0 { minSupport = 2 }
	out := make([]schema.Hypothesis, 0, len(aggregates))
	for rank, candidate := range aggregates {
		regionIDs := make([]string, 0, len(candidate.regions))
		evidence := []schema.EvidenceRef{{Kind: "semantic-image-aggregation", Stage: "semantic-aggregate", Value: round(candidate.score), Note: fmt.Sprintf("label=%s rank=%d support=%d min_support=%d top_margin=%.2f", candidate.label, rank, len(candidate.regions), minSupport, margin)}}
		for _, item := range candidate.regions {
			regionIDs = append(regionIDs, item.id)
			evidence = append(evidence, schema.EvidenceRef{Kind: "semantic-region-support", Stage: "semantic-aggregate", RegionID: item.id, Value: round(item.score), Note: "top-k region evidence"})
		}
		status := "candidate"
		if rank == 0 && candidate.score >= model.MinEvidence && margin >= model.MinMargin && len(candidate.regions) >= minSupport {
			status = "accepted"
		} else if rank == 0 && (candidate.score < model.MinEvidence || margin < model.MinMargin || len(candidate.regions) < minSupport) {
			status = "unknown"
		}
		out = append(out, schema.Hypothesis{ID: fmt.Sprintf("image-sem-%s", safeLabel(candidate.label)), RegionIDs: regionIDs, Label: candidate.label, Score: round(candidate.score), Uncertainty: round(1 - candidate.score), Status: status, Evidence: evidence})
	}
	return out
}

func InferWithView(regions []schema.Region, imageWidth, imageHeight int, view imageview.View, model Model) []schema.Hypothesis {
	return inferWithFeatures(regions, imageWidth, imageHeight, model, func(r schema.Region) []float64 {
		return FeaturesNamedWithView(r, imageWidth, imageHeight, view, model.FeatureNames)
	})
}

func inferWithFeatures(regions []schema.Region, imageWidth, imageHeight int, model Model, featureFn func(schema.Region) []float64) []schema.Hypothesis {
	if len(regions) == 0 {
		return nil
	}
	var out []schema.Hypothesis
	for _, r := range regions {
		features := featureFn(r)
		type scored struct {
			label string
			score float64
		}
		scores := make([]scored, 0, len(model.Labels))
		for _, label := range model.Labels {
			weights := model.Weights[label]
			if len(weights) != len(features) {
				continue
			}
			value := model.Bias[label]
			for i := range features {
				value += weights[i] * features[i]
			}
			scores = append(scores, scored{label: label, score: sigmoid(value)})
		}
		sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
		if len(scores) == 0 {
			continue
		}
		margin := scores[0].score
		if len(scores) > 1 {
			margin -= scores[1].score
		}
		quality := evidenceQuality(r)
		for rank, candidate := range scores {
			evidenceScore := candidate.score * quality
			status := "candidate"
			if rank == 0 && evidenceScore >= model.MinEvidence && margin >= model.MinMargin {
				status = "accepted"
			} else if rank == 0 && evidenceScore < model.MinEvidence {
				status = "unknown"
			}
			out = append(out, schema.Hypothesis{
				ID: fmt.Sprintf("%s-sem-%s", r.ID, safeLabel(candidate.label)), RegionIDs: []string{r.ID},
				Label: candidate.label, Score: round(evidenceScore), Uncertainty: round(1 - evidenceScore), Status: status,
				Evidence: []schema.EvidenceRef{
					{Kind: "semantic-logistic", Stage: "semantic", RegionID: r.ID, Value: round(candidate.score), Note: fmt.Sprintf("label=%s rank=%d", candidate.label, rank)},
					{Kind: "visual-evidence-quality", Stage: "semantic", RegionID: r.ID, Value: round(quality)},
				},
			})
		}
	}
	return out
}

func cropValue(values map[string]float64, key string) float64 {
	if values == nil {
		return 0
	}
	return values[key]
}

func cropValues(r schema.Region, view imageview.View) map[string]float64 {
	values := cropValuesBase(r, view)
	for key, value := range ringValues(r, view) {
		values[key] = value
	}
	return values
}

func cropValuesBase(r schema.Region, view imageview.View) map[string]float64 {
	out := map[string]float64{}
	if view.Width <= 0 || view.Height <= 0 || r.BBox.W <= 0 || r.BBox.H <= 0 {
		return out
	}
	const grid = 8
	var sum, sum2, cr, cb, cm, edge, dark, bright, count float64
	previous := make([]float64, grid)
	seenRow := make([]bool, grid)
	for gy := 0; gy < grid; gy++ {
		for gx := 0; gx < grid; gx++ {
			x := r.BBox.X + (gx*r.BBox.W+r.BBox.W/2)/grid
			y := r.BBox.Y + (gy*r.BBox.H+r.BBox.H/2)/grid
			l, vr, vb, ok := view.At(x, y)
			if !ok {
				continue
			}
			sum += l
			sum2 += l * l
			cr += vr
			cb += vb
			cm += math.Sqrt(vr*vr + vb*vb)
			if l < 0.25 {
				dark++
			}
			if l > 0.75 {
				bright++
			}
			count++
			if gx > 0 {
				edge += math.Abs(l - previous[gx-1])
			}
			if seenRow[gx] {
				edge += math.Abs(l - previous[gx])
			}
			previous[gx] = l
			seenRow[gx] = true
		}
	}
	if count == 0 {
		return out
	}
	mean := sum / count
	variance := sum2/count - mean*mean
	if variance < 0 {
		variance = 0
	}
	out["crop_luma_mean"] = mean
	out["crop_luma_std"] = math.Sqrt(variance)
	out["crop_chroma_r_mean"] = clampSignedValue(cr / count)
	out["crop_chroma_b_mean"] = clampSignedValue(cb / count)
	out["crop_chroma_mag"] = clamp01(cm / count)
	out["crop_edge_density"] = clamp01(edge / (2 * count))
	out["crop_dark_fraction"] = dark / count
	out["crop_bright_fraction"] = bright / count
	return out
}

func ringValues(r schema.Region, view imageview.View) map[string]float64 {
	out := map[string]float64{}
	padX := max(2, r.BBox.W/2)
	padY := max(2, r.BBox.H/2)
	outer := schema.Rect{X: r.BBox.X - padX, Y: r.BBox.Y - padY, W: r.BBox.W + 2*padX, H: r.BBox.H + 2*padY}
	var sum, sum2, chroma, count float64
	const grid = 8
	for gy := 0; gy < grid; gy++ {
		for gx := 0; gx < grid; gx++ {
			x := outer.X + (gx*outer.W+outer.W/2)/grid
			y := outer.Y + (gy*outer.H+outer.H/2)/grid
			if x >= r.BBox.X && x < r.BBox.X+r.BBox.W && y >= r.BBox.Y && y < r.BBox.Y+r.BBox.H {
				continue
			}
			l, vr, vb, ok := view.At(x, y)
			if !ok {
				continue
			}
			sum += l
			sum2 += l * l
			chroma += math.Sqrt(vr*vr + vb*vb)
			count++
		}
	}
	if count == 0 {
		return out
	}
	mean := sum / count
	variance := sum2/count - mean*mean
	if variance < 0 {
		variance = 0
	}
	crop := cropValuesBase(r, view)
	cropMean := cropValue(crop, "crop_luma_mean")
	cropChroma := cropValue(crop, "crop_chroma_mag")
	out["ring_luma_mean"] = mean
	out["ring_luma_std"] = math.Sqrt(variance)
	out["ring_chroma_mag"] = clamp01(chroma / count)
	out["object_ring_luma_delta"] = clampSignedValue((cropMean - mean) * 4)
	out["object_ring_chroma_delta"] = clampSignedValue((cropChroma - clamp01(chroma/count)) * 4)
	return out
}

func evidenceQuality(r schema.Region) float64 {
	return clamp01(0.45*clamp01(r.Features.BoundaryStrength) + 0.40*clamp01(r.Features.ScaleStability) + 0.15*clamp01(r.Features.AreaRatio*4))
}

func sigmoid(v float64) float64 {
	if v >= 0 {
		z := math.Exp(-v)
		return 1 / (1 + z)
	}
	z := math.Exp(v)
	return z / (1 + z)
}
func valueAt(values []float64, index int) float64 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}
func clampSigned(values []float64, index int) float64 {
	return clampSignedValue(valueAt(values, index))
}
func clampSignedValue(v float64) float64 { return clamp01((v + 4) / 8) }
func first(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}
func safeLabel(label string) string {
	out := ""
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			out += string(r)
		}
	}
	return out
}
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
func round(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
