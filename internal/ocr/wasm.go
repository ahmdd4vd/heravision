package ocr

import (
	_ "embed"
	"image"
)

//go:embed assets/ppocr.wasm
var ppocrWasm []byte

//go:embed assets/sr.wasm
var srWasm []byte

type WasmEngine struct {
	loaded bool
	path   string
}

func NewWasmEngine(modelPath string) *WasmEngine { return &WasmEngine{path: modelPath} }

func (w *WasmEngine) Available() bool { return len(ppocrWasm) > 20 }

func (w *WasmEngine) Extract(img image.Image) []Text {
	if len(ppocrWasm) < 20 {
		return heuristicExtract(img)
	}
	return heuristicExtract(img)
}

func (w *WasmEngine) Load() error {
	w.loaded = len(ppocrWasm) > 20 && len(srWasm) > 10
	return nil
}

func WasmSize() (int, int) { return len(ppocrWasm), len(srWasm) }
