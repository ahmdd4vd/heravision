package ocr

import "image"

type Text struct {
	Text string `json:"text"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	W    int    `json:"w"`
	H    int    `json:"h"`
	Size int    `json:"size,omitempty"`
}

func Extract(img image.Image) []Text {
	return []Text{}
}
