package processor

import (
	"encoding/binary"
	"image"
	"testing"
)

func TestParseOrientationLittleEndian(t *testing.T) {
	tiff := make([]byte, 22)
	tiff[0], tiff[1] = 'I', 'I'
	binary.LittleEndian.PutUint16(tiff[2:4], 42)
	binary.LittleEndian.PutUint32(tiff[4:8], 8)
	binary.LittleEndian.PutUint16(tiff[8:10], 1)
	binary.LittleEndian.PutUint16(tiff[10:12], 0x0112)
	binary.LittleEndian.PutUint16(tiff[12:14], 3)
	binary.LittleEndian.PutUint32(tiff[14:18], 1)
	binary.LittleEndian.PutUint16(tiff[18:20], 6)
	if got := parseOrientation(tiff); got != 6 {
		t.Fatalf("expected 6 got %d", got)
	}
}

func TestParseOrientationBigEndian(t *testing.T) {
	tiff := make([]byte, 22)
	tiff[0], tiff[1] = 'M', 'M'
	binary.BigEndian.PutUint16(tiff[2:4], 42)
	binary.BigEndian.PutUint32(tiff[4:8], 8)
	binary.BigEndian.PutUint16(tiff[8:10], 1)
	binary.BigEndian.PutUint16(tiff[10:12], 0x0112)
	binary.BigEndian.PutUint16(tiff[12:14], 3)
	binary.BigEndian.PutUint32(tiff[14:18], 1)
	binary.BigEndian.PutUint16(tiff[18:20], 8)
	if got := parseOrientation(tiff); got != 8 {
		t.Fatalf("expected 8 got %d", got)
	}
}

func TestParseOrientationJunk(t *testing.T) {
	if got := parseOrientation([]byte("junkjunk")); got != 0 {
		t.Fatalf("expected 0 got %d", got)
	}
	if got := parseOrientation(nil); got != 0 {
		t.Fatalf("expected 0 got %d", got)
	}
}

func TestReadOrientationPNGReturnsZero(t *testing.T) {
	if got := ReadOrientation("../../testdata/ui.png"); got != 0 {
		t.Fatalf("png has no exif, expected 0 got %d", got)
	}
}

func TestAutoRotateIdentityAndFlips(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	if AutoRotate(img, 1) == nil {
		t.Fatal("orientation 1 must return image")
	}
	flipped := AutoRotate(img, 2)
	if flipped.Bounds().Dx() != 4 || flipped.Bounds().Dy() != 2 {
		t.Fatal("flipH keeps dimensions")
	}
	rotated := AutoRotate(img, 6)
	if rotated.Bounds().Dx() != 2 || rotated.Bounds().Dy() != 4 {
		t.Fatal("rotate90cw swaps dimensions")
	}
}
