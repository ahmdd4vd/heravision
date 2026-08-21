package evidence

import (
	"math"

	"heravision/internal/visionnext/imageview"
)

type Field struct {
	Width           int
	Height          int
	Luminance       []float64
	ChromaR         []float64
	ChromaB         []float64
	Edge            []float64
	Orientation     []float64
	LocalContrast   []float64
	Flatness        []float64
	ChromaMagnitude []float64
}

func Compute(v imageview.View) Field {
	f := Field{
		Width:           v.Width,
		Height:          v.Height,
		Luminance:       append([]float64(nil), v.Luminance...),
		ChromaR:         append([]float64(nil), v.ChromaR...),
		ChromaB:         append([]float64(nil), v.ChromaB...),
		Edge:            make([]float64, v.Size()),
		Orientation:     make([]float64, v.Size()),
		LocalContrast:   make([]float64, v.Size()),
		Flatness:        make([]float64, v.Size()),
		ChromaMagnitude: make([]float64, v.Size()),
	}
	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
			i := y*v.Width + x
			center, _, _, _ := v.At(x, y)
			left := sample(v, x-1, y, center)
			right := sample(v, x+1, y, center)
			up := sample(v, x, y-1, center)
			down := sample(v, x, y+1, center)
			gx := (right - left) * 0.5
			gy := (down - up) * 0.5
			f.Edge[i] = math.Min(1, math.Hypot(gx, gy)*2)
			f.Orientation[i] = math.Atan2(gy, gx)
			f.LocalContrast[i] = localStd(v, x, y)
			f.Flatness[i] = 1 / (1 + 8*f.LocalContrast[i])
			_, cr, cb, _ := v.At(x, y)
			f.ChromaMagnitude[i] = math.Hypot(cr, cb)
		}
	}
	return f
}

func sample(v imageview.View, x, y int, fallback float64) float64 {
	l, _, _, ok := v.At(x, y)
	if !ok {
		return fallback
	}
	return l
}

func localStd(v imageview.View, x, y int) float64 {
	var sum, sumsq float64
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			l, _, _, ok := v.At(x+dx, y+dy)
			if !ok {
				continue
			}
			sum += l
			sumsq += l * l
			n++
		}
	}
	if n == 0 {
		return 0
	}
	mean := sum / float64(n)
	variance := sumsq/float64(n) - mean*mean
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}
