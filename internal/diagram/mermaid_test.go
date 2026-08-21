package diagram

import (
	"strings"
	"testing"

	"heravision/internal/detector"
)

func TestToMermaidBasic(t *testing.T) {
	boxes := []detector.Box{
		{Type: "card", X: 0, Y: 0, W: 100, H: 50},
		{Type: "button", X: 10, Y: 30, W: 60, H: 20},
	}
	out := ToMermaid(boxes)
	if !strings.HasPrefix(out, "flowchart TD") {
		t.Fatalf("expected flowchart TD prefix got %q", out)
	}
	if !strings.Contains(out, "N0[") || !strings.Contains(out, "N1([") {
		t.Fatalf("expected node shapes got %q", out)
	}
	if !strings.Contains(out, "N0 --> N1") {
		t.Fatalf("expected edge got %q", out)
	}
}

func TestToMermaidEmpty(t *testing.T) {
	out := ToMermaid(nil)
	if !strings.Contains(out, "empty") {
		t.Fatalf("expected empty fallback got %q", out)
	}
}
