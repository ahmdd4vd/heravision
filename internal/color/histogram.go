package color

import (
	"fmt"
	"image"
	"math"
	"sort"
)

func Dominant(img image.Image, n int) []string {
	return DominantCfg(img, n, 5, 12)
}

func DominantCfg(img image.Image, n, k int, deltaEMerge float64) []string {
	samples := collectSamples(img, 8000)
	if len(samples) == 0 {
		return []string{"#FFFFFF"}
	}
	centers := kmeans(samples, k, 8)
	merged := mergeClose(centers, deltaEMerge)
	sort.Slice(merged, func(i, j int) bool { return merged[i].count > merged[j].count })
	out := make([]string, 0, len(merged))
	for _, c := range merged {
		r, g, bl := labToRGB(c.l, c.a, c.b)
		out = append(out, fmt.Sprintf("#%02X%02X%02X", r, g, bl))
	}
	if len(out) == 0 {
		return []string{"#FFFFFF"}
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func Background(img image.Image) string {
	b := img.Bounds()
	var r, g, bl, cnt int
	for y := b.Min.Y; y < b.Min.Y+2 && y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += int(cr >> 8)
			g += int(cg >> 8)
			bl += int(cb >> 8)
			cnt++
		}
	}
	for y := b.Max.Y - 2; y < b.Max.Y; y++ {
		if y < b.Min.Y {
			continue
		}
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += int(cr >> 8)
			g += int(cg >> 8)
			bl += int(cb >> 8)
			cnt++
		}
	}
	if cnt == 0 {
		return "#FFFFFF"
	}
	return fmt.Sprintf("#%02X%02X%02X", r/cnt, g/cnt, bl/cnt)
}

type lab struct {
	l, a, b float64
	count   int
}

func collectSamples(img image.Image, max int) []lab {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	total := w * h
	step := 1
	if total > max {
		step = int(math.Sqrt(float64(total) / float64(max)))
		if step < 1 {
			step = 1
		}
	}
	var out []lab
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			r, g, bl, _ := img.At(x, y).RGBA()
			l, a, bv := rgbToLab(uint8(r>>8), uint8(g>>8), uint8(bl>>8))
			out = append(out, lab{l: l, a: a, b: bv})
		}
	}
	return out
}

func kmeans(samples []lab, k, iter int) []lab {
	if k > len(samples) {
		k = len(samples)
	}
	if k == 0 {
		return nil
	}
	centers := make([]lab, k)
	for i := 0; i < k; i++ {
		centers[i] = samples[i*len(samples)/k]
		centers[i].count = 0
	}
	for it := 0; it < iter; it++ {
		sums := make([]lab, k)
		counts := make([]int, k)
		for _, s := range samples {
			best := 0
			bestD := 1e9
			for i, c := range centers {
				d := deltaE(s.l, s.a, s.b, c.l, c.a, c.b)
				if d < bestD {
					bestD = d
					best = i
				}
			}
			sums[best].l += s.l
			sums[best].a += s.a
			sums[best].b += s.b
			counts[best]++
		}
		for i := range centers {
			if counts[i] > 0 {
				centers[i].l = sums[i].l / float64(counts[i])
				centers[i].a = sums[i].a / float64(counts[i])
				centers[i].b = sums[i].b / float64(counts[i])
				centers[i].count = counts[i]
			}
		}
	}
	return centers
}

func mergeClose(centers []lab, thresh float64) []lab {
	var out []lab
	for _, c := range centers {
		merged := false
		for i, o := range out {
			if deltaE(c.l, c.a, c.b, o.l, o.a, o.b) < thresh {
				total := float64(out[i].count + c.count)
				out[i].l = (out[i].l*float64(out[i].count) + c.l*float64(c.count)) / total
				out[i].a = (out[i].a*float64(out[i].count) + c.a*float64(c.count)) / total
				out[i].b = (out[i].b*float64(out[i].count) + c.b*float64(c.count)) / total
				out[i].count += c.count
				merged = true
				break
			}
		}
		if !merged {
			out = append(out, c)
		}
	}
	return out
}

func deltaE(l1, a1, b1, l2, a2, b2 float64) float64 {
	dl := l1 - l2
	da := a1 - a2
	db := b1 - b2
	return math.Sqrt(dl*dl + da*da + db*db)
}

func rgbToLab(r, g, b uint8) (float64, float64, float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255
	rf = gammaInv(rf)
	gf = gammaInv(gf)
	bf = gammaInv(bf)
	x := rf*0.4124564 + gf*0.3575761 + bf*0.1804375
	y := rf*0.2126729 + gf*0.7151522 + bf*0.0721750
	z := rf*0.0193339 + gf*0.1191920 + bf*0.9503041
	x /= 0.95047
	z /= 1.08883
	fx := labF(x)
	fy := labF(y)
	fz := labF(z)
	l := 116*fy - 16
	a := 500 * (fx - fy)
	bb := 200 * (fy - fz)
	return l, a, bb
}

func labToRGB(l, a, b float64) (uint8, uint8, uint8) {
	fy := (l + 16) / 116
	fx := fy + a/500
	fz := fy - b/200
	x := labFInv(fx) * 0.95047
	y := labFInv(fy)
	z := labFInv(fz) * 1.08883
	r := x*3.2404542 + y*-1.5371385 + z*-0.4985314
	g := x*-0.9692660 + y*1.8760108 + z*0.0415560
	bl := x*0.0556434 + y*-0.2040259 + z*1.0572252
	r = gamma(r)
	g = gamma(g)
	bl = gamma(bl)
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	if g < 0 {
		g = 0
	}
	if g > 1 {
		g = 1
	}
	if bl < 0 {
		bl = 0
	}
	if bl > 1 {
		bl = 1
	}
	return uint8(r*255 + 0.5), uint8(g*255 + 0.5), uint8(bl*255 + 0.5)
}

func gammaInv(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func gamma(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

func labF(t float64) float64 {
	if t > 0.008856 {
		return math.Pow(t, 1.0/3.0)
	}
	return 7.787*t + 16.0/116
}

func labFInv(t float64) float64 {
	if t > 0.206893 {
		return t * t * t
	}
	return (t - 16.0/116) / 7.787
}
