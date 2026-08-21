package region

import (
	"image"
	"image/color"
	"testing"
)

func TestProposeStableFlatImageHasMultiScaleSupport(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			img.SetGray(x, y, color.Gray{Y: 120})
		}
	}
	regions, err := ProposeStable(img, 32, StableConfig{Base: Config{MergeThreshold: 0.2, MinArea: 2, MaxRegions: 256}})
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 1 {
		t.Fatalf("expected one stable flat region, got %d", len(regions))
	}
	if regions[0].Features.ScaleStability < 0.66 {
		t.Fatalf("expected multi-scale support, got %.3f", regions[0].Features.ScaleStability)
	}
	if len(regions[0].Evidence) < 2 {
		t.Fatalf("expected scale provenance evidence, got %d refs", len(regions[0].Evidence))
	}
}

func TestProposeStablePreservesStrongBoundary(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if x < 32 {
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	regions, err := ProposeStable(img, 32, StableConfig{Base: Config{MergeThreshold: 0.2, MinArea: 2, MaxRegions: 256}})
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 2 {
		t.Fatalf("expected two stable regions, got %d: %+v", len(regions), regions)
	}
	for _, r := range regions {
		if r.Features.ScaleStability < 0.66 {
			t.Fatalf("expected stable boundary region, got %.3f", r.Features.ScaleStability)
		}
		if r.BBox.W < 10 || r.BBox.H < 10 {
			t.Fatalf("unexpected stable bbox: %+v", r.BBox)
		}
	}
}
