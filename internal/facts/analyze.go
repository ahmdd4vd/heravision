package facts

import (
	"fmt"
	"sort"
	"strings"

	"heravision/internal/detector"
)

var namedColors = []struct {
	name string
	r, g, b float64
}{
	{"black", 0, 0, 0},
	{"white", 255, 255, 255},
	{"red", 215, 45, 45},
	{"green", 45, 160, 60},
	{"blue", 45, 85, 220},
	{"yellow", 230, 200, 40},
	{"orange", 240, 140, 30},
	{"purple", 130, 60, 180},
	{"pink", 230, 120, 170},
	{"brown", 140, 90, 50},
}

func colorName(hex string) string {
	r, g, b := parseHex(hex)
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	mx := max(max(r, g), b)
	mn := min(min(r, g), b)
	if mx-mn < 30 {
		switch {
		case lum < 40:
			return "black"
		case lum > 220:
			return "white"
		case lum > 160:
			return "light gray"
		case lum < 80:
			return "dark gray"
		default:
			return "gray"
		}
	}
	best := "gray"
	bestD := 1e12
	for _, nc := range namedColors {
		dr, dg, db := float64(r)-nc.r, float64(g)-nc.g, float64(b)-nc.b
		d := dr*dr + dg*dg + db*db
		if d < bestD {
			bestD = d
			best = nc.name
		}
	}
	if bestD > 40000 {
		return "gray"
	}
	return best
}

func positionWords(bx detector.Box, imgW, imgH int) (string, string) {
	cx := float64(bx.X+bx.W/2) / float64(imgW)
	cy := float64(bx.Y+bx.H/2) / float64(imgH)
	h := "center"
	switch {
	case cx < 0.36:
		h = "left"
	case cx > 0.64:
		h = "right"
	}
	v := "middle"
	switch {
	case cy < 0.33:
		v = "top"
	case cy > 0.67:
		v = "bottom"
	}
	return h, v
}

var typePhrases = map[string]string{
	"button":     "button",
	"input":      "input field",
	"card":       "panel",
	"image":      "image region",
	"text_block": "text line",
	"icon":       "icon",
	"checkbox":   "checkbox",
	"avatar":     "avatar",
}

func buildCaption(bx detector.Box, imgW, imgH int) string {
	phrase := typePhrases[bx.Type]
	if phrase == "" {
		phrase = bx.Type
	}
	h, v := positionWords(bx, imgW, imgH)
	areaRatio := float64(bx.W*bx.H) / float64(imgW*imgH)
	size := "medium"
	switch {
	case areaRatio < 0.01:
		size = "small"
	case areaRatio > 0.15:
		size = "large"
	}
	parts := []string{}
	if bx.Color != "" {
		parts = append(parts, colorName(bx.Color))
	}
	parts = append(parts, phrase, fmt.Sprintf("%dx%d", bx.W, bx.H), size)
	cap := strings.Join(parts, " ")
	if h != "center" || v != "middle" {
		cap += fmt.Sprintf(" at %s-%s", v, h)
	} else {
		cap += " at center"
	}
	return cap
}

type Grid struct {
	Cell  string `json:"cell"`
	Cols  int    `json:"cols"`
	Rows  int    `json:"rows"`
	Count int    `json:"count"`
}

func clusterCount(values []int, tol int) int {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int(nil), values...)
	sort.Ints(sorted)
	clusters := 1
	anchor := sorted[0]
	for _, v := range sorted[1:] {
		if v-anchor > tol {
			clusters++
			anchor = v
		}
	}
	return clusters
}

