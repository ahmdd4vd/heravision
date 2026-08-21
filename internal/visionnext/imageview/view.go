package imageview

import (
	"errors"
	"image"
	"math"
)

type View struct {
	Width     int
	Height    int
	Luminance []float64
	ChromaR   []float64
	ChromaB   []float64
}

func (v View) Size() int { return v.Width * v.Height }

func (v View) At(x, y int) (luminance, chromaR, chromaB float64, ok bool) {
	if x < 0 || y < 0 || x >= v.Width || y >= v.Height {
		return 0, 0, 0, false
	}
	i := y*v.Width + x
	return v.Luminance[i], v.ChromaR[i], v.ChromaB[i], true
}

// FromImage creates a bounded canonical view. The downsampling path uses
// nearest sampling intentionally for a deterministic, dependency-free first
// implementation; a later phase can replace it with area/Lanczos sampling.
func FromImage(img image.Image, maxSide int) (View, error) {
	if img == nil {
		return View{}, errors.New("image is nil")
	}
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return View{}, errors.New("image has empty bounds")
	}
	w, h := b.Dx(), b.Dy()
	if maxSide > 0 {
		longest := w
		if h > longest {
			longest = h
		}
		if longest > maxSide {
			scale := float64(maxSide) / float64(longest)
			w = max(1, int(float64(w)*scale+0.5))
			h = max(1, int(float64(h)*scale+0.5))
		}
	}
	v := View{
		Width:     w,
		Height:    h,
		Luminance: make([]float64, w*h),
		ChromaR:   make([]float64, w*h),
		ChromaB:   make([]float64, w*h),
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := b.Min.X + min(b.Dx()-1, int(float64(x)*float64(b.Dx())/float64(w)))
			srcY := b.Min.Y + min(b.Dy()-1, int(float64(y)*float64(b.Dy())/float64(h)))
			r, g, bl, _ := img.At(srcX, srcY).RGBA()
			fr := float64(r>>8) / 255
			fg := float64(g>>8) / 255
			fb := float64(bl>>8) / 255
			yv := 0.2126*fr + 0.7152*fg + 0.0722*fb
			i := y*w + x
			v.Luminance[i] = yv
			v.ChromaR[i] = math.Log(fr+1e-4) - math.Log(yv+1e-4)
			v.ChromaB[i] = math.Log(fb+1e-4) - math.Log(yv+1e-4)
		}
	}
	return v, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
