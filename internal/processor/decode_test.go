package processor

import (
	"image"
	"image/color"
	"testing"
)

func makeSolid(w, h int, r, g, b uint8) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func TestDecodeMaxPixelsLimit(t *testing.T) {
	old := MaxPixels
	defer func() { MaxPixels = old }()
	MaxPixels = 100
	if _, _, err := Decode("../../testdata/ui.png"); err == nil {
		t.Fatal("expected error when image exceeds MaxPixels")
	}
}

func TestDecodeOversizeSide(t *testing.T) {
	old := MaxPixels
	defer func() { MaxPixels = old }()
	MaxPixels = 1 << 40
	if _, _, err := Decode("nope_missing.png"); err == nil {
		t.Fatal("expected open error")
	}
}

func TestDecodeValid(t *testing.T) {
	img, format, err := Decode("../../testdata/ui.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil || format == "" {
		t.Fatal("expected valid image and format")
	}
}

func TestDecodeWithMaxPixelsDoesNotMutateGlobal(t *testing.T) {
	old := MaxPixels
	defer func() { MaxPixels = old }()
	MaxPixels = 1 << 40
	if _, _, err := DecodeWithMaxPixels("../../testdata/ui.png", 100); err == nil {
		t.Fatal("expected per-call pixel limit error")
	}
	if MaxPixels != 1<<40 {
		t.Fatalf("per-call decode mutated global MaxPixels: %d", MaxPixels)
	}
}
