package semantic

import (
	"testing"

	"heravision/internal/visionnext/imageview"
	"heravision/internal/visionnext/schema"
)

func TestInferAcceptsStrongCandidateWithMargin(t *testing.T) {
	model := Model{
		Labels:       []string{"animal", "artifact"},
		FeatureNames: []string{"area_ratio", "aspect_ratio", "compactness", "solidity", "boundary_strength", "scale_stability", "color", "texture", "x", "y", "w", "h"},
		Weights: map[string][]float64{
			"animal":   {0, 0, 0, 0, 3, 3, 0, 0, 0, 0, 0, 0},
			"artifact": {0, 0, 0, 0, -3, -3, 0, 0, 0, 0, 0, 0},
		},
		Bias: map[string]float64{"animal": 1, "artifact": -1}, MinEvidence: 0.65, MinMargin: 0.10,
	}
	regions := []schema.Region{{ID: "r-1", Area: 1000, BBox: schema.Rect{X: 10, Y: 10, W: 80, H: 80}, Features: schema.Features{AreaRatio: 0.2, AspectRatio: 1, BoundaryStrength: 0.9, ScaleStability: 0.9}}}
	out := Infer(regions, 100, 100, model)
	if len(out) == 0 || out[0].Label != "animal" || out[0].Status != "accepted" {
		t.Fatalf("expected accepted animal hypothesis, got %+v", out)
	}
	if len(out[0].Evidence) < 2 {
		t.Fatalf("expected evidence provenance, got %+v", out[0].Evidence)
	}
}

func TestFeaturesNamedWithViewUsesCropEvidence(t *testing.T) {
	view := imageview.View{Width: 8, Height: 8, Luminance: make([]float64, 64), ChromaR: make([]float64, 64), ChromaB: make([]float64, 64)}
	for i := range view.Luminance {
		view.Luminance[i] = float64(i) / 63
		view.ChromaR[i] = 0.4
		view.ChromaB[i] = -0.2
	}
	region := schema.Region{BBox: schema.Rect{X: 0, Y: 0, W: 8, H: 8}}
	features := FeaturesNamedWithView(region, 8, 8, view, []string{"crop_luma_mean", "crop_luma_std", "crop_chroma_r_mean", "crop_chroma_b_mean", "crop_chroma_mag", "crop_edge_density", "crop_dark_fraction", "crop_bright_fraction"})
	if len(features) != 8 || features[0] <= 0 || features[1] <= 0 || features[4] <= 0 || features[5] <= 0 {
		t.Fatalf("expected nontrivial crop features, got %v", features)
	}
}

func TestInferMarksWeakEvidenceUnknown(t *testing.T) {
	model := Model{
		Labels: []string{"animal", "artifact"}, FeatureNames: []string{"boundary_strength", "scale_stability"},
		Weights: map[string][]float64{"animal": {2, 2}, "artifact": {-2, -2}}, Bias: map[string]float64{"animal": 0, "artifact": 0}, MinEvidence: 0.9, MinMargin: 0.1,
	}
	regions := []schema.Region{{ID: "r-1", Area: 20, BBox: schema.Rect{X: 1, Y: 1, W: 5, H: 4}, Features: schema.Features{BoundaryStrength: 0.1, ScaleStability: 0.1}}}
	out := Infer(regions, 100, 100, model)
	if len(out) == 0 || out[0].Status != "unknown" {
		t.Fatalf("expected unknown weak hypothesis, got %+v", out)
	}
}
