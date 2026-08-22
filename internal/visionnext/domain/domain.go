package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/schema"
)

type Label string

const (
	NaturalPhoto       Label = "natural_photo"
	DiagramDocument    Label = "diagram_document"
	ScreenshotDocument Label = "screenshot_document"
	Ambiguous          Label = "ambiguous"
)

type Features struct {
	FlatnessMean             float64 `json:"flatness_mean"`
	EdgeDensity              float64 `json:"edge_density"`
	ContrastMean             float64 `json:"contrast_mean"`
	ChromaMean               float64 `json:"chroma_mean"`
	ChromaStd                float64 `json:"chroma_std"`
	OrientationEntropy       float64 `json:"orientation_entropy"`
	LumaStd                  float64 `json:"luma_std"`
	AspectRatio              float64 `json:"aspect_ratio"`
	BlankFraction            float64 `json:"blank_fraction"`
	LowInfoFraction          float64 `json:"low_info_fraction"`
	EdgeHighFraction         float64 `json:"edge_high_fraction"`
	OrientationConcentration float64 `json:"orientation_concentration"`
	AxisConcentration        float64 `json:"axis_concentration"`
	LumaRange                float64 `json:"luma_range"`
	LineStructure            float64 `json:"line_structure"`
}

var FeatureNames = []string{"flatness_mean", "edge_density", "contrast_mean", "chroma_mean", "chroma_std", "orientation_entropy", "luma_std", "aspect_ratio", "blank_fraction", "low_info_fraction", "edge_high_fraction", "orientation_concentration", "axis_concentration", "luma_range", "line_structure"}

type Model struct {
	Name         string               `json:"name"`
	FeatureNames []string             `json:"features"`
	Labels       []string             `json:"labels"`
	Weights      map[string][]float64 `json:"weights"`
	Bias         map[string]float64   `json:"bias"`
	MinScore     float64              `json:"min_score"`
	MinMargin    float64              `json:"min_margin"`
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
func Load(path string) (Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return Model{}, err
	}
	if len(model.Labels) == 0 || len(model.FeatureNames) == 0 {
		return Model{}, fmt.Errorf("domain model has no labels or features")
	}
	if model.MinScore <= 0 {
		model.MinScore = 0.65
	}
	if model.MinMargin <= 0 {
		model.MinMargin = 0.10
	}
	return model, nil
}

func FeatureVector(field evidence.Field) []float64 {
	f := summarize(field)
	return []float64{f.FlatnessMean, f.EdgeDensity, f.ContrastMean, f.ChromaMean, f.ChromaStd, f.OrientationEntropy, f.LumaStd, f.AspectRatio, f.BlankFraction, f.LowInfoFraction, f.EdgeHighFraction, f.OrientationConcentration, f.AxisConcentration, f.LumaRange, f.LineStructure}
}

