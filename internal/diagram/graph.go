package diagram

import (
	"fmt"

	"heravision/internal/detector"
	"heravision/internal/layout"
)

type pt struct{ x, y int }

type connector struct {
	pts []pt
	minX, minY, maxX, maxY int
}

func ToMermaidGraph(boxes []detector.Box, edges [][]uint8, tree layout.Node) string {
	out := "flowchart TD\n"
	if len(boxes) == 0 {
		return out + "  A[empty]\n"
	}
	for i, b := range boxes {
		out += nodeLine(i, b)
	}
	conns := findConnectors(edges, boxes)
	graphEdges := [][2]int{}
	for _, c := range conns {
		a, b := mapEndpoints(c, boxes)
		if a >= 0 && b >= 0 && a != b {
			graphEdges = append(graphEdges, [2]int{a, b})
		}
	}
	for _, e := range barConnectorEdges(boxes) {
		graphEdges = append(graphEdges, e)
	}
	if len(graphEdges) > 0 {
		seen := map[[2]int]bool{}
		for _, e := range graphEdges {
			key := [2]int{min(e[0], e[1]), max(e[0], e[1])}
			if seen[key] {
				continue
			}
			seen[key] = true
			out += fmt.Sprintf("  N%d --> N%d\n", e[0], e[1])
		}
		return out
	}
	out += hierarchyEdges(tree)
	return out
}

func barConnectorEdges(boxes []detector.Box) [][2]int {
	var out [][2]int
	for i, b := range boxes {
		bw, bh := b.W, b.H
		if bw >= bh {
			if float64(bw)/float64(bh) < 6 || bh > 30 {
				continue
			}
			cy := b.Y + bh/2
			ia := nearestBoxCenter(pt{b.X - 2, cy}, i, boxes)
			ib := nearestBoxCenter(pt{b.X + b.W + 2, cy}, i, boxes)
			if ia != -1 && ib != -1 && ia != ib {
				out = append(out, [2]int{ia, ib})
			}
		} else {
			if float64(bh)/float64(bw) < 6 || bw > 30 {
				continue
			}
			cx := b.X + bw/2
			ia := nearestBoxCenter(pt{cx, b.Y - 2}, i, boxes)
			ib := nearestBoxCenter(pt{cx, b.Y + b.H + 2}, i, boxes)
			if ia != -1 && ib != -1 && ia != ib {
				out = append(out, [2]int{ia, ib})
			}
		}
	}
	return out
}

func nearestBoxCenter(p pt, exclude int, boxes []detector.Box) int {
	best := -1
	bestD := 140 * 140
	for i, b := range boxes {
		if i == exclude {
			continue
		}
		dx := p.x - (b.X + b.W/2)
		dy := p.y - (b.Y + b.H/2)
		d := dx*dx + dy*dy
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best
}

func nodeLine(i int, b detector.Box) string {
	id := fmt.Sprintf("N%d", i)
	label := fmt.Sprintf("%s #%d", b.Type, b.Order)
	if b.Order == 0 {
		label = fmt.Sprintf("%s %dx%d", b.Type, b.W, b.H)
	}
	switch b.Type {
	case "button":
		return fmt.Sprintf("  %s([%s])\n", id, label)
	case "image":
		return fmt.Sprintf("  %s[[%s]]\n", id, label)
	default:
		return fmt.Sprintf("  %s[%s]\n", id, label)
	}
}

func findConnectors(edges [][]uint8, boxes []detector.Box) []connector {
	h := len(edges)
	if h == 0 {
		return nil
	}
	w := len(edges[0])
	dirs := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}
	visited := make([][]bool, h)
	for y := range visited {
		visited[y] = make([]bool, w)
	}
	var conns []connector
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if edges[y][x] == 0 || visited[y][x] {
				continue
			}
			c := connector{minX: x, maxX: x, minY: y, maxY: y}
			queue := []pt{{x, y}}
			visited[y][x] = true
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				c.pts = append(c.pts, p)
				if p.x < c.minX {
					c.minX = p.x
				}
				if p.x > c.maxX {
					c.maxX = p.x
				}
				if p.y < c.minY {
					c.minY = p.y
				}
				if p.y > c.maxY {
					c.maxY = p.y
				}
				for _, d := range dirs {
					nx, ny := p.x+d[0], p.y+d[1]
					if nx < 0 || nx >= w || ny < 0 || ny >= h || visited[ny][nx] || edges[ny][nx] == 0 {
						continue
					}
					visited[ny][nx] = true
					queue = append(queue, pt{nx, ny})
				}
			}
			if len(c.pts) < 12 || c.maxX-c.minX+c.maxY-c.minY < 30 {
				continue
			}
			a, b := farthestEnds(c)
			ia := nearestBox(a, boxes)
			ib := nearestBox(b, boxes)
			if ia == -1 || ib == -1 || ia == ib {
				continue
			}
			conns = append(conns, c)
		}
	}
	return conns
}

func farthestEnds(c connector) (pt, pt) {
	var a, b pt
	best := -1
	for _, p := range c.pts {
		d := p.x + p.y
		if best == -1 || d < best {
			best = d
			a = p
		}
	}
	best = -1
	for _, p := range c.pts {
		d := p.x + p.y
		if d > best {
			best = d
			b = p
		}
	}
	return a, b
}

func densityAt(c connector, p pt, r int) int {
	cnt := 0
	for _, q := range c.pts {
		dx := q.x - p.x
		dy := q.y - p.y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= r && dy <= r {
			cnt++
		}
	}
	return cnt
}

func nearestBox(p pt, boxes []detector.Box) int {
	best := -1
	bestD := 1 << 30
	for i, b := range boxes {
		cx := b.X + b.W/2
		cy := b.Y + b.H/2
		dx := abs(p.x-cx) - b.W/2
		dy := abs(p.y-cy) - b.H/2
		if dx < 0 {
			dx = 0
		}
		if dy < 0 {
			dy = 0
		}
		d := dx*dx + dy*dy
		if d < bestD {
			bestD = d
			best = i
		}
	}
	if bestD > 90*90 {
		return -1
	}
	return best
}

func mapEndpoints(c connector, boxes []detector.Box) (int, int) {
	a, b := farthestEnds(c)
	da := densityAt(c, a, 6)
	db := densityAt(c, b, 6)
	src, dst := a, b
	if da > db {
		src, dst = b, a
	}
	return nearestBox(src, boxes), nearestBox(dst, boxes)
}

func hierarchyEdges(tree layout.Node) string {
	var lines []string
	var walk func(n layout.Node, parentID string, depth int) string
	id := 0
	walk = func(n layout.Node, parentID string, depth int) string {
		myID := fmt.Sprintf("H%d", id)
		id++
		label := n.Type
		if n.Order > 0 {
			label = fmt.Sprintf("%s #%d", n.Type, n.Order)
		}
		lines = append(lines, fmt.Sprintf("  %s[%s]", myID, label))
		if parentID != "" {
			lines[len(lines)-1] = fmt.Sprintf("  %s --> %s", parentID, myID)
		}
		for _, ch := range n.Children {
			walk(ch, myID, depth+1)
		}
		return myID
	}
	for _, ch := range tree.Children {
		walk(ch, "", 0)
	}
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}
