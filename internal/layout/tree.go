package layout

import (
	"sort"

	"heravision/internal/detector"
)

type Node struct {
	Type     string `json:"type"`
	X        int    `json:"x,omitempty"`
	Y        int    `json:"y,omitempty"`
	W        int    `json:"w,omitempty"`
	H        int    `json:"h,omitempty"`
	Order    int    `json:"order,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Children []Node `json:"children,omitempty"`
}

const (
	maxDepth = 6
	minGap   = 12
)

func Build(boxes []detector.Box, imgW, imgH int) Node {
	root := Node{Type: "root", W: imgW, H: imgH}
	if len(boxes) == 0 {
		root.Children = []Node{{Type: "region", X: 0, Y: 0, W: imgW, H: imgH}}
		return root
	}
	root.Children = xyCut(boxes, true, 0)
	return root
}

func xyCut(boxes []detector.Box, horizontal bool, depth int) []Node {
	if len(boxes) == 0 {
		return nil
	}
	if len(boxes) == 1 || depth >= maxDepth {
		return boxLeaves(boxes)
	}
	groups := splitByGaps(boxes, horizontal)
	if len(groups) == 1 {
		if horizontal {
			return xyCut(boxes, false, depth+1)
		}
		return flatRegion(boxes)
	}
	containerType := "col"
	if horizontal {
		containerType = "row"
	}
	var nodes []Node
	for _, g := range groups {
		kids := xyCut(g, !horizontal, depth+1)
		if len(kids) == 1 {
			nodes = append(nodes, kids[0])
			continue
		}
		nodes = append(nodes, Node{
			Type:     containerType,
			X:        bboxMin(g, true),
			Y:        bboxMin(g, false),
			W:        bboxSize(g, true),
			H:        bboxSize(g, false),
			Children: kids,
		})
	}
	return nodes
}

func boxLeaves(boxes []detector.Box) []Node {
	out := make([]Node, 0, len(boxes))
	for _, b := range boxes {
		out = append(out, Node{
			Type: b.Type, X: b.X, Y: b.Y, W: b.W, H: b.H,
			Order: b.Order, Caption: b.Caption,
		})
	}
	return out
}

func flatRegion(boxes []detector.Box) []Node {
	return []Node{{
		Type: "region", X: bboxMin(boxes, true), Y: bboxMin(boxes, false),
		W: bboxSize(boxes, true), H: bboxSize(boxes, false),
		Children: boxLeaves(boxes),
	}}
}

func splitByGaps(boxes []detector.Box, horizontal bool) [][]detector.Box {
	sorted := append([]detector.Box(nil), boxes...)
	if horizontal {
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Y < sorted[j].Y })
	} else {
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].X < sorted[j].X })
	}
	type interval struct{ lo, hi int }
	intervals := make([]interval, len(sorted))
	for i, b := range sorted {
		if horizontal {
			intervals[i] = interval{b.Y, b.Y + b.H}
		} else {
			intervals[i] = interval{b.X, b.X + b.W}
		}
	}
	var groups [][]detector.Box
	start := 0
	curHi := intervals[0].hi
	for i := 1; i < len(sorted); i++ {
		if intervals[i].lo-curHi >= minGap {
			groups = append(groups, sorted[start:i])
			start = i
		}
		if intervals[i].hi > curHi {
			curHi = intervals[i].hi
		}
	}
	groups = append(groups, sorted[start:])
	return groups
}

func bboxMin(boxes []detector.Box, x bool) int {
	m := 1 << 30
	for _, b := range boxes {
		v := b.X
		if !x {
			v = b.Y
		}
		if v < m {
			m = v
		}
	}
	return m
}

func bboxSize(boxes []detector.Box, x bool) int {
	lo, hi := 1<<30, 0
	for _, b := range boxes {
		a, c := b.X, b.X+b.W
		if !x {
			a, c = b.Y, b.Y+b.H
		}
		if a < lo {
			lo = a
		}
		if c > hi {
			hi = c
		}
	}
	return hi - lo
}
