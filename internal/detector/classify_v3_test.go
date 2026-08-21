package detector

import (
	"image"
	"image/color"
	"testing"
)

func splitImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < w/2 {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			} else {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	return img
}

func TestNmsThinVerticalEdge(t *testing.T) {
	gray := toGray(splitImage(100, 60))
	blurred := gaussian3x3(gray)
	mag, dir := sobelGrad(blurred)
	thin := nonMaxSuppress(mag, dir)
	cols := map[int]bool{}
	for y := 10; y < 50; y++ {
		for x := 1; x < 99; x++ {
			if thin[y][x] != 0 {
				cols[x] = true
			}
		}
	}
	if len(cols) == 0 {
		t.Fatal("expected edge pixels on vertical boundary")
	}
	if len(cols) > 3 {
		t.Fatalf("nms must keep edge thin (<=3 cols), got %d cols: %v", len(cols), cols)
	}
}

func TestCannyHysteresisStillWorks(t *testing.T) {
	gray := toGray(splitImage(100, 60))
	blurred := gaussian3x3(gray)
	edges := canny(blurred, 50, 150)
	count := 0
	for y := range edges {
		for x := range edges[y] {
			if edges[y][x] != 0 {
				count++
			}
		}
	}
	if count == 0 {
		t.Fatal("canny must produce edges")
	}
}

func TestClassifyV3Checkbox(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	fillRect(img, 0, 0, 300, 300, color.RGBA{255, 255, 255, 255})
	drawOutline(img, 100, 100, 20, 20, 2, color.RGBA{0, 0, 0, 255})
	boxes := Detect(img)
	found := false
	for _, b := range boxes {
		if b.Type == "checkbox" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected checkbox type, got %+v", boxes)
	}
}

func TestClassifyV3Avatar(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	fillRect(img, 0, 0, 300, 300, color.RGBA{255, 255, 255, 255})
	seed := uint32(12345)
	for y := 60; y < 140; y++ {
		for x := 60; x < 140; x++ {
			seed = seed*1664525 + 1013904223
			r := uint8(seed >> 24)
			g := uint8(seed >> 16)
			b := uint8(seed >> 8)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	boxes := Detect(img)
	found := false
	for _, bx := range boxes {
		if bx.Type == "avatar" || bx.Type == "image" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected avatar/image for noisy square, got %+v", boxes)
	}
}

func fillRect(img *image.RGBA, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			img.Set(x, y, c)
		}
	}
}

func drawOutline(img *image.RGBA, x0, y0, w, h, t int, c color.RGBA) {
	fillRect(img, x0, y0, w, t, c)
	fillRect(img, x0, y0+h-t, w, t, c)
	fillRect(img, x0, y0, t, h, c)
	fillRect(img, x0+w-t, y0, t, h, c)
}
