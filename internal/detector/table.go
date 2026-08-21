package detector

import (
	"image"
	"sort"
)

type Table struct {
	X    int `json:"x"`
	Y    int `json:"y"`
	W    int `json:"w"`
	H    int `json:"h"`
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

type seg struct {
	pos      int
	lo, hi   int
	consumed bool
}

func EdgeMap(img image.Image, p Params) [][]uint8 {
	return canny(gaussian3x3(toGray(img)), p.CannyLow, p.CannyHigh)
}

func DetectTables(img image.Image, p Params) []Table {
	return TablesFromEdges(EdgeMap(img, p))
}

func TablesFromEdges(edges [][]uint8) []Table {
	h := len(edges)
	if h == 0 {
		return nil
	}
	w := len(edges[0])
	minH := w / 6
	if minH < 40 {
		minH = 40
	}
	minV := h / 6
	if minV < 40 {
		minV = 40
	}
	hLines := scanSegments(edges, true, minH)
	vLines := scanSegments(edges, false, minV)
	if len(hLines) < 2 || len(vLines) < 2 {
		return nil
	}
	sort.Slice(hLines, func(i, j int) bool { return hLines[i].pos < hLines[j].pos })
	sort.Slice(vLines, func(i, j int) bool { return vLines[i].pos < vLines[j].pos })
	var tables []Table
	for i := 0; i < len(hLines); i++ {
		if hLines[i].consumed {
			continue
		}
		band := []seg{hLines[i]}
		for j := i + 1; j < len(hLines); j++ {
			if hLines[j].pos-hLines[i].pos > h/2 {
				break
			}
			if overlapRatio(hLines[i], hLines[j]) > 0.6 {
				band = append(band, hLines[j])
			}
		}
		if len(band) < 2 {
			continue
		}
		top, bottom := band[0], band[len(band)-1]
		var cross []seg
		for _, v := range vLines {
			if v.pos >= top.lo-8 && v.pos <= top.hi+8 && v.pos >= bottom.lo-8 && v.pos <= bottom.hi+8 &&
				v.lo <= top.pos+8 && v.hi >= bottom.pos-8 {
				cross = append(cross, v)
			}
		}
		if len(cross) < 2 {
			continue
		}
		x0, x1 := top.lo, top.hi
		for _, b := range band {
			if b.lo < x0 {
				x0 = b.lo
			}
			if b.hi > x1 {
				x1 = b.hi
			}
		}
		t := Table{
			X: x0, Y: top.pos,
			W:    x1 - x0,
			H:    bottom.pos - top.pos,
			Rows: len(band) - 1,
			Cols: len(cross) - 1,
		}
		if t.W < 60 || t.H < 40 || t.Rows < 1 || t.Cols < 1 {
			continue
		}
		if t.Rows <= 1 && t.Cols <= 1 {
			continue
		}
		tables = append(tables, t)
		for k := range band {
			band[k].consumed = true
		}
		for k := range cross {
			cross[k].consumed = true
		}
	}
	return tables
}

func scanSegments(edges [][]uint8, horizontal bool, minLen int) []seg {
	h := len(edges)
	w := len(edges[0])
	outer, inner := h, w
	if horizontal {
		outer, inner = h, w
	} else {
		outer, inner = w, h
	}
	at := func(o, i int) uint8 {
		if horizontal {
			return edges[o][i]
		}
		return edges[i][o]
	}
	var raw []seg
	for o := 1; o < outer-1; o++ {
		runStart := -1
		for i := 1; i < inner-1; i++ {
			if at(o, i) != 0 {
				if runStart == -1 {
					runStart = i
				}
				continue
			}
			if runStart != -1 && i-runStart >= minLen {
				raw = append(raw, seg{pos: o, lo: runStart, hi: i - 1})
			}
			runStart = -1
		}
		if runStart != -1 && inner-1-runStart >= minLen {
			raw = append(raw, seg{pos: o, lo: runStart, hi: inner - 2})
		}
	}
	var merged []seg
	const gapTol = 14
	for _, s := range raw {
		found := false
		for k := range merged {
			m := &merged[k]
			if abs(s.pos-m.pos) <= 2 && s.lo <= m.hi+gapTol && s.hi >= m.lo-gapTol {
				nlo, nhi := m.lo, m.hi
				if s.lo < nlo {
					nlo = s.lo
				}
				if s.hi > nhi {
					nhi = s.hi
				}
				m.lo, m.hi = nlo, nhi
				m.pos = (m.pos + s.pos) / 2
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, s)
		}
	}
	for pass := 0; pass < 3; pass++ {
		changed := false
		for i := 0; i < len(merged); i++ {
			for j := i + 1; j < len(merged); j++ {
				a, b := &merged[i], &merged[j]
				if abs(a.pos-b.pos) <= 2 && a.lo <= b.hi+gapTol && a.hi >= b.lo-gapTol {
					if b.lo < a.lo {
						a.lo = b.lo
					}
					if b.hi > a.hi {
						a.hi = b.hi
					}
					a.pos = (a.pos + b.pos) / 2
					merged = append(merged[:j], merged[j+1:]...)
					j--
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
	return merged
}

func overlapRatio(a, b seg) float64 {
	lo := max(a.lo, b.lo)
	hi := min(a.hi, b.hi)
	if hi <= lo {
		return 0
	}
	shorter := min(a.hi-a.lo, b.hi-b.lo)
	if shorter <= 0 {
		return 0
	}
	return float64(hi-lo) / float64(shorter)
}
