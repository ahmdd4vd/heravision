package semantic

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

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
	w, h := float64(max(1, imageWidth)), float64(max(1, imageHeight))
	values := map[string]float64{
		"area_ratio": clamp01(r.Features.AreaRatio), "aspect_ratio": clamp01(r.Features.AspectRatio / 8),
		"compactness": clamp01(r.Features.Compactness), "solidity": clamp01(r.Features.Solidity),
		"boundary_strength": clamp01(r.Features.BoundaryStrength), "scale_stability": clamp01(r.Features.ScaleStability),
		"color0": clamp01(valueAt(r.Features.Color, 0)), "color1": clampSigned(r.Features.Color, 1), "color2": clampSigned(r.Features.Color, 2), "texture0": clamp01(first(r.Features.Texture)),
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
	if len(regions) == 0 {
		return nil
	}
	var out []schema.Hypothesis
	for _, r := range regions {
		features := FeaturesNamed(r, imageWidth, imageHeight, model.FeatureNames)
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
	return clamp01((valueAt(values, index) + 4) / 8)
}
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
