package learned

import (
	"encoding/json"
	"os"
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestLoadAndApplyRegionFilter(t *testing.T) {
	payload := map[string]any{
		"weights": make([]float64, 13),
		"bias":    1.0,
		"mean":    make([]float64, 13),
		"std":     make([]float64, 13),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/filter.json"
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	model, err := LoadRegionFilter(path)
	if err != nil {
		t.Fatal(err)
	}
	regions := model.Apply([]schema.Region{{ID: "r1", Area: 10, BBox: schema.Rect{W: 4, H: 4}}}, 10, 10, 0.5)
	if len(regions) != 1 {
		t.Fatalf("expected region to pass positive bias, got %d", len(regions))
	}
	if len(regions[0].Evidence) != 1 || regions[0].Evidence[0].Kind != "learned-region-filter" {
		t.Fatalf("missing filter provenance: %+v", regions[0].Evidence)
	}
}
