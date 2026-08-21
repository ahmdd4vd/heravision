package relation

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestBuildPrunedDropsDistantPairs(t *testing.T) {
	regions := []schema.Region{
		{ID: "a", BBox: schema.Rect{X: 0, Y: 0, W: 10, H: 10}},
		{ID: "b", BBox: schema.Rect{X: 100, Y: 0, W: 10, H: 10}},
	}
	if got := len(BuildPruned(regions, PruneConfig{MaxGapPixels: 8})); got != 0 {
		t.Fatalf("expected no relations for distant pair, got %d", got)
	}
}

func TestBuildPrunedKeepsNearbyAndContainment(t *testing.T) {
	regions := []schema.Region{
		{ID: "outer", BBox: schema.Rect{X: 0, Y: 0, W: 20, H: 20}},
		{ID: "inner", BBox: schema.Rect{X: 2, Y: 2, W: 4, H: 4}},
		{ID: "near", BBox: schema.Rect{X: 22, Y: 0, W: 10, H: 10}},
	}
	relations := BuildPruned(regions, PruneConfig{MaxGapPixels: 4})
	seenContains, seenNear := false, false
	for _, r := range relations {
		if r.Predicate == "contains" && r.From == "outer" && r.To == "inner" {
			seenContains = true
		}
		if r.Predicate == "left_of" && r.From == "outer" && r.To == "near" {
			seenNear = true
		}
	}
	if !seenContains {
		t.Fatal("expected containment relation to survive pruning")
	}
	if !seenNear {
		t.Fatal("expected nearby left_of relation to survive pruning")
	}
}
