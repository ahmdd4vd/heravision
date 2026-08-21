package region

import (
	"image"
	"image/color"
	"testing"

	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/imageview"
)

func TestProposeSeparatesStrongColorBoundary(t *testing.T) {
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
	regions := Propose(evidence.Compute(v), Config{MergeThreshold: 0.2, MinArea: 2})
	if len(regions) != 2 {
		t.Fatalf("expected two regions, got %d: %+v", len(regions), regions)
	}
	if regions[0].Area != 16 || regions[1].Area != 16 {
		t.Fatalf("unexpected areas: %d and %d", regions[0].Area, regions[1].Area)
	}
}

func TestProposeMergesFlatImage(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 6, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 6; x++ {
			img.SetGray(x, y, color.Gray{Y: 120})
		}
	}
	v, err := imageview.FromImage(img, 0)
	if err != nil {
		t.Fatal(err)
	}
	regions := Propose(evidence.Compute(v), DefaultConfig())
	if len(regions) != 1 {
		t.Fatalf("expected one flat region, got %d", len(regions))
	}
	if regions[0].Area != 30 {
		t.Fatalf("unexpected flat region area: %d", regions[0].Area)
	}
}
