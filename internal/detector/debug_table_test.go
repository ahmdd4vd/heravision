package detector

import (
	"image"
	"image/color"
	"testing"
)

func TestDebugTableSegs(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	drawTableFixture(img, 50, 50, 240, 140, 3, 2)
	edges := EdgeMap(img, DefaultParams)
	hLines := scanSegments(edges, true, 66)
	vLines := scanSegments(edges, false, 50)
	t.Logf("hLines=%d vLines=%d", len(hLines), len(vLines))
	for _, l := range hLines {
		t.Logf("H y=%d x[%d..%d]", l.pos, l.lo, l.hi)
	}
	for _, l := range vLines {
		t.Logf("V x=%d y[%d..%d]", l.pos, l.lo, l.hi)
	}
}
