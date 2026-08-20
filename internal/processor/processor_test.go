package processor

import (
	"image"
	"image/color"
	"testing"
)

func TestDecodeAndResize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2000, 1000))
	for y := 0; y < 1000; y++ {
		for x := 0; x < 2000; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	resized := Resize(img, 1024)
	if resized.Bounds().Dx() != 1024 {
		t.Fatalf("expected 1024 width got %d", resized.Bounds().Dx())
	}
}

func TestDecodeMissing(t *testing.T) {
	if _, _, err := Decode("nope.png"); err == nil {
		t.Fatal("expected error")
	}
}
