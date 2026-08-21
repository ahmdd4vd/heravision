package detector

import (
	"fmt"
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

type Params struct {
	CannyLow  uint8
	CannyHigh uint8
	MinArea   int
}

var DefaultParams = Params{CannyLow: 50, CannyHigh: 150, MinArea: 200}

func Detect(img image.Image) []Box {
	return DetectCfg(img, DefaultParams)
}

func DetectCfg(img image.Image, p Params) []Box {
	gray := toGray(img)
	blurred := gaussian3x3(gray)
	edges := canny(blurred, p.CannyLow, p.CannyHigh)
	closed := morphClose(edges, 2)
	components := findComponents(closed)
	b := img.Bounds()
	imgW, imgH := b.Dx(), b.Dy()
	boxes := make([]Box, 0, len(components))
	for _, c := range components {
		if c.w < 12 || c.h < 10 {
			continue
		}
		area := c.w * c.h
		minArea := imgW * imgH / 5000
		if minArea < p.MinArea {
			minArea = p.MinArea
		}
		if area < minArea {
			continue
		}
		if area > imgW*imgH*9/10 {
			continue
		}
		ed := edgeDensity(closed, c)
		box := classifyV2(c, imgW, imgH, ed)
		box.Color = avgColor(img, c)
		boxes = append(boxes, box)
	}
	boxes = dedup(boxes)
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].Y/20 == boxes[j].Y/20 {
			return boxes[i].X < boxes[j].X
		}
		return boxes[i].Y < boxes[j].Y
	})
	if len(boxes) > 40 {
		boxes = boxes[:40]
	}
	return boxes
}

func avgColor(img image.Image, c comp) string {
	step := (c.w / 16)
	if step < 2 {
		step = 2
	}
	var r, g, bl, n int
	for y := c.y + 1; y < c.y+c.h-1; y += step {
		for x := c.x + 1; x < c.x+c.w-1; x += step {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += int(cr >> 8)
			g += int(cg >> 8)
			bl += int(cb >> 8)
			n++
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("#%02X%02X%02X", r/n, g/n, bl/n)
}

func dedup(boxes []Box) []Box {
	out := []Box{}
	for _, b := range boxes {
		overlap := false
		for j, o := range out {
			if iou(b, o) > 0.45 {
				if b.Score > o.Score {
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
	union := float64(a.W*a.H+b.W*b.H) - inter
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

func gaussian3x3(gray [][]uint8) [][]uint8 {
	h := len(gray)
	if h == 0 {
		return gray
	}
	w := len(gray[0])
	out := make([][]uint8, h)
	for y := range out {
		out[y] = make([]uint8, w)
		copy(out[y], gray[y])
	}
	kernel := [3][3]int{{1, 2, 1}, {2, 4, 2}, {1, 2, 1}}
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			sum := 0
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					sum += int(gray[y+ky][x+kx]) * kernel[ky+1][kx+1]
				}
			}
			out[y][x] = uint8(sum / 16)
		}
	}
	return out
}

func canny(gray [][]uint8, low, high uint8) [][]uint8 {
	edges := sobel(gray)
	strong := threshold(edges, high)
	weak := threshold(edges, low)
	h := len(edges)
	if h == 0 {
		return edges
	}
	w := len(edges[0])
	out := make([][]uint8, h)
	for y := range out {
		out[y] = make([]uint8, w)
		copy(out[y], strong[y])
	}
	dirs := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}
	changed := true
	for changed {
		changed = false
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				if out[y][x] != 0 || weak[y][x] == 0 {
					continue
				}
				for _, d := range dirs {
					if out[y+d[1]][x+d[0]] != 0 {
						out[y][x] = 255
						changed = true
						break
					}
				}
			}
		}
	}
	return out
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

func morphClose(bin [][]uint8, iter int) [][]uint8 {
	h := len(bin)
	if h == 0 {
		return bin
	}
	w := len(bin[0])
	cur := bin
	for k := 0; k < iter; k++ {
		dilated := make([][]uint8, h)
		for y := range dilated {
			dilated[y] = make([]uint8, w)
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if cur[y][x] != 0 {
					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							ny, nx := y+dy, x+dx
							if ny >= 0 && ny < h && nx >= 0 && nx < w {
								dilated[ny][nx] = 255
							}
						}
					}
				}
			}
		}
		eroded := make([][]uint8, h)
		for y := range eroded {
			eroded[y] = make([]uint8, w)
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				all := true
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						ny, nx := y+dy, x+dx
						if ny < 0 || ny >= h || nx < 0 || nx >= w || dilated[ny][nx] == 0 {
							all = false
						}
					}
				}
				if all {
					eroded[y][x] = 255
				}
			}
		}
		cur = eroded
	}
	return cur
}

func edgeDensity(bin [][]uint8, c comp) float64 {
	cnt := 0
	for y := c.y; y < c.y+c.h; y++ {
		if y < 0 || y >= len(bin) {
			continue
		}
		for x := c.x; x < c.x+c.w; x++ {
			if x < 0 || x >= len(bin[0]) {
				continue
			}
			if bin[y][x] != 0 {
				cnt++
			}
		}
	}
	area := c.w * c.h
	if area == 0 {
		return 0
	}
	return float64(cnt) / float64(area)
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
	dirs := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}
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

func classifyV2(c comp, imgW, imgH int, ed float64) Box {
	ar := float64(c.w) / float64(c.h)
	area := c.w * c.h
	typ := "card"
	switch {
	case ar > 3.2 && c.h < 70 && c.h > 14 && ed < 0.3:
		typ = "input"
	case ar > 1.4 && ar < 6 && c.w > 50 && c.h > 20 && c.h < 85 && area < 40000 && ed > 0.08 && ed < 0.5:
		typ = "button"
	case ar < 1.25 && c.w > 35 && c.h > 35 && c.w < 400 && c.h < 400 && ed < 0.4:
		typ = "image"
	case c.w > imgW*5/10 && c.h > 50:
		typ = "card"
	case ed > 0.02 && ed < 0.25 && c.h < 35 && c.w > 25:
		typ = "text_block"
	}
	score := float64(area)/float64(imgW*imgH)*0.5 + ed*0.5
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
