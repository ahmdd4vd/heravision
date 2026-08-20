package processor

import (
	"image"

	"github.com/disintegration/imaging"
)

func FixOrientation(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > h*3 || h > w*3 {
		return img
	}
	return imaging.Clone(img)
}

func AutoRotate(img image.Image, orientation int) image.Image {
	switch orientation {
	case 3:
		return imaging.Rotate180(img)
	case 6:
		return imaging.Rotate270(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}
