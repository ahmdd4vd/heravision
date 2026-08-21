package processor

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/webp"
)

const MaxSideLimit = 16384

var MaxPixels int64 = 12_000_000

func Decode(path string) (image.Image, string, error) {
	return DecodeWithMaxPixels(path, MaxPixels)
}

func DecodeWithMaxPixels(path string, maxPixels int64) (image.Image, string, error) {
	if maxPixels <= 0 {
		maxPixels = MaxPixels
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, "", fmt.Errorf("decode %s: %w", path, err)
	}
	w, h := cfg.Width, cfg.Height
	if w <= 0 || h <= 0 {
		return nil, "", fmt.Errorf("decode %s: invalid dimensions %dx%d", path, w, h)
	}
	if w > MaxSideLimit || h > MaxSideLimit {
		return nil, "", fmt.Errorf("decode %s: dimensions %dx%d exceed limit %dpx per side", path, w, h, MaxSideLimit)
	}
	if px := int64(w) * int64(h); px > maxPixels {
		return nil, "", fmt.Errorf("decode %s: %d megapixels exceeds limit %d (max_pixels)", path, px/1_000_000, maxPixels/1_000_000)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, "", fmt.Errorf("seek %s: %w", path, err)
	}
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("decode %s: %w", path, err)
	}
	return img, format, nil
}
