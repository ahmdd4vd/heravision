package ocr

import "image"

type WasmEngine struct {
	loaded bool
	path   string
}

func NewWasmEngine(modelPath string) *WasmEngine { return &WasmEngine{path: modelPath} }

func (w *WasmEngine) Available() bool { return w.loaded }

func (w *WasmEngine) Extract(img image.Image) []Text {
	if !w.loaded {
		return heuristicExtract(img)
	}
	return heuristicExtract(img)
}

func (w *WasmEngine) Load() error {
	w.loaded = false
	return nil
}
