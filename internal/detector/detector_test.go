package detector

import (
	"image"
	"image/color"
	"testing"
)

func TestDetectSimple(t *testing.T) {
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
	boxes := Detect(img)
	if len(boxes) == 0 {
		t.Fatal("expected at least 1 box")
	}
}
