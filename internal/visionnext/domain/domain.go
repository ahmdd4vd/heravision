package domain

import (
	"fmt"
	"math"
	"sort"

	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/schema"
)

type Label string

const (
	NaturalPhoto    Label = "natural_photo"
	DiagramDocument Label = "diagram_document"
	Ambiguous       Label = "ambiguous"
)

type Features struct {
	FlatnessMean       float64 `json:"flatness_mean"`
	EdgeDensity        float64 `json:"edge_density"`
	ContrastMean       float64 `json:"contrast_mean"`
	ChromaMean         float64 `json:"chroma_mean"`
	ChromaStd          float64 `json:"chroma_std"`
	OrientationEntropy float64 `json:"orientation_entropy"`
	LumaStd            float64 `json:"luma_std"`
}

type Result struct {
	Label      Label
	Score      float64
	Margin     float64
	Features   Features
	Evidence   []schema.EvidenceRef
	Confidence string
	Action     string // allow_object_semantic | block_object_semantic | abstain_domain
}

// Classify uses bounded evidence-field statistics. It is deliberately a gate,
// not a semantic recognizer: ambiguous inputs are returned as ambiguous.
func Classify(field evidence.Field) Result {
	features := summarize(field)
	diagram := clamp01(0.38*features.FlatnessMean + 0.22*features.EdgeDensity + 0.18*(1-features.ChromaMean) + 0.12*(1-features.ChromaStd) + 0.10*features.OrientationEntropy)
	natural := clamp01(0.34*(1-features.FlatnessMean) + 0.24*features.ContrastMean + 0.20*features.ChromaStd + 0.12*features.ChromaMean + 0.10*(1-features.OrientationEntropy))
	scores := []struct {
		label Label
		score float64
	}{{NaturalPhoto, natural}, {DiagramDocument, diagram}}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	margin := scores[0].score - scores[1].score
	label := Ambiguous
	score := scores[0].score
	confidence := "low"
	if score >= 0.58 && margin >= 0.10 {
		label = scores[0].label
		confidence = "medium"
		if score >= 0.65 && margin >= 0.18 {
			confidence = "high"
		}
	}
	action := "abstain_domain"
	if label == NaturalPhoto || (label == DiagramDocument && confidence != "high") {
		action = "allow_object_semantic"
	}
	if label == DiagramDocument && confidence == "high" {
		action = "block_object_semantic"
	}
	evidenceRefs := []schema.EvidenceRef{
		{Kind: "domain-statistics", Stage: "domain-gate", Value: round(features.FlatnessMean), Note: fmt.Sprintf("flatness_mean=%.3f edge_density=%.3f contrast_mean=%.3f", features.FlatnessMean, features.EdgeDensity, features.ContrastMean)},
		{Kind: "domain-score", Stage: "domain-gate", Value: round(score), Note: fmt.Sprintf("natural=%.3f diagram_document=%.3f margin=%.3f action=%s", natural, diagram, margin, action)},
	}
	return Result{Label: label, Score: round(score), Margin: round(margin), Features: features, Evidence: evidenceRefs, Confidence: confidence, Action: action}
}

func summarize(field evidence.Field) Features {
	n := len(field.Luminance)
	if n == 0 {
		return Features{}
	}
	var lum, lum2, flat, edge, contrast, chroma, chroma2 float64
	bins := make([]float64, 8)
	orientTotal := 0.0
	for i := 0; i < n; i++ {
		lum += field.Luminance[i]
		lum2 += field.Luminance[i] * field.Luminance[i]
		flat += value(field.Flatness, i)
		edge += value(field.Edge, i)
		contrast += value(field.LocalContrast, i)
		c := value(field.ChromaMagnitude, i)
		chroma += c
		chroma2 += c * c
		e := value(field.Edge, i)
		if e > 1e-6 {
			angle := value(field.Orientation, i)
			bin := int(math.Floor((angle + math.Pi) / (2 * math.Pi) * 8))
			if bin < 0 {
				bin = 0
			}
			if bin >= 8 {
				bin = 7
			}
			bins[bin] += e
			orientTotal += e
		}
	}
	mean := lum / float64(n)
	lvar := lum2/float64(n) - mean*mean
	if lvar < 0 {
		lvar = 0
	}
	cmean := chroma / float64(n)
	cvar := chroma2/float64(n) - cmean*cmean
	if cvar < 0 {
		cvar = 0
	}
	entropy := 0.0
	if orientTotal > 0 {
		for _, b := range bins {
			if b > 0 {
				p := b / orientTotal
				entropy -= p * math.Log(p) / math.Log(8)
			}
		}
	}
	return Features{FlatnessMean: clamp01(flat / float64(n)), EdgeDensity: clamp01(edge / float64(n)), ContrastMean: clamp01(contrast * 4 / float64(n)), ChromaMean: clamp01(cmean), ChromaStd: clamp01(math.Sqrt(cvar)), OrientationEntropy: clamp01(entropy), LumaStd: clamp01(math.Sqrt(lvar))}
}

func Hypothesis(regions []schema.Region, result Result) schema.Hypothesis {
	ids := make([]string, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}
	status := "candidate"
	if result.Confidence == "high" {
		status = "accepted"
	}
	if result.Label == Ambiguous || result.Action == "abstain_domain" {
		status = "unknown"
	}
	return schema.Hypothesis{ID: "image-domain-" + string(result.Label), RegionIDs: ids, Label: string(result.Label), Score: result.Score, Uncertainty: round(1 - result.Score), Status: status, Evidence: result.Evidence}
}

func value(v []float64, i int) float64 {
	if i < 0 || i >= len(v) {
		return 0
	}
	return v[i]
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
func round(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
