package facts

import (
	"fmt"
	"image"
	"math"
	"sort"
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
	Meta           Meta           `json:"meta"`
	PageType       string         `json:"page_type"`
	PageConfidence float64        `json:"page_confidence"`
	Texts          []ocr.Text     `json:"texts"`
	Boxes          []detector.Box `json:"boxes"`
	Grids          []Grid         `json:"grids"`
	Tables         []detector.Table `json:"tables"`
	Colors         Colors         `json:"colors"`
	Layout         layout.Node    `json:"layout"`
	Lines          []interface{}  `json:"lines"`
	Mermaid        string         `json:"mermaid"`
	Markdown       string         `json:"markdown"`
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
	PathA   string         `json:"path_a"`
	PathB   string         `json:"path_b"`
	Summary string         `json:"summary"`
	Diff    Diff           `json:"diff"`
	Counts  map[string]int `json:"counts"`
}

const compareSide = 512

type scaleRun struct {
	img    image.Image
	factor float64
}

func detectMultiScale(src, base image.Image, params detector.Params, maxSide int, enabled bool) []detector.Box {
	if !enabled {
		return detector.DetectCfg(base, params)
	}
	b := base.Bounds()
	baseW := b.Dx()
	srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()
	srcLongest := max(srcW, srcH)
	var runs []scaleRun
	if srcLongest >= maxSide*3/2 {
		hi := maxSide * 2
		if hi > 2048 {
			hi = 2048
		}
		if hi > srcLongest {
			hi = srcLongest
		}
		if hi > maxSide {
			himg := processor.Resize(src, hi)
			hw := himg.Bounds().Dx()
			if hw > baseW {
				runs = append(runs, scaleRun{img: himg, factor: float64(baseW) / float64(hw)})
			}
		}
	}
	runs = append(runs, scaleRun{img: base, factor: 1})
	half := processor.Resize(src, maxSide/2)
	hw := half.Bounds().Dx()
	if hw < baseW {
		runs = append(runs, scaleRun{img: half, factor: float64(baseW) / float64(hw)})
	}
	var all []detector.Box
	for _, r := range runs {
		for _, bx := range detector.DetectCfg(r.img, params) {
			all = append(all, mapBox(bx, r.factor))
		}
	}
	return mergeScales(all)
}

func mapBox(bx detector.Box, f float64) detector.Box {
	if f == 1 {
		return bx
	}
	return detector.Box{
		Type:  bx.Type,
		X:     int(float64(bx.X)*f + 0.5),
		Y:     int(float64(bx.Y)*f + 0.5),
		W:     int(float64(bx.W)*f + 0.5),
		H:     int(float64(bx.H)*f + 0.5),
		Color: bx.Color,
		Text:  bx.Text,
		Score: bx.Score,
	}
}

func mergeScales(boxes []detector.Box) []detector.Box {
	out := []detector.Box{}
	for _, bx := range boxes {
		dup := false
		for j := range out {
			if iou(bx, out[j]) > 0.55 {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, bx)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y/20 == out[j].Y/20 {
			return out[i].X < out[j].X
		}
		return out[i].Y < out[j].Y
	})
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

func Extract(path, mode, version string, cfg config.Config) (*Result, error) {
	start := time.Now()
	img, _, err := processor.Decode(path)
	if err != nil {
		return nil, err
	}
	orientation := processor.ReadOrientation(path)
	img = processor.AutoRotate(img, orientation)
	img = processor.PreprocessCfg(img, mode, cfg.Preprocess.BlurThreshold)
	base := processor.Resize(img, cfg.MaxSide)
	b := base.Bounds()
	w, h := b.Dx(), b.Dy()
	params := detector.Params{
		CannyLow:  cfg.Detector.CannyLow,
		CannyHigh: cfg.Detector.CannyHigh,
		MinArea:   cfg.Detector.MinArea,
	}
	boxes := detectMultiScale(img, base, params, cfg.MaxSide, cfg.Multiscale)
	edgeMap := detector.EdgeMap(base, params)
	tables := detector.TablesFromEdges(edgeMap)
	if tables == nil {
		tables = []detector.Table{}
	}
	texts := ocr.Extract(base)
	dominant := color.DominantCfg(base, cfg.Color.K, cfg.Color.K, cfg.Color.DeltaEMerge)
	bg := color.Background(base)
	if len(dominant) == 0 {
		dominant = []string{"#FFFFFF"}
	}
	if bg == "" {
		bg = dominant[0]
	}
	for i := range boxes {
		boxes[i].Order = i + 1
		boxes[i].Caption = buildCaption(boxes[i], w, h)
	}
	grids := detectGrids(boxes)
	if grids == nil {
		grids = []Grid{}
	}
	pageType, pageConf := classifyPage(boxes, bg, w, h, grids)
	tree := layout.Build(boxes, w, h)
	var mermaid string
	if mode == "diagram" {
		mermaid = diagram.ToMermaidGraph(boxes, edgeMap, tree)
	}
	r := &Result{
		Meta: Meta{
			Width: w, Height: h, Mode: mode, Path: path,
			Version: version, Orientation: orientation,
			ElapsedMs: time.Since(start).Milliseconds(),
		},
		PageType:       pageType,
		PageConfidence: pageConf,
		Texts:          texts,
		Boxes:          boxes,
		Grids:          grids,
		Tables:         tables,
		Colors:         Colors{Dominant: dominant, Background: bg},
		Layout:         tree,
		Lines:          []interface{}{},
		Mermaid:        mermaid,
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
	sb.WriteString(fmt.Sprintf("- Page type: %s (confidence %.2f)\n", r.PageType, r.PageConfidence))
	for _, g := range r.Grids {
		sb.WriteString(fmt.Sprintf("- Grid: %dx%d of %s (%d cells)\n", g.Cols, g.Rows, g.Cell, g.Count))
	}
	for _, tb := range r.Tables {
		sb.WriteString(fmt.Sprintf("- Table at (%d,%d) %dx%d: %d rows x %d cols\n", tb.X, tb.Y, tb.W, tb.H, tb.Rows, tb.Cols))
	}
	sb.WriteString(fmt.Sprintf("- Elements detected: %d boxes\n", len(r.Boxes)))
	for i, bx := range r.Boxes {
		if i >= 15 {
			sb.WriteString(fmt.Sprintf("  - ... and %d more\n", len(r.Boxes)-i))
			break
		}
		cap := bx.Caption
		if cap == "" {
			cap = fmt.Sprintf("%s at (%d,%d)", bx.Type, bx.X, bx.Y)
		}
		sb.WriteString(fmt.Sprintf("  - #%d %s\n", bx.Order, cap))
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
	res := &CompareResult{
		PathA: pathA,
		PathB: pathB,
		Diff: Diff{
			Added: added, Removed: removed,
			Moved: moved, ColorChanged: colorChanged,
		},
		Counts: map[string]int{"a_boxes": len(boxesA), "b_boxes": len(boxesB)},
	}
	res.Summary = buildDiffSummary(res)
	return res, nil
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
