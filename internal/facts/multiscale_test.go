package facts

import (
	"image"
	"image/color"
	"testing"

	"heravision/internal/detector"
	"heravision/internal/processor"
)

func TestMultiScaleFindsSmallObject(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1600, 900))
	for y := 0; y < 900; y++ {
		for x := 0; x < 1600; x++ {
			src.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 600; y < 616; y++ {
		for x := 1200; x < 1216; x++ {
			src.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	base := resizeForTest(src, 1024)
	params := detector.DefaultParams
	boxesOnBase := detector.DetectCfg(base, params)
	smallOnBase := false
	for _, b := range boxesOnBase {
		if b.X >= 700 && b.X <= 840 && b.Y >= 340 && b.Y <= 430 {
			smallOnBase = true
		}
	}
	if smallOnBase {
		t.Skip("base scale already detects it; multi-scale merge trivially passes")
	}
	merged := detectMultiScale(src, base, params, 1024, true)
	found := false
	for _, b := range merged {
		if b.X >= 700 && b.X <= 840 && b.Y >= 340 && b.Y <= 430 && b.W <= 40 && b.H <= 40 {
			found = true
		}
	}
	if !found {
		t.Fatalf("multi-scale must find 16px object near (768,384) in base coords, got %+v", merged)
	}
}

func TestMultiScaleDisabledEqualsBase(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 800, 500))
	for y := 0; y < 500; y++ {
		for x := 0; x < 800; x++ {
			src.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 100; y < 180; y++ {
		for x := 200; x < 500; x++ {
			src.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	base := resizeForTest(src, 1024)
	params := detector.DefaultParams
	a := detectMultiScale(src, base, params, 1024, false)
	b := detector.DetectCfg(base, params)
	if len(a) != len(b) {
		t.Fatalf("disabled multiscale must equal base detection: %d vs %d", len(a), len(b))
	}
}

func TestMergeScalesDedup(t *testing.T) {
	boxes := []detector.Box{
		{Type: "card", X: 10, Y: 10, W: 100, H: 50},
		{Type: "card", X: 12, Y: 11, W: 98, H: 49},
		{Type: "button", X: 300, Y: 300, W: 60, H: 24},
	}
	merged := mergeScales(boxes)
	if len(merged) != 2 {
		t.Fatalf("expected 2 boxes after dedup got %d: %+v", len(merged), merged)
	}
}

func resizeForTest(src image.Image, maxSide int) image.Image {
	return processor.Resize(src, maxSide)
}