func detectGrids(boxes []detector.Box) []Grid {
	groups := map[string][]detector.Box{}
	for _, b := range boxes {
		switch b.Type {
		case "card", "image", "button", "avatar":
			groups[b.Type] = append(groups[b.Type], b)
		}
	}
	var grids []Grid
	for cell, group := range groups {
		if len(group) < 4 {
			continue
		}
		ws := make([]int, 0, len(group))
		hs := make([]int, 0, len(group))
		for _, b := range group {
			ws = append(ws, b.W)
			hs = append(hs, b.H)
		}
		sort.Ints(ws)
		sort.Ints(hs)
		mw, mh := ws[len(ws)/2], hs[len(hs)/2]
		uniform := true
		xs := make([]int, 0, len(group))
		ys := make([]int, 0, len(group))
		for _, b := range group {
			if b.W < mw*3/4 || b.W > mw*5/4 || b.H < mh*3/4 || b.H > mh*5/4 {
				uniform = false
				break
			}
			xs = append(xs, b.X)
			ys = append(ys, b.Y)
		}
		if !uniform {
			continue
		}
		cols := clusterCount(xs, 10)
		rows := clusterCount(ys, 10)
		if cols >= 2 && rows >= 2 && cols*rows <= len(group)*4/3 {
			grids = append(grids, Grid{Cell: cell, Cols: cols, Rows: rows, Count: len(group)})
		}
	}
	sort.Slice(grids, func(i, j int) bool { return grids[i].Count > grids[j].Count })
	if len(grids) > 3 {
		grids = grids[:3]
	}
	return grids
}

func luminance(hex string) float64 {
	r, g, b := parseHex(hex)
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

func classifyPage(boxes []detector.Box, bgHex string, imgW, imgH int, grids []Grid) (string, float64) {
	counts := map[string]int{}
	for _, b := range boxes {
		counts[b.Type]++
	}
	total := len(boxes)
	inputs := counts["input"]
	buttons := counts["button"]
	texts := counts["text_block"]
	bgLum := luminance(bgHex)

	scores := map[string]float64{}

	if bgLum < 60 && texts >= 3 {
		scores["terminal"] = 0.55 + 0.05*float64(min(texts, 8))
	}

	if inputs >= 1 && inputs <= 3 && buttons <= 2 && total <= 14 {
		centered := 0
		for _, b := range boxes {
			h, v := positionWords(b, imgW, imgH)
			if h == "center" && (v == "middle" || v == "bottom") {
				centered++
			}
		}
		frac := float64(centered) / float64(max(total, 1))
		scores["login"] = 0.35 + 0.45*frac
	}

	for _, g := range grids {
		if g.Cell == "card" && g.Cols*g.Rows >= 4 {
			scores["dashboard"] = 0.6 + 0.03*float64(min(g.Count, 10))
		}
	}

	leftN, rightN := 0, 0
	bottomInput := false
	for _, b := range boxes {
		h, v := positionWords(b, imgW, imgH)
		if b.Type == "card" || b.Type == "text_block" {
			if h == "left" {
				leftN++
			} else if h == "right" {
				rightN++
			}
		}
		if b.Type == "input" && v == "bottom" {
			bottomInput = true
		}
	}
	if min(leftN, rightN) >= 2 && bottomInput {
		scores["chat"] = 0.55 + 0.05*float64(min(leftN+rightN, 8))
	}

	if inputs >= 3 {
		scores["form"] = 0.45 + 0.05*float64(min(inputs, 6))
	}

	bestType := "general"
	bestScore := 0.0
	for t, s := range scores {
		if s > bestScore {
			bestType = t
			bestScore = s
		}
	}
	if bestScore < 0.45 {
		return "general", 0.4
	}
	if bestScore > 0.95 {
		bestScore = 0.95
	}
	return bestType, round2(bestScore)
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildDiffSummary(res *CompareResult) string {
	var parts []string
	d := res.Diff
	if len(d.Moved) > 0 {
		m := d.Moved[0]
		parts = append(parts, fmt.Sprintf("%d element(s) moved (e.g. %s (%d,%d) -> (%d,%d))",
			len(d.Moved), m.From.Type, m.From.X, m.From.Y, m.To.X, m.To.Y))
	}
	if len(d.ColorChanged) > 0 {
		c := d.ColorChanged[0]
		parts = append(parts, fmt.Sprintf("%d element(s) changed color (e.g. %s %s -> %s)",
			len(d.ColorChanged), c.From.Type, c.From.Color, c.To.Color))
	}
	if len(d.Added) > 0 {
		a := d.Added[0]
		parts = append(parts, fmt.Sprintf("%d added (e.g. %s at (%d,%d))",
			len(d.Added), a.Type, a.X, a.Y))
	}
	if len(d.Removed) > 0 {
		r := d.Removed[0]
		parts = append(parts, fmt.Sprintf("%d removed (e.g. %s at (%d,%d))",
			len(d.Removed), r.Type, r.X, r.Y))
	}
	if len(parts) == 0 {
		return "no structural changes detected"
	}
	return strings.Join(parts, "; ")
}
