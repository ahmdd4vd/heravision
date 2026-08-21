package ocr

import (
	"image"
	"os"
	"strings"
	"sync"

	"github.com/getcharzp/go-ocr/paddle"
)

type OnnxEngine struct {
	mu     sync.Mutex
	cfg    paddle.Config
	engine *paddle.Engine
	failed bool
}

var (
	onnxMu   sync.Mutex
	onnxInst *OnnxEngine
)

func ConfigureOnnx(libPath, detPath, recPath, dictPath string, detMaxSideLen, threadCount, numThreads int) {
	onnxMu.Lock()
	defer onnxMu.Unlock()
	if libPath == "" || detPath == "" || recPath == "" || dictPath == "" {
		onnxInst = nil
		return
	}
	for _, p := range []string{libPath, detPath, recPath, dictPath} {
		if _, err := os.Stat(p); err != nil {
			onnxInst = nil
			return
		}
	}
	onnxInst = &OnnxEngine{cfg: paddle.Config{
		OnnxRuntimeLibPath: libPath,
		DetModelPath:       detPath,
		RecModelPath:       recPath,
		DictPath:           dictPath,
		DetMaxSideLen:      detMaxSideLen,
		ThreadCount:        threadCount,
		NumThreads:         numThreads,
	}}
}

func currentOnnx() *OnnxEngine {
	onnxMu.Lock()
	defer onnxMu.Unlock()
	return onnxInst
}

// BundlePresent reports whether all OCR sidecar files exist on disk.
func BundlePresent(libPath, detPath, recPath, dictPath string) bool {
	if libPath == "" || detPath == "" || recPath == "" || dictPath == "" {
		return false
	}
	for _, p := range []string{libPath, detPath, recPath, dictPath} {
		if _, err := os.Stat(p); err != nil {
			return false
		}
	}
	return true
}

func (e *OnnxEngine) ensure() (*paddle.Engine, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.engine != nil {
		return e.engine, nil
	}
	if e.failed {
		return nil, errOnnxUnavailable
	}
	eng, err := paddle.NewEngine(e.cfg)
	if err != nil {
		e.failed = true
		return nil, err
	}
	e.engine = eng
	return eng, nil
}

func (e *OnnxEngine) Available() bool {
	_, err := e.ensure()
	return err == nil
}

func (e *OnnxEngine) Extract(img image.Image) []Text {
	eng, err := e.ensure()
	if err != nil {
		return heuristicExtract(img)
	}
	results, err := eng.RunOCR(img)
	if err != nil {
		return heuristicExtract(img)
	}
	out := make([]Text, 0, len(results))
	for _, r := range results {
		t := strings.TrimSpace(r.Text)
		if t == "" {
			continue
		}
		w := r.Box[2] - r.Box[0]
		h := r.Box[3] - r.Box[1]
		out = append(out, Text{
			Text: t,
			X:    r.Box[0],
			Y:    r.Box[1],
			W:    w,
			H:    h,
			Size: h,
			Conf: float64(r.Score),
		})
	}
	return out
}

func CloseOnnx() {
	inst := currentOnnx()
	if inst == nil {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if inst.engine != nil {
		inst.engine.Destroy()
		inst.engine = nil
	}
}
