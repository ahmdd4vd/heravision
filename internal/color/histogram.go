package color

import (
	"fmt"
	"image"
	"sort"
)

func Dominant(img image.Image, n int) []string {
	b := img.Bounds()
	buckets := make(map[uint32]int)
	total := 0
	step := 1
	if b.Dx()*b.Dy() > 300000 {
		step = 2
	}
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			r, g, b2, _ := img.At(x, y).RGBA()
			qr := uint32((r>>8)/32) * 32
			qg := uint32((g>>8)/32) * 32
			qb := uint32((b2>>8)/32) * 32
			key := (qr << 16) | (qg << 8) | qb
			buckets[key]++
			total++
		}
	}
	type kv struct {
		k uint32
		v int
	}
	var list []kv
	for k, v := range buckets {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
	if n > len(list) {
		n = len(list)
	}
	if n == 0 {
		return []string{"#FFFFFF"}
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		k := list[i].k
		r := (k >> 16) & 0xFF
		g := (k >> 8) & 0xFF
		bl := k & 0xFF
		out = append(out, fmt.Sprintf("#%02X%02X%02X", r, g, bl))
	}
	return out
}
