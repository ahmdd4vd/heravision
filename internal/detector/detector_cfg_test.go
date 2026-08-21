package detector

import (
	"image"
	"image/color"
	"testing"
)

func TestDetectCfgMinAreaFilters(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 300; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 50; y < 90; y++ {
		for x := 80; x < 220; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	p := Params{CannyLow: 50, CannyHigh: 150, MinArea: 1 << 20}
	if boxes := DetectCfg(img, p); len(boxes) != 0 {
		t.Fatalf("huge min_area must filter all boxes got %d", len(boxes))
	}
}

func TestDetectBoxHasColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 300; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 50; y < 90; y++ {
		for x := 80; x < 220; x++ {
			img.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}
	boxes := Detect(img)
	found := false
	for _, b := range boxes {
		if b.Color != "" {
			found = true
			if len(b.Color) != 7 || b.Color[0] != '#' {
				t.Fatalf("color format must be #RRGGBB got %q", b.Color)
			}
		}
	}
	if !found {
		t.Fatal("expected at least one box with avg color")
	}
}
