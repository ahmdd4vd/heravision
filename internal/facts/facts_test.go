package facts

import (
	"strings"
	"testing"

	"heravision/internal/config"
)

func TestExtractOnFixture(t *testing.T) {
	cfg := config.Default()
	r, err := Extract("../../testdata/ui.png", "ui", "test", cfg)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if r.Meta.Width <= 0 || r.Meta.Height <= 0 {
		t.Fatalf("bad meta: %+v", r.Meta)
	}
	if len(r.Boxes) == 0 {
		t.Fatal("expected at least one box")
	}
	for _, bx := range r.Boxes {
		if bx.Color != "" && len(bx.Color) != 7 {
			t.Fatalf("box color must be #RRGGBB got %q", bx.Color)
		}
	}
	if !strings.Contains(r.Markdown, "Image Facts") {
		t.Fatal("markdown must contain Image Facts header")
	}
	if !strings.Contains(r.Markdown, "NOT OCR-read") {
		t.Fatal("markdown must disclose no-OCR honestly")
	}
}

func TestCompareDiff(t *testing.T) {
	cfg := config.Default()
	res, err := Compare("../../testdata/ui.png", "../../testdata/ui.png", cfg)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if len(res.Diff.Added) != 0 || len(res.Diff.Removed) != 0 || len(res.Diff.Moved) != 0 {
		t.Fatalf("identical images must produce empty diff: %+v", res.Diff)
	}
}

func TestParseHex(t *testing.T) {
	r, g, b := parseHex("#3B82F6")
	if r != 0x3B || g != 0x82 || b != 0xF6 {
		t.Fatalf("parseHex failed: %d %d %d", r, g, b)
	}
	r2, g2, b2 := parseHex("bad")
	if r2 != 0 || g2 != 0 || b2 != 0 {
		t.Fatal("invalid hex must return zeros")
	}
}

func TestColorDistance(t *testing.T) {
	if d := colorDistance("#FF0000", "#FF0000"); d != 0 {
		t.Fatalf("same color distance must be 0 got %.2f", d)
	}
	if d := colorDistance("#000000", "#FFFFFF"); d < 430 {
		t.Fatalf("black-white distance too small: %.2f", d)
	}
}
