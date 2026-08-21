package processor

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

func BlurMetric(gray [][]uint8) float64 {
	h := len(gray)
	if h < 3 {
		return 1000
	}
	w := len(gray[0])
	var sum, sumsq float64
	n := 0
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			lap := int(gray[y-1][x]) + int(gray[y+1][x]) + int(gray[y][x-1]) + int(gray[y][x+1]) - 4*int(gray[y][x])
			if lap < 0 {
				lap = -lap
			}
			sum += float64(lap)
			sumsq += float64(lap * lap)
			n++
		}
	}
	if n == 0 {
		return 1000
	}
	mean := sum / float64(n)
	variance := sumsq/float64(n) - mean*mean
	if variance < 0 {
		variance = 0
	}
	return variance
}

func Preprocess(img image.Image, mode string) image.Image {
	return PreprocessCfg(img, mode, 80)
}

func PreprocessCfg(img image.Image, mode string, blurThreshold float64) image.Image {
	gray := toGrayQuick(img)
	bm := BlurMetric(gray)
	if mode == "blur" || bm < blurThreshold {
		img = imaging.AdjustContrast(img, 15)
		img = imaging.Sharpen(img, 0.8)
		if isSmallText(img) {
			img = imaging.Resize(img, img.Bounds().Dx()*2, img.Bounds().Dy()*2, imaging.Lanczos)
		}
	} else if mode == "diagram" {
		img = imaging.AdjustContrast(img, 10)
	} else if bm < blurThreshold*2 {
		img = imaging.Sharpen(img, 0.4)
	}
	return img
}

func toGrayQuick(img image.Image) [][]uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	gray := make([][]uint8, h)
	for y := 0; y < h; y++ {
		row := make([]uint8, w)
		for x := 0; x < w; x++ {
			r, g, b2, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			v := uint8((19595*int(r>>8) + 38470*int(g>>8) + 7471*int(b2>>8)) >> 16)
			row[x] = v
		}
		gray[y] = row
	}
	return gray
}

func isSmallText(img image.Image) bool {
	b := img.Bounds()
	return b.Dx() < 800 || b.Dy() < 600
}

func SauvolaBinarize(gray [][]uint8, window int, k float64) [][]uint8 {
	h := len(gray)
	if h == 0 || window < 1 {
		return gray
	}
	w := len(gray[0])
	out := make([][]uint8, h)
	for y := range out {
		out[y] = make([]uint8, w)
	}
	half := window / 2
	if half < 1 {
		half = 1
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum, sumsq float64
			n := 0
			for dy := -half; dy <= half; dy++ {
				ny := y + dy
				if ny < 0 || ny >= h {
					continue
				}
				for dx := -half; dx <= half; dx++ {
					nx := x + dx
					if nx < 0 || nx >= w {
						continue
					}
					v := float64(gray[ny][nx])
					sum += v
					sumsq += v * v
					n++
				}
			}
			mean := sum / float64(n)
			variance := sumsq/float64(n) - mean*mean
			if variance < 0 {
				variance = 0
			}
			std := math.Sqrt(variance)
			thresh := mean * (1 + k*(std/128-1))
			if float64(gray[y][x]) > thresh {
				out[y][x] = 255
			} else {
				out[y][x] = 0
			}
		}
	}
	return out
}

func OtsuThreshold(hist [256]int, total int) uint8 {
	var sum int
	for i := 0; i < 256; i++ {
		sum += i * hist[i]
	}
	var sumB, wB, wF int
	var maxVar float64
	thresh := 0
	for i := 0; i < 256; i++ {
		wB += hist[i]
		if wB == 0 {
			continue
		}
		wF = total - wB
		if wF == 0 {
			break
		}
		sumB += i * hist[i]
		mB := float64(sumB) / float64(wB)
		mF := float64(sum-sumB) / float64(wF)
		between := float64(wB) * float64(wF) * (mB - mF) * (mB - mF)
		if between > maxVar {
			maxVar = between
			thresh = i
		}
	}
	return uint8(thresh)
}

func EnhanceContrast(img image.Image) image.Image {
	return imaging.AdjustContrast(img, 20)
}

func Upscale2x(img image.Image) image.Image {
	b := img.Bounds()
	return imaging.Resize(img, b.Dx()*2, b.Dy()*2, imaging.Lanczos)
}
