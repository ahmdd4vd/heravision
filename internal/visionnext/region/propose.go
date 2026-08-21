package region

import (
	"fmt"
	"sort"

	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/schema"
)

type Config struct {
	MergeThreshold float64
	MinArea        int
	MaxRegions     int
}

func DefaultConfig() Config {
	return Config{MergeThreshold: 0.20, MinArea: 8, MaxRegions: 256}
}

type dsu struct {
	parent []int
	size   []int
}

func newDSU(n int) *dsu {
	p := make([]int, n)
	s := make([]int, n)
	for i := range p {
		p[i] = i
		s[i] = 1
	}
	return &dsu{parent: p, size: s}
}

func (d *dsu) find(x int) int {
	for d.parent[x] != x {
		d.parent[x] = d.parent[d.parent[x]]
		x = d.parent[x]
	}
	return x
}

func (d *dsu) union(a, b int) {
	ra, rb := d.find(a), d.find(b)
	if ra == rb {
		return
	}
	if d.size[ra] < d.size[rb] {
		ra, rb = rb, ra
	}
	d.parent[rb] = ra
	d.size[ra] += d.size[rb]
}

func Propose(f evidence.Field, cfg Config) []schema.Region {
	if f.Width <= 0 || f.Height <= 0 || len(f.Luminance) != f.Width*f.Height {
		return nil
	}
	if cfg.MergeThreshold <= 0 {
		cfg.MergeThreshold = DefaultConfig().MergeThreshold
	}
	if cfg.MinArea < 1 {
		cfg.MinArea = DefaultConfig().MinArea
	}
	if cfg.MaxRegions < 1 {
		cfg.MaxRegions = DefaultConfig().MaxRegions
	}
	n := f.Width * f.Height
	d := newDSU(n)
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			i := y*f.Width + x
			if x+1 < f.Width && mergeCost(f, x, y, x+1, y) <= cfg.MergeThreshold {
				d.union(i, i+1)
			}
			if y+1 < f.Height && mergeCost(f, x, y, x, y+1) <= cfg.MergeThreshold {
				d.union(i, i+f.Width)
			}
		}
	}

	type aggregate struct {
		minX, minY, maxX, maxY int
		area                   int
		lum, cr, cb, texture   float64
		boundary, perimeter    float64
	}
	aggs := make(map[int]*aggregate)
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			i := y*f.Width + x
			root := d.find(i)
			a := aggs[root]
			if a == nil {
				a = &aggregate{minX: x, minY: y, maxX: x, maxY: y}
				aggs[root] = a
			}
			a.area++
			if x < a.minX {
				a.minX = x
			}
			if y < a.minY {
				a.minY = y
			}
			if x > a.maxX {
				a.maxX = x
			}
			if y > a.maxY {
				a.maxY = y
			}
			a.lum += f.Luminance[i]
			a.cr += f.ChromaR[i]
			a.cb += f.ChromaB[i]
			a.texture += f.LocalContrast[i]
		}
	}
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			i := y*f.Width + x
			root := d.find(i)
			a := aggs[root]
			for _, n := range [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
				if n[0] < 0 || n[1] < 0 || n[0] >= f.Width || n[1] >= f.Height {
					a.perimeter++
					continue
				}
				j := n[1]*f.Width + n[0]
				if d.find(j) != root {
					a.perimeter++
					a.boundary += f.Edge[i]
				}
			}
		}
	}

	type item struct {
		root int
		a    *aggregate
	}
	items := make([]item, 0, len(aggs))
	for root, a := range aggs {
		if a.area >= cfg.MinArea {
			items = append(items, item{root: root, a: a})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].a.area == items[j].a.area {
			if items[i].a.minY == items[j].a.minY {
				return items[i].a.minX < items[j].a.minX
			}
			return items[i].a.minY < items[j].a.minY
		}
		return items[i].a.area > items[j].a.area
	})
	if len(items) > cfg.MaxRegions {
		items = items[:cfg.MaxRegions]
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].a.minY == items[j].a.minY {
			return items[i].a.minX < items[j].a.minX
		}
		return items[i].a.minY < items[j].a.minY
	})
	out := make([]schema.Region, 0, len(items))
	for i, it := range items {
		a := it.a
		w, h := a.maxX-a.minX+1, a.maxY-a.minY+1
		boundary := 0.0
		if a.perimeter > 0 {
			boundary = a.boundary / a.perimeter
		}
		out = append(out, schema.Region{
			ID:   fmt.Sprintf("r-%04d", i+1),
			BBox: schema.Rect{X: a.minX, Y: a.minY, W: w, H: h},
			Area: a.area,
			Features: schema.Features{
				AreaRatio:        float64(a.area) / float64(f.Width*f.Height),
				AspectRatio:      float64(w) / float64(h),
				Compactness:      float64(a.perimeter*a.perimeter) / float64(max(1, a.area)),
				BoundaryStrength: boundary,
				ScaleStability:   1,
				Color:            []float64{a.lum / float64(a.area), a.cr / float64(a.area), a.cb / float64(a.area)},
				Texture:          []float64{a.texture / float64(a.area)},
			},
			Evidence: []schema.EvidenceRef{{Kind: "region-membership", Stage: "graph-merge", Value: float64(a.area)}},
		})
	}
	return out
}

func mergeCost(f evidence.Field, ax, ay, bx, by int) float64 {
	a := ay*f.Width + ax
	b := by*f.Width + bx
	lum := clamp01(abs(f.Luminance[a]-f.Luminance[b]) * 4)
	chroma := clamp01((abs(f.ChromaR[a]-f.ChromaR[b]) + abs(f.ChromaB[a]-f.ChromaB[b])) * 0.5)
	texture := clamp01(abs(f.LocalContrast[a]-f.LocalContrast[b]) * 4)
	barrier := (f.Edge[a] + f.Edge[b]) * 0.5
	return 0.45*lum + 0.2*chroma + 0.1*texture + 0.25*barrier
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
