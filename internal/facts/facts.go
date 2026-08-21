package facts

import (
	"fmt"
	"math"
	"strings"
	"time"

	"heravision/internal/color"
	"heravision/internal/config"
	"heravision/internal/detector"
	"heravision/internal/diagram"
	"heravision/internal/layout"
	"heravision/internal/ocr"
	"heravision/internal/processor"
)

type Meta struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Mode        string `json:"mode"`
	Path        string `json:"path"`
	Version     string `json:"version,omitempty"`
	ElapsedMs   int64  `json:"elapsed_ms,omitempty"`
	Orientation int    `json:"orientation,omitempty"`
}

type Colors struct {
	Dominant   []string `json:"dominant"`
	Background string   `json:"background"`
}

type Result struct {
	Meta     Meta           `json:"meta"`
	Texts    []ocr.Text     `json:"texts"`
	Boxes    []detector.Box `json:"boxes"`
	Colors   Colors         `json:"colors"`
	Layout   layout.Node    `json:"layout"`
	Lines    []interface{}  `json:"lines"`
	Mermaid  string         `json:"mermaid"`
	Markdown string         `json:"markdown"`
}

type Move struct {
	From detector.Box `json:"from"`
	To   detector.Box `json:"to"`
}

type Diff struct {
	Added        []detector.Box `json:"added"`
	Removed      []detector.Box `json:"removed"`
	Moved        []Move         `json:"moved"`
	ColorChanged []Move         `json:"color_changed"`
}

type CompareResult struct {
	PathA  string         `json:"path_a"`
	PathB  string         `json:"path_b"`
	Diff   Diff           `json:"diff"`
	Counts map[string]int `json:"counts"`
}

const compareSide = 512

func Extract(path, mode, version string, cfg config.Config) (*Result, error) {
	start := time.Now()
	img, _, err := processor.Decode(path)
	if err != nil {
		return nil, err
	}
	orientation := processor.ReadOrientation(path)
	img = processor.AutoRotate(img, orientation)
	img = processor.PreprocessCfg(img, mode, cfg.Preprocess.BlurThreshold)
	img = processor.Resize(img, cfg.MaxSide)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	params := detector.Params{
		CannyLow:  cfg.Detector.CannyLow,
		CannyHigh: cfg.Detector.CannyHigh,
		MinArea:   cfg.Detector.MinArea,
	}
	boxes := detector.DetectCfg(img, params)
	texts := ocr.Extract(img)
	dominant := color.DominantCfg(img, cfg.Color.K, cfg.Color.K, cfg.Color.DeltaEMerge)
	bg := color.Background(img)
	if len(dominant) == 0 {
		dominant = []string{"#FFFFFF"}
	}
	if bg == "" {
		bg = dominant[0]
	}
	tree := layout.Build(boxes, w, h)
	var mermaid string
	if mode == "diagram" {
		mermaid = diagram.ToMermaid(boxes)
	}
	r := &Result{
		Meta: Meta{
			Width: w, Height: h, Mode: mode, Path: path,
			Version: version, Orientation: orientation,
			ElapsedMs: time.Since(start).Milliseconds(),
		},
		Texts:   texts,
		Boxes:   boxes,
		Colors:  Colors{Dominant: dominant, Background: bg},
		Layout:  tree,
		Lines:   []interface{}{},
		Mermaid: mermaid,
	}
	r.Markdown = BuildMarkdown(r)
	return r, nil
}

