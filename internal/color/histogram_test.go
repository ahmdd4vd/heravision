package color

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestBackgroundSolidRed(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	if got := Background(img); got != "#FF0000" {
		t.Fatalf("expected #FF0000 got %s", got)
	}
}

func TestDominantSolidColorClose(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 120, 90))
	for y := 0; y < 90; y++ {
		for x := 0; x < 120; x++ {
			img.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}
	dom := Dominant(img, 3)
	if len(dom) == 0 {
		t.Fatal("expected dominant colors")
	}
	r, g, b := parseHexT(dom[0])
	dr, dg, db := float64(r-0), float64(g-0), float64(b-255)
	d := math.Sqrt(dr*dr + dg*dg + db*db)
	if d > 12 {
		t.Fatalf("dominant too far from blue: %s (dist %.1f)", dom[0], d)
	}
}

func TestLabRoundtripWhiteBlack(t *testing.T) {
	cases := [][3]uint8{{255, 255, 255}, {0, 0, 0}, {59, 130, 246}}
	for _, c := range cases {
		l, a, bb := rgbToLab(c[0], c[1], c[2])
		r, g, bl := labToRGB(l, a, bb)
		dr, dg, db := int(r)-int(c[0]), int(g)-int(c[1]), int(bl)-int(c[2])
		if abs(dr) > 2 || abs(dg) > 2 || abs(db) > 2 {
			t.Fatalf("lab roundtrip drift for %v: got (%d,%d,%d)", c, r, g, bl)
		}
	}
}

func parseHexT(s string) (int, int, int) {
	if len(s) != 7 || s[0] != '#' {
		return 999, 999, 999
	}
	v := make([]int, 3)
	for i := 0; i < 3; i++ {
		var hi, lo int
		var ok bool
		if hi, ok = hexDigit(s[1+i*2]); !ok {
			return 999, 999, 999
		}
		if lo, ok = hexDigit(s[2+i*2]); !ok {
			return 999, 999, 999
		}
		v[i] = hi*16 + lo
	}
	return v[0], v[1], v[2]
}

func hexDigit(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10, true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	}
	return 0, false
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
