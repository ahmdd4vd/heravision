package relation

import (
	"fmt"
	"math"

	"heravision/internal/visionnext/schema"
)

func Build(regions []schema.Region) []schema.Relation {
	var out []schema.Relation
	for i := 0; i < len(regions); i++ {
		for j := i + 1; j < len(regions); j++ {
			a, b := regions[i], regions[j]
			out = append(out, pair(a, b)...)
		}
	}
	return out
}

func pair(a, b schema.Region) []schema.Relation {
	var out []schema.Relation
	if a.ID == "" || b.ID == "" {
		return out
	}
	if contains(a.BBox, b.BBox) {
		out = append(out, visible(a.ID, b.ID, "contains", containmentScore(a.BBox, b.BBox)))
	} else if contains(b.BBox, a.BBox) {
		out = append(out, visible(b.ID, a.ID, "contains", containmentScore(b.BBox, a.BBox)))
	}
	if overlap := iou(a.BBox, b.BBox); overlap > 0 {
		out = append(out, visible(a.ID, b.ID, "overlapping", clamp(overlap)))
	}
	if gap := boundaryGap(a.BBox, b.BBox); gap == 0 {
		out = append(out, visible(a.ID, b.ID, "touching", 0.8))
	}
	acx, acy := center(a.BBox)
	bcx, bcy := center(b.BBox)
	if horizontalOverlap(a.BBox, b.BBox) && acy+float64(a.BBox.H)/2 <= bcy-float64(b.BBox.H)/2 {
		out = append(out, visible(a.ID, b.ID, "above", directionalScore(acy, bcy)))
	} else if horizontalOverlap(a.BBox, b.BBox) && bcy+float64(b.BBox.H)/2 <= acy-float64(a.BBox.H)/2 {
		out = append(out, visible(b.ID, a.ID, "above", directionalScore(bcy, acy)))
	}
	if verticalOverlap(a.BBox, b.BBox) && acx+float64(a.BBox.W)/2 <= bcx-float64(b.BBox.W)/2 {
		out = append(out, visible(a.ID, b.ID, "left_of", directionalScore(acx, bcx)))
	} else if verticalOverlap(a.BBox, b.BBox) && bcx+float64(b.BBox.W)/2 <= acx-float64(a.BBox.W)/2 {
		out = append(out, visible(b.ID, a.ID, "left_of", directionalScore(bcx, acx)))
	}
	return out
}

func visible(from, to, predicate string, score float64) schema.Relation {
	return schema.Relation{
		From: from, To: to, Predicate: predicate, Status: "visible", Score: clamp(score),
		Evidence: []schema.EvidenceRef{{Kind: "bbox-geometry", Stage: "relation", RegionID: from, Note: fmt.Sprintf("%s -> %s", predicate, to)}},
	}
}

func center(r schema.Rect) (float64, float64) {
	return float64(r.X) + float64(r.W)/2, float64(r.Y) + float64(r.H)/2
}

func contains(a, b schema.Rect) bool {
	return b.X >= a.X && b.Y >= a.Y && b.X+b.W <= a.X+a.W && b.Y+b.H <= a.Y+a.H && (a.W > b.W || a.H > b.H)
}

func containmentScore(a, b schema.Rect) float64 {
	if a.W <= 0 || a.H <= 0 {
		return 0
	}
	return clamp(float64(b.W*b.H) / float64(a.W*a.H))
}

func iou(a, b schema.Rect) float64 {
	x1, y1 := max(a.X, b.X), max(a.Y, b.Y)
	x2, y2 := min(a.X+a.W, b.X+b.W), min(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	inter := float64((x2 - x1) * (y2 - y1))
	union := float64(a.W*a.H+b.W*b.H) - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

func boundaryGap(a, b schema.Rect) int {
	dx := intervalGap(a.X, a.X+a.W, b.X, b.X+b.W)
	dy := intervalGap(a.Y, a.Y+a.H, b.Y, b.Y+b.H)
	if dx == 0 && dy == 0 {
		return 0
	}
	return int(math.Hypot(float64(dx), float64(dy)))
}

func intervalGap(a0, a1, b0, b1 int) int {
	if a1 < b0 {
		return b0 - a1
	}
	if b1 < a0 {
		return a0 - b1
	}
	return 0
}

func horizontalOverlap(a, b schema.Rect) bool { return min(a.X+a.W, b.X+b.W) > max(a.X, b.X) }
func verticalOverlap(a, b schema.Rect) bool   { return min(a.Y+a.H, b.Y+b.H) > max(a.Y, b.Y) }

func directionalScore(a, b float64) float64 {
	d := math.Abs(b - a)
	return clamp(0.5 + d/100)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
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
