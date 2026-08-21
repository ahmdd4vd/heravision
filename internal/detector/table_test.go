package detector

import (
	"image"
	"image/color"
	"testing"
)

func drawTableFixture(img *image.RGBA, x0, y0, w, h, cols, rows int) {
	black := color.RGBA{0, 0, 0, 255}
	cw := w / cols
	rh := h / rows
	for c := 0; c <= cols; c++ {
		x := x0 + c*cw
		if c == cols {
			x = x0 + w - 2
		}
		for k := 0; k < 2; k++ {
			for y := y0; y <= y0+h; y++ {
				img.Set(x+k, y, black)
			}
		}
	}
	for r := 0; r <= rows; r++ {
		y := y0 + r*rh
		if r == rows {
			y = y0 + h - 2
		}
		for k := 0; k < 2; k++ {
			for x := x0; x <= x0+w; x++ {
				img.Set(x, y+k, black)
			}
		}
	}
}

func TestDetectTables3x2(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	drawTableFixture(img, 50, 50, 240, 140, 3, 2)
	tables := DetectTables(img, DefaultParams)
	if len(tables) == 0 {
		t.Fatal("expected table detection")
	}
	tb := tables[0]
	if tb.Rows != 2 || tb.Cols != 3 {
		t.Fatalf("expected 2x3 got %+v", tb)
	}
	if tb.X < 30 || tb.X > 70 || tb.Y < 30 || tb.Y > 70 {
		t.Fatalf("table position off: %+v", tb)
	}
}

func TestNoTableOnPlainUI(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 100; y < 160; y++ {
		for x := 50; x < 350; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	if tables := DetectTables(img, DefaultParams); len(tables) != 0 {
		t.Fatalf("solid rect must not be a table got %+v", tables)
	}
}
