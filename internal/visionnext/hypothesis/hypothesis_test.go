package hypothesis

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestGenerateKeepsGenericRegionHypothesis(t *testing.T) {
	regions := []schema.Region{{
		ID: "r-0001", Area: 20,
		Features: schema.Features{AreaRatio: 0.1, AspectRatio: 5, ScaleStability: 1},
	}}
	got := Generate(regions, 20, 10)
	if len(got) != 2 {
		t.Fatalf("expected generic and shape hypotheses, got %d", len(got))
	}
	for _, h := range got {
		if h.Label == "person" || h.Label == "object" {
			t.Fatalf("semantic hallucination in early hypothesis stage: %q", h.Label)
		}
	}
}
