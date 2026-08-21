package relation

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestBuildVisibleGeometryRelations(t *testing.T) {
	regions := []schema.Region{
		{ID: "top", BBox: schema.Rect{X: 2, Y: 1, W: 6, H: 3}},
		{ID: "bottom", BBox: schema.Rect{X: 2, Y: 8, W: 6, H: 3}},
		{ID: "side", BBox: schema.Rect{X: 12, Y: 8, W: 3, H: 3}},
	}
	edges := Build(regions)
	seen := map[string]bool{}
	for _, e := range edges {
		if e.Status != "visible" {
			t.Fatalf("unexpected non-visible relation: %+v", e)
		}
		seen[e.From+":"+e.Predicate+":"+e.To] = true
	}
	if !seen["top:above:bottom"] {
		t.Fatalf("missing above relation: %+v", edges)
	}
	if !seen["bottom:left_of:side"] {
		t.Fatalf("missing left_of relation: %+v", edges)
	}
}

func TestBuildDoesNotInventHoldingRelation(t *testing.T) {
	regions := []schema.Region{
		{ID: "a", BBox: schema.Rect{X: 0, Y: 0, W: 5, H: 5}},
		{ID: "b", BBox: schema.Rect{X: 5, Y: 5, W: 5, H: 5}},
	}
	for _, e := range Build(regions) {
		if e.Predicate == "holding" || e.Predicate == "using" {
			t.Fatalf("semantic relation invented by geometry engine: %+v", e)
		}
	}
}
