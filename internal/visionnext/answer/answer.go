package answer

import "heravision/internal/visionnext/schema"

// FromRegions produces a deliberately non-semantic answer. It only claims
// that stable visual structure was found, and abstains when pixel/region
// evidence is weak or absent.
func FromRegions(regions []schema.Region) schema.Answer {
	return FromRegionsWithMinScore(regions, 0.65)
}

func FromRegionsWithMinScore(regions []schema.Region, minScore float64) schema.Answer {
	if minScore <= 0 {
		minScore = 0.65
	}
	if minScore > 1 {
		minScore = 1
	}
	if len(regions) == 0 {
		return schema.Answer{
			Text:       "insufficient visual evidence",
			Status:     "insufficient_evidence",
			Confidence: 0,
			Warnings:   []schema.Warning{{Code: "no-regions", Message: "no supported visual region was produced"}},
		}
	}
	best := regions[0]
	bestScore := evidenceScore(best)
	for _, r := range regions[1:] {
		score := evidenceScore(r)
		if score > bestScore {
			best, bestScore = r, score
		}
	}
	bestScore = round(bestScore)
	evidence := append([]schema.EvidenceRef(nil), best.Evidence...)
	if bestScore < minScore {
		return schema.Answer{
			Text:       "abstain: visual evidence is weak or unstable",
			Status:     "abstain",
			Confidence: bestScore,
			Evidence:   evidence,
			Warnings:   []schema.Warning{{Code: "weak-evidence", Message: "candidate region lacks enough stable boundary evidence", RegionID: best.ID}},
		}
	}
	return schema.Answer{
		Text:       "stable visual structure detected",
		Status:     "answered",
		Confidence: bestScore,
		Evidence:   evidence,
	}
}

func evidenceScore(r schema.Region) float64 {
	stability := clamp01(r.Features.ScaleStability)
	boundary := clamp01(r.Features.BoundaryStrength)
	area := clamp01(r.Features.AreaRatio * 4)
	return clamp01(0.45*stability + 0.40*boundary + 0.15*area)
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

func round(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
