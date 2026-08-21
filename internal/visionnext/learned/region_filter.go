package learned

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"heravision/internal/visionnext/schema"
)

type RegionFilter struct {
	Weights []float64 `json:"weights"`
	Bias    float64   `json:"bias"`
	Mean    []float64 `json:"mean"`
	Std     []float64 `json:"std"`
}

func LoadRegionFilter(path string) (RegionFilter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RegionFilter{}, fmt.Errorf("read region filter: %w", err)
	}
	var model RegionFilter
	if err := json.Unmarshal(data, &model); err != nil {
		return RegionFilter{}, fmt.Errorf("parse region filter: %w", err)
	}
	if len(model.Weights) != 13 || len(model.Mean) != 13 || len(model.Std) != 13 {
		return RegionFilter{}, fmt.Errorf("region filter expects 13 features, got weights=%d mean=%d std=%d", len(model.Weights), len(model.Mean), len(model.Std))
	}
	for i := range model.Std {
		if model.Std[i] == 0 {
			model.Std[i] = 1
		}
	}
	return model, nil
}

func (m RegionFilter) Score(r schema.Region, width, height int) float64 {
	f := features(r, width, height)
	z := m.Bias
	for i, value := range f {
		z += ((value - m.Mean[i]) / m.Std[i]) * m.Weights[i]
	}
	if z < -30 {
		z = -30
	}
	if z > 30 {
		z = 30
	}
	return 1 / (1 + math.Exp(-z))
}

func (m RegionFilter) Apply(regions []schema.Region, width, height int, threshold float64) []schema.Region {
	if threshold <= 0 {
		threshold = 0.95
	}
	out := make([]schema.Region, 0, len(regions))
	for _, r := range regions {
		score := m.Score(r, width, height)
		if score < threshold {
			continue
		}
		r.Evidence = append(r.Evidence, schema.EvidenceRef{Kind: "learned-region-filter", Stage: "b1-filter", Value: score, RegionID: r.ID})
		out = append(out, r)
	}
	return out
}

func features(r schema.Region, width, height int) []float64 {
	texture := 0.0
	if len(r.Features.Texture) > 0 {
		texture = r.Features.Texture[0]
	}
	return []float64{
		math.Log1p(float64(maxInt(0, r.Area))),
		math.Log1p(float64(maxInt(0, r.BBox.W))),
		math.Log1p(float64(maxInt(0, r.BBox.H))),
		r.Features.AspectRatio,
		r.Features.AreaRatio,
		math.Log1p(maxFloat(0, r.Features.Compactness)),
		r.Features.BoundaryStrength,
		r.Features.ScaleStability,
		safeRatio(r.BBox.X, width),
		safeRatio(r.BBox.Y, height),
		safeRatio(r.BBox.W, width),
		safeRatio(r.BBox.H, height),
		texture,
	}
}

func safeRatio(a, b int) float64 {
	if b <= 0 {
		return 0
	}
	return float64(a) / float64(b)
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
