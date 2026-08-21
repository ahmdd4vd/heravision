package evidence

import (
	"image"
	"image/color"
	"testing"

	"heravision/internal/visionnext/imageview"
)

func TestComputeDetectsBoundaryAndChroma(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			if x < 4 {
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	v, err := imageview.FromImage(img, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := Compute(v)
	if f.Width != 8 || f.Height != 4 {
		t.Fatalf("unexpected field size: %dx%d", f.Width, f.Height)
	}
	if f.ChromaMagnitude[0] == 0 || f.ChromaMagnitude[7] == 0 {
		t.Fatal("expected chroma evidence on colored pixels")
	}
	if f.Edge[1*8+3] <= f.Edge[1*8+1] {
		t.Fatal("expected stronger edge near color boundary")
	}
}

func TestComputeFlatImageHasLowEdge(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.SetGray(x, y, color.Gray{Y: 120})
		}
	}
	v, err := imageview.FromImage(img, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := Compute(v)
	for i, e := range f.Edge {
		if e != 0 {
			t.Fatalf("flat image has edge at %d: %v", i, e)
		}
	}
	if f.Flatness[12] < 0.99 {
		t.Fatalf("flatness too low: %v", f.Flatness[12])
	}
}
