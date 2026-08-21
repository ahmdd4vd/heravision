package eval

import (
	"path/filepath"
	"testing"

	"heravision/internal/config"
)

func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "testdata", "ui.png")
}

func TestRunB1ProducesNormalizedObservation(t *testing.T) {
	obs, err := RunB1(fixturePath(t), RunOptions{Mode: "pure", MaxSide: 128})
	if err != nil {
		t.Fatal(err)
	}
	if obs.Engine != "B1" || len(obs.Graph.Nodes) == 0 {
		t.Fatalf("unexpected B1 observation: %+v", obs)
	}
	if err := obs.Graph.Validate(); err != nil {
		t.Fatalf("B1 graph invalid: %v", err)
	}
}

func TestRunB0ProducesNormalizedObservation(t *testing.T) {
	obs, err := RunB0(fixturePath(t), RunOptions{Mode: "ui", EngineVersion: "test-b0", LegacyConfig: config.Default()})
	if err != nil {
		t.Fatal(err)
	}
	if obs.Engine != "B0" {
		t.Fatalf("unexpected engine: %s", obs.Engine)
	}
	if obs.Width <= 0 || obs.Height <= 0 {
		t.Fatalf("invalid dimensions: %dx%d", obs.Width, obs.Height)
	}
	if err := obs.Graph.Validate(); err != nil {
		t.Fatalf("B0 graph invalid: %v", err)
	}
}
