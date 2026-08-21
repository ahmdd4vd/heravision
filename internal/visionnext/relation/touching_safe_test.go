package relation

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestBuildPrunedSafeTouchingSuppressesContainmentTouching(t *testing.T) {
	regions := []schema.Region{
		{ID: "outer", BBox: schema.Rect{X: 0, Y: 0, W: 100, H: 100}},
		{ID: "inner", BBox: schema.Rect{X: 20, Y: 20, W: 20, H: 20}},
	}
	edges := BuildPruned(regions, PruneConfig{MaxGapPixels: 48, SafeTouching: true})
	for _, edge := range edges {
		if edge.Predicate == "touching" {
			t.Fatalf("containment must not imply touching: %+v", edge)
		}
	}
}

func TestBuildPrunedSafeTouchingKeepsAdjacentBoundaryContact(t *testing.T) {
	regions := []schema.Region{
		{ID: "left", BBox: schema.Rect{X: 0, Y: 0, W: 20, H: 20}},
		{ID: "right", BBox: schema.Rect{X: 20, Y: 0, W: 20, H: 20}},
	}
	edges := BuildPruned(regions, PruneConfig{MaxGapPixels: 48, SafeTouching: true})
	found := false
	for _, edge := range edges {
		if edge.Predicate == "touching" {
			found = true
			if len(edge.Evidence) != 1 || edge.Evidence[0].Kind != "boundary-contact" {
				t.Fatalf("missing boundary-contact evidence: %+v", edge)
			}
		}
	}
	if !found {
		t.Fatal("expected boundary contact relation")
	}
}