func ClassifyWithModel(field evidence.Field, model Model) Result {
	features := FeatureVector(field)
	type scored struct {
		label string
		score float64
	}
	scores := make([]scored, 0, len(model.Labels))
	for _, label := range model.Labels {
		w := model.Weights[label]
		if len(w) != len(features) {
			continue
		}
		z := model.Bias[label]
		for i, v := range features {
			z += w[i] * v
		}
		scores = append(scores, scored{label: label, score: sigmoid(z)})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if len(scores) < 2 {
		return Classify(field)
	}
	margin := scores[0].score - scores[1].score
	label := Ambiguous
	score := scores[0].score
	confidence := "low"
	if score >= model.MinScore && margin >= model.MinMargin {
		label = Label(scores[0].label)
		confidence = "medium"
		if score >= 0.80 && margin >= 0.20 {
			confidence = "high"
		}
	}
	action := "abstain_domain"
	if label == NaturalPhoto || (label == DiagramDocument && confidence != "high") {
		action = "allow_object_semantic"
	}
	if (label == DiagramDocument || label == ScreenshotDocument) && confidence == "high" {
		action = "block_object_semantic"
	}
	evidenceRefs := []schema.EvidenceRef{{Kind: "domain-calibrated-score", Stage: "domain-gate", Value: round(score), Note: fmt.Sprintf("label=%s margin=%.3f action=%s", label, margin, action)}}
	return Result{Label: label, Score: round(score), Margin: round(margin), Features: summarize(field), Evidence: evidenceRefs, Confidence: confidence, Action: action}
}

func sigmoid(v float64) float64 {
	if v >= 0 {
		z := math.Exp(-v)
		return 1 / (1 + z)
	}
	z := math.Exp(v)
	return z / (1 + z)
}

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
	if (label == DiagramDocument || label == ScreenshotDocument) && confidence == "high" {
		action = "block_object_semantic"
	}
	evidenceRefs := []schema.EvidenceRef{
		{Kind: "domain-statistics", Stage: "domain-gate", Value: round(features.FlatnessMean), Note: fmt.Sprintf("flatness_mean=%.3f edge_density=%.3f contrast_mean=%.3f chroma_mean=%.3f chroma_std=%.3f orientation_entropy=%.3f luma_std=%.3f aspect_ratio=%.3f blank_fraction=%.3f low_info_fraction=%.3f edge_high_fraction=%.3f orientation_concentration=%.3f axis_concentration=%.3f luma_range=%.3f line_structure=%.3f", features.FlatnessMean, features.EdgeDensity, features.ContrastMean, features.ChromaMean, features.ChromaStd, features.OrientationEntropy, features.LumaStd, features.AspectRatio, features.BlankFraction, features.LowInfoFraction, features.EdgeHighFraction, features.OrientationConcentration, features.AxisConcentration, features.LumaRange, features.LineStructure)},
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
	var blank, lowInfo, edgeHigh float64
	bins := make([]float64, 8)
	orientTotal := 0.0
	lumaMin, lumaMax := 1.0, 0.0
	for i := 0; i < n; i++ {
		l := value(field.Luminance, i)
		lum += l
		lum2 += l * l
		if l < lumaMin {
			lumaMin = l
		}
		if l > lumaMax {
			lumaMax = l
		}
		flatness := value(field.Flatness, i)
		e := value(field.Edge, i)
		flat += flatness
		edge += e
		contrast += value(field.LocalContrast, i)
		c := value(field.ChromaMagnitude, i)
		chroma += c
		chroma2 += c * c
		if l <= 0.04 || l >= 0.96 {
			blank++
		}
		if flatness >= 0.85 && e <= 0.08 {
			lowInfo++
		}
		if e >= 0.20 {
			edgeHigh++
		}
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
	entropy, orientationConcentration, axisConcentration := 0.0, 0.0, 0.0
	if orientTotal > 0 {
		maxBin := 0.0
		for _, b := range bins {
			if b > maxBin {
				maxBin = b
			}
			if b > 0 {
				p := b / orientTotal
				entropy -= p * math.Log(p) / math.Log(8)
			}
		}
		orientationConcentration = maxBin / orientTotal
		for i := 0; i < 4; i++ {
			pair := (bins[i] + bins[i+4]) / orientTotal
			if pair > axisConcentration {
				axisConcentration = pair
			}
		}
	}
	aspect := 0.0
	if field.Width > 0 && field.Height > 0 {
		aspect = float64(field.Width) / float64(field.Height)
		if aspect > 1 {
			aspect = 1 / aspect
		}
	}
	flatMean := clamp01(flat / float64(n))
	edgeMean := clamp01(edge / float64(n))
	edgeHighMean := clamp01(edgeHigh / float64(n))
	return Features{
		FlatnessMean: flatMean, EdgeDensity: edgeMean, ContrastMean: clamp01(contrast * 4 / float64(n)),
		ChromaMean: clamp01(cmean), ChromaStd: clamp01(math.Sqrt(cvar)), OrientationEntropy: clamp01(entropy), LumaStd: clamp01(math.Sqrt(lvar)),
		AspectRatio: clamp01(aspect), BlankFraction: clamp01(blank / float64(n)), LowInfoFraction: clamp01(lowInfo / float64(n)),
		EdgeHighFraction: edgeHighMean, OrientationConcentration: clamp01(orientationConcentration), AxisConcentration: clamp01(axisConcentration),
		LumaRange: clamp01(lumaMax - lumaMin), LineStructure: clamp01(2 * edgeHighMean * orientationConcentration),
	}
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
