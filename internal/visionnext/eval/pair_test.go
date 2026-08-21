package eval

import (
	"path/filepath"
	"testing"

	"heravision/internal/config"
)

func TestEvaluatePairOnRepositoryFixture(t *testing.T) {
	result := EvaluatePair(Sample{
		ID:        "fixture-ui",
		ImagePath: filepath.Join("..", "..", "..", "testdata", "ui.png"),
	}, RunOptions{Mode: "ui", MaxSide: 256, LegacyConfig: config.Default()})
	if result.Status != "ok" {
		t.Fatalf("pair evaluation failed: %+v", result)
	}
	if result.B0.Engine != "B0" || result.B1.Engine != "B1" {
		t.Fatalf("unexpected engines: B0=%s B1=%s", result.B0.Engine, result.B1.Engine)
	}
	if result.Geometry.Threshold != 0.5 {
		t.Fatalf("unexpected matching threshold: %v", result.Geometry.Threshold)
	}
}
