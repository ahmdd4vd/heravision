package processor

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
	"os"

	"github.com/disintegration/imaging"
)

var (
	errNotJPEG = errors.New("not a jpeg file")
	errNoExif  = errors.New("no exif segment")
)

func ReadOrientation(path string) int {
	data, err := readExifBlock(path)
	if err != nil {
		return 0
	}
	return parseOrientation(data)
}

func readExifBlock(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var soi [2]byte
	if _, err := io.ReadFull(f, soi[:]); err != nil {
		return nil, err
	}
	if soi[0] != 0xFF || soi[1] != 0xD8 {
		return nil, errNotJPEG
	}
	for {
		var marker [2]byte
		if _, err := io.ReadFull(f, marker[:]); err != nil {
			return nil, err
		}
		if marker[0] != 0xFF {
			return nil, errNotJPEG
		}
		m := marker[1]
		if m == 0x01 || (m >= 0xD0 && m <= 0xD7) {
			continue
		}
		if m == 0xD8 {
			continue
		}
		var sizeBytes [2]byte
		if _, err := io.ReadFull(f, sizeBytes[:]); err != nil {
			return nil, err
		}
		size := int(binary.BigEndian.Uint16(sizeBytes[:]))
		if size < 2 {
			return nil, errNotJPEG
		}
		payload := make([]byte, size-2)
		if _, err := io.ReadFull(f, payload); err != nil {
			return nil, err
		}
		if m == 0xE1 && len(payload) > 6 && string(payload[:6]) == "Exif\x00\x00" {
			return payload[6:], nil
		}
		if m == 0xDA {
			return nil, errNoExif
		}
	}
}

func parseOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 0
	}
	var order binary.ByteOrder = binary.LittleEndian
	switch {
	case tiff[0] == 'I' && tiff[1] == 'I':
		order = binary.LittleEndian
	case tiff[0] == 'M' && tiff[1] == 'M':
		order = binary.BigEndian
	default:
		return 0
	}
	if order.Uint16(tiff[2:4]) != 42 {
		return 0
	}
	off := int(order.Uint32(tiff[4:8]))
	if off < 8 || off+2 > len(tiff) {
		return 0
	}
	count := int(order.Uint16(tiff[off : off+2]))
	for i := 0; i < count; i++ {
		e := tiff[off+2+i*12:]
		if len(e) < 12 {
			return 0
		}
		tag := order.Uint16(e[0:2])
		typ := order.Uint16(e[2:4])
		if tag == 0x0112 && typ == 3 {
			v := order.Uint16(e[8:10])
			if v >= 1 && v <= 8 {
				return int(v)
			}
			return 0
		}
	}
	return 0
}

func AutoRotate(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.FlipH(imaging.Rotate270(img))
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.FlipH(imaging.Rotate90(img))
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}
