package detector

import (
	"image"
	"sort"
)

type Box struct {
	Type  string  `json:"type"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	W     int     `json:"w"`
	H     int     `json:"h"`
	Color string  `json:"color,omitempty"`
	Text  string  `json:"text,omitempty"`
	Score float64 `json:"score,omitempty"`
}

func Detect(img image.Image) []Box {
	gray := toGray(img)
	edges := sobel(gray)
	binary := threshold(edges, 30)
	components := findComponents(binary)
	boxes := make([]Box, 0, len(components))
	b := img.Bounds()
	imgW, imgH := b.Dx(), b.Dy()
	for _, c := range components {
		if c.w < 20 || c.h < 12 {
			continue
		}
		area := c.w * c.h
		if area < 400 {
			continue
		}
		if area > imgW*imgH*9/10 {
			continue
		}
		box := classify(c, imgW, imgH)
		boxes = append(boxes, box)
	}
	boxes = dedup(boxes)
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].Y == boxes[j].Y {
			return boxes[i].X < boxes[j].X
		}
		return boxes[i].Y < boxes[j].Y
	})
	if len(boxes) > 30 {
		boxes = boxes[:30]
	}
	return boxes
}

func dedup(boxes []Box) []Box {
	out := []Box{}
	for _, b := range boxes {
		overlap := false
		for j, o := range out {
			if iou(b, o) > 0.6 {
				if b.W*b.H > o.W*o.H {
					out[j] = b
				}
				overlap = true
				break
			}
		}
		if !overlap {
			out = append(out, b)
		}
	}
	return out
}

func iou(a, b Box) float64 {
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.W, b.X+b.W)
	y2 := min(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	inter := float64((x2 - x1) * (y2 - y1))
	union := float64(a.W*a.H + b.W*b.H) - inter
	if union == 0 {
		return 0
	}
	return inter / union
}

type comp struct{ x, y, w, h int }

func toGray(img image.Image) [][]uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	gray := make([][]uint8, h)
	for y := 0; y < h; y++ {
		row := make([]uint8, w)
		for x := 0; x < w; x++ {
			r, g, b2, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			v := uint8((19595*int(r>>8) + 38470*int(g>>8) + 7471*int(b2>>8)) >> 16)
			row[x] = v
		}
		gray[y] = row
	}
	return gray
}

func sobel(gray [][]uint8) [][]uint8 {
	h := len(gray)
	if h == 0 {
		return nil
	}
	w := len(gray[0])
	out := make([][]uint8, h)
	for y := range out {
		out[y] = make([]uint8, w)
	}
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			gx := int(gray[y-1][x+1]) + 2*int(gray[y][x+1]) + int(gray[y+1][x+1]) - int(gray[y-1][x-1]) - 2*int(gray[y][x-1]) - int(gray[y+1][x-1])
			gy := int(gray[y+1][x-1]) + 2*int(gray[y+1][x]) + int(gray[y+1][x+1]) - int(gray[y-1][x-1]) - 2*int(gray[y-1][x]) - int(gray[y-1][x+1])
			mag := abs(gx) + abs(gy)
			if mag > 255 {
				mag = 255
			}
			out[y][x] = uint8(mag)
		}
	}
	return out
}

func threshold(img [][]uint8, t uint8) [][]uint8 {
	h := len(img)
	out := make([][]uint8, h)
	for y, row := range img {
		nr := make([]uint8, len(row))
		for x, v := range row {
			if v > t {
				nr[x] = 255
			}
		}
		out[y] = nr
	}
	return out
}

func findComponents(bin [][]uint8) []comp {
	h := len(bin)
	if h == 0 {
		return nil
	}
	w := len(bin[0])
	visited := make([][]bool, h)
	for i := range visited {
		visited[i] = make([]bool, w)
	}
	var comps []comp
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bin[y][x] == 0 || visited[y][x] {
				continue
			}
			minX, minY, maxX, maxY := x, y, x, y
			queue := [][2]int{{x, y}}
			visited[y][x] = true
			for len(queue) > 0 {
				cx, cy := queue[0][0], queue[0][1]
				queue = queue[1:]
				if cx < minX {
					minX = cx
				}
				if cx > maxX {
					maxX = cx
				}
				if cy < minY {
					minY = cy
				}
				if cy > maxY {
					maxY = cy
				}
				for _, d := range dirs {
					nx, ny := cx+d[0], cy+d[1]
					if nx < 0 || nx >= w || ny < 0 || ny >= h {
						continue
					}
					if visited[ny][nx] || bin[ny][nx] == 0 {
						continue
					}
					visited[ny][nx] = true
					queue = append(queue, [2]int{nx, ny})
				}
			}
			comps = append(comps, comp{x: minX, y: minY, w: maxX - minX + 1, h: maxY - minY + 1})
		}
	}
	return comps
}

func classify(c comp, imgW, imgH int) Box {
	ar := float64(c.w) / float64(c.h)
	area := c.w * c.h
	typ := "card"
	switch {
	case ar > 3 && c.h < 60 && c.h > 18:
		typ = "input"
	case ar > 1.5 && ar < 5 && c.w > 60 && c.h > 28 && c.h < 80 && area < 30000:
		typ = "button"
	case ar < 1.2 && c.w > 40 && c.h > 40 && c.w < 300 && c.h < 300:
		typ = "image"
	case c.w > imgW*6/10 && c.h > 60:
		typ = "card"
	case c.w > 200 && c.h < 30:
		typ = "text_block"
	}
	score := float64(area) / float64(imgW*imgH)
	if score > 1 {
		score = 1
	}
	return Box{Type: typ, X: c.x, Y: c.y, W: c.w, H: c.h, Score: score}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
