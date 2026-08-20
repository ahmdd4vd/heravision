package diagram

import (
	"fmt"
	"sort"

	"heravision/internal/detector"
)

func ToMermaid(boxes []detector.Box) string {
	if len(boxes) == 0 {
		return "flowchart TD\n  A[empty]"
	}
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].Y == boxes[j].Y {
			return boxes[i].X < boxes[j].X
		}
		return boxes[i].Y < boxes[j].Y
	})
	out := "flowchart TD\n"
	for i, b := range boxes {
		id := fmt.Sprintf("N%d", i)
		label := fmt.Sprintf("%s %dx%d", b.Type, b.W, b.H)
		switch b.Type {
		case "card":
			out += fmt.Sprintf("  %s[%s]\n", id, label)
		case "button":
			out += fmt.Sprintf("  %s([%s])\n", id, label)
		case "image":
			out += fmt.Sprintf("  %s[[%s]]\n", id, label)
		default:
			out += fmt.Sprintf("  %s[%s]\n", id, label)
		}
		if i > 0 {
			prev := fmt.Sprintf("N%d", i-1)
			if abs(b.Y-boxes[i-1].Y) < 80 {
				out += fmt.Sprintf("  %s --> %s\n", prev, id)
			} else {
				out += fmt.Sprintf("  %s -.-> %s\n", prev, id)
			}
		}
	}
	return out
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
