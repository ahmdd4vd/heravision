package eval

import "heravision/internal/visionnext/schema"

type RegionMatch struct {
	AIndex int     `json:"a_index"`
	BIndex int     `json:"b_index"`
	IoU    float64 `json:"iou"`
}

type MatchSummary struct {
	Threshold    float64       `json:"threshold"`
	Matches      []RegionMatch `json:"matches"`
	UnmatchedA   []int         `json:"unmatched_a"`
	UnmatchedB   []int         `json:"unmatched_b"`
	MeanIoU      float64       `json:"mean_iou"`
	CoverageA    float64       `json:"coverage_a"`
	NovelB       int           `json:"novel_b"`
	FragmentRate float64       `json:"fragment_rate"`
}

func MatchRegions(a, b []schema.Region, threshold float64) MatchSummary {
	if threshold <= 0 {
		threshold = 0.5
	}
	usedB := make([]bool, len(b))
	matches := make([]RegionMatch, 0)
	unmatchedA := make([]int, 0)
	for i, ar := range a {
		bestIndex := -1
		bestIoU := 0.0
		for j, br := range b {
			if usedB[j] {
				continue
			}
			iou := IoU(ar.BBox, br.BBox)
			if iou > bestIoU {
				bestIoU = iou
				bestIndex = j
			}
		}
		if bestIndex < 0 || bestIoU < threshold {
			unmatchedA = append(unmatchedA, i)
			continue
		}
		usedB[bestIndex] = true
		matches = append(matches, RegionMatch{AIndex: i, BIndex: bestIndex, IoU: bestIoU})
	}
	unmatchedB := make([]int, 0)
	for j, used := range usedB {
		if !used {
			unmatchedB = append(unmatchedB, j)
		}
	}
	mean := 0.0
	for _, m := range matches {
		mean += m.IoU
	}
	if len(matches) > 0 {
		mean /= float64(len(matches))
	}
	coverage := 0.0
	if len(a) > 0 {
		coverage = float64(len(matches)) / float64(len(a))
	}
	fragmentRate := 0.0
	if len(a) > 0 {
		fragmentRate = float64(len(b)) / float64(len(a))
	}
	return MatchSummary{
		Threshold: threshold, Matches: matches, UnmatchedA: unmatchedA, UnmatchedB: unmatchedB,
		MeanIoU: mean, CoverageA: coverage, NovelB: len(unmatchedB), FragmentRate: fragmentRate,
	}
}

func IoU(a, b schema.Rect) float64 {
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