func BuildMarkdown(r *Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Image Facts (%s)\n", r.Meta.Mode))
	sb.WriteString(fmt.Sprintf("- Size: %dx%d px\n", r.Meta.Width, r.Meta.Height))
	if r.Meta.Orientation > 1 {
		sb.WriteString(fmt.Sprintf("- EXIF orientation: %d (auto-rotated)\n", r.Meta.Orientation))
	}
	sb.WriteString(fmt.Sprintf("- Elements detected: %d boxes\n", len(r.Boxes)))
	for i, bx := range r.Boxes {
		if i >= 15 {
			sb.WriteString(fmt.Sprintf("  - ... and %d more\n", len(r.Boxes)-i))
			break
		}
		c := ""
		if bx.Color != "" {
			c = " " + bx.Color
		}
		sb.WriteString(fmt.Sprintf("  - %s at (%d,%d) %dx%d%s score %.2f\n", bx.Type, bx.X, bx.Y, bx.W, bx.H, c, bx.Score))
	}
	sb.WriteString("- Text content: NOT OCR-read; text fields are shape placeholders like [button]/[text]\n")
	for _, t := range r.Texts {
		sb.WriteString(fmt.Sprintf("  - %s at (%d,%d) %dx%d\n", t.Text, t.X, t.Y, t.W, t.H))
	}
	sb.WriteString(fmt.Sprintf("- Dominant colors: %v\n", r.Colors.Dominant))
	sb.WriteString(fmt.Sprintf("- Background: %s\n", r.Colors.Background))
	if r.Mermaid != "" {
		sb.WriteString("\n```mermaid\n" + r.Mermaid + "```\n")
	}
	return sb.String()
}

func Compare(pathA, pathB string, cfg config.Config) (*CompareResult, error) {
	imgA, _, err := processor.Decode(pathA)
	if err != nil {
		return nil, fmt.Errorf("decode a: %w", err)
	}
	imgB, _, err := processor.Decode(pathB)
	if err != nil {
		return nil, fmt.Errorf("decode b: %w", err)
	}
	imgA = processor.Resize(imgA, compareSide)
	imgB = processor.Resize(imgB, compareSide)
	params := detector.Params{
		CannyLow:  cfg.Detector.CannyLow,
		CannyHigh: cfg.Detector.CannyHigh,
		MinArea:   cfg.Detector.MinArea,
	}
	boxesA := detector.DetectCfg(imgA, params)
	boxesB := detector.DetectCfg(imgB, params)
	added, removed, moved, colorChanged := diffBoxes(boxesA, boxesB)
	return &CompareResult{
		PathA: pathA,
		PathB: pathB,
		Diff: Diff{
			Added: added, Removed: removed,
			Moved: moved, ColorChanged: colorChanged,
		},
		Counts: map[string]int{"a_boxes": len(boxesA), "b_boxes": len(boxesB)},
	}, nil
}

func diffBoxes(a, b []detector.Box) ([]detector.Box, []detector.Box, []Move, []Move) {
	added := []detector.Box{}
	removed := []detector.Box{}
	moved := []Move{}
	colorChanged := []Move{}
	used := make([]bool, len(b))
	for _, ba := range a {
		found := -1
		for j, bb := range b {
			if used[j] {
				continue
			}
			if iou(ba, bb) > 0.5 {
				found = j
				break
			}
		}
		if found == -1 {
			removed = append(removed, ba)
			continue
		}
		used[found] = true
		bb := b[found]
		if abs(ba.X-bb.X) > 5 || abs(ba.Y-bb.Y) > 5 {
			moved = append(moved, Move{From: ba, To: bb})
		}
		if ba.Color != "" && bb.Color != "" && colorDistance(ba.Color, bb.Color) > 25 {
			colorChanged = append(colorChanged, Move{From: ba, To: bb})
		}
	}
	for j, bb := range b {
		if !used[j] {
			added = append(added, bb)
		}
	}
	return added, removed, moved, colorChanged
}

func colorDistance(hexA, hexB string) float64 {
	ar, ag, ab := parseHex(hexA)
	br, bg, bb := parseHex(hexB)
	dr, dg, db := float64(ar-br), float64(ag-bg), float64(ab-bb)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func parseHex(s string) (int, int, int) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0
	}
	v := make([]int, 3)
	for i := 0; i < 3; i++ {
		hi, ok1 := hexVal(s[i*2])
		lo, ok2 := hexVal(s[i*2+1])
		if !ok1 || !ok2 {
			return 0, 0, 0
		}
		v[i] = hi*16 + lo
	}
	return v[0], v[1], v[2]
}

func hexVal(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10, true
	}
	return 0, false
}

func iou(a, b detector.Box) float64 {
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.W, b.X+b.W)
	y2 := min(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	inter := float64((x2 - x1) * (y2 - y1))
	union := float64(a.W*a.H+b.W*b.H) - inter
	if union == 0 {
		return 0
	}
	return inter / union
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
