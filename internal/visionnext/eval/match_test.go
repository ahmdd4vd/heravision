package eval

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestMatchRegions(t *testing.T) {
	a := []schema.Region{{ID: "a1", BBox: schema.Rect{X: 0, Y: 0, W: 10, H: 10}}}
	b := []schema.Region{
		{ID: "b1", BBox: schema.Rect{X: 1, Y: 1, W: 10, H: 10}},
		{ID: "b2", BBox: schema.Rect{X: 30, Y: 30, W: 5, H: 5}},
	}
	got := MatchRegions(a, b, 0.5)
	if len(got.Matches) != 1 || got.NovelB != 1 {
		t.Fatalf("unexpected match summary: %+v", got)
	}
	if got.CoverageA != 1 || got.MeanIoU <= 0.5 {
		t.Fatalf("unexpected quality metrics: %+v", got)
	}
}

func TestIoUNoOverlap(t *testing.T) {
	if got := IoU(schema.Rect{W: 2, H: 2}, schema.Rect{X: 3, Y: 3, W: 2, H: 2}); got != 0 {
		t.Fatalf("expected zero IoU, got %v", got)
	}
}
