package ocr

import (
	"errors"
	"image"

	"heravision/internal/detector"
)

var errOnnxUnavailable = errors.New("onnx engine unavailable")

type Text struct {
	Text string  `json:"text"`
	X    int     `json:"x"`
	Y    int     `json:"y"`
	W    int     `json:"w"`
	H    int     `json:"h"`
	Size int     `json:"size,omitempty"`
	Conf float64 `json:"conf,omitempty"`
}

type Engine interface {
	Extract(img image.Image) []Text
	Available() bool
}

var engine Engine = heuristicEngine{}

func Extract(img image.Image) []Text {
	if inst := currentOnnx(); inst != nil && inst.Available() {
		return inst.Extract(img)
	}
	if engine != nil && engine.Available() {
		return engine.Extract(img)
	}
	return heuristicExtract(img)
}

func SetEngine(e Engine) { engine = e }

func OcrReady() bool {
	return currentOnnx() != nil
}

type heuristicEngine struct{}

func (heuristicEngine) Available() bool            { return true }
func (heuristicEngine) Extract(img image.Image) []Text { return heuristicExtract(img) }

func heuristicExtract(img image.Image) []Text {
	boxes := detector.Detect(img)
	var out []Text
	for _, b := range boxes {
		if b.Type == "text_block" || (b.H >= 14 && b.H <= 40 && b.W > 30 && b.W < 600) {
			t := guessText(b)
			out = append(out, Text{Text: t, X: b.X, Y: b.Y, W: b.W, H: b.H, Size: b.H})
		}
	}
	if out == nil {
		out = []Text{}
	}
	return out
}

func guessText(b detector.Box) string {
	switch b.Type {
	case "button":
		return "[button]"
	case "input":
		return "[input]"
	case "text_block":
		return "[text]"
	default:
		if b.W > b.H*3 {
			return "[text line]"
		}
		return "[block]"
	}
}
