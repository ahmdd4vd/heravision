package ocr

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func solidWithRect(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 60; y < 90; y++ {
		for x := 50; x < 250; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	return img
}

func TestHeuristicPlaceholders(t *testing.T) {
	texts := Extract(solidWithRect(300, 200))
	if len(texts) == 0 {
		t.Fatal("expected at least one placeholder text")
	}
	for _, tx := range texts {
		if !strings.HasPrefix(tx.Text, "[") {
			t.Fatalf("placeholder must start with '[' got %q", tx.Text)
		}
	}
}

func TestEngineInterfaceFallback(t *testing.T) {
	old := engine
	defer func() { engine = old }()
	engine = nil
	texts := Extract(solidWithRect(300, 200))
	if len(texts) == 0 {
		t.Fatal("fallback heuristic must still work when engine is nil")
	}
}
