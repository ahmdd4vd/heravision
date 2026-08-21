package hypothesis

import (
	"fmt"
	"sort"

	"heravision/internal/visionnext/schema"
)

func Generate(regions []schema.Region, imageWidth, imageHeight int) []schema.Hypothesis {
	if imageWidth <= 0 || imageHeight <= 0 {
		return nil
	}
	out := make([]schema.Hypothesis, 0, len(regions)*2)
	for _, r := range regions {
		if r.ID == "" || r.Area <= 0 {
			continue
		}
		areaRatio := float64(r.Area) / float64(imageWidth*imageHeight)
		base := 0.35 + 0.35*clamp01(r.Features.ScaleStability)
		out = append(out, schema.Hypothesis{
			ID:          fmt.Sprintf("%s-h0", r.ID),
			RegionIDs:   []string{r.ID},
			Label:       "region",
			Score:       round(base),
			Uncertainty: round(1 - base),
			Status:      "candidate",
			Evidence: []schema.EvidenceRef{{
				Kind: "region-geometry", Stage: "hypothesis", RegionID: r.ID, Value: areaRatio,
			}},
		})

		label, score := shapeLabel(r)
		if label != "region" {
			out = append(out, schema.Hypothesis{
				ID:          fmt.Sprintf("%s-h1", r.ID),
				RegionIDs:   []string{r.ID},
				Label:       label,
				Score:       round(score),
				Uncertainty: round(1 - score),
				Status:      "candidate",
				Evidence: []schema.EvidenceRef{{
					Kind: "shape-statistics", Stage: "hypothesis", RegionID: r.ID, Value: score,
				}},
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RegionIDs[0] == out[j].RegionIDs[0] {
			return out[i].Score > out[j].Score
		}
		return out[i].RegionIDs[0] < out[j].RegionIDs[0]
	})
	return out
}

func shapeLabel(r schema.Region) (string, float64) {
	ar := r.Features.AspectRatio
	area := r.Features.AreaRatio
	texture := 0.0
	if len(r.Features.Texture) > 0 {
		texture = r.Features.Texture[0]
	}
	switch {
	case ar >= 4 && area < 0.25:
		return "elongated_region", 0.62
	case ar >= 0.75 && ar <= 1.33 && area < 0.35:
		return "compact_region", 0.58
	case texture < 0.03 && area > 0.15:
		return "flat_surface_region", 0.55
	default:
		return "region", 0
	}
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

func round(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
