package domain

import (
	"testing"

	"heravision/internal/visionnext/evidence"
)

func TestClassifyFlatNeutralAsDiagramDocument(t *testing.T) {
	field := evidence.Field{Luminance: make([]float64, 64), Edge: make([]float64, 64), LocalContrast: make([]float64, 64), Flatness: make([]float64, 64), ChromaMagnitude: make([]float64, 64), Orientation: make([]float64, 64)}
	for i := range field.Flatness {
		field.Flatness[i] = 1
	}
	result := Classify(field)
	if result.Label != DiagramDocument {
		t.Fatalf("expected diagram_document, got %+v", result)
	}
	if result.Score < 0.58 || result.Confidence == "low" || result.Action != "block_object_semantic" {
		t.Fatalf("expected strong blocking domain evidence, got %+v", result)
	}
}

func TestClassifyEmptyFieldIsAmbiguous(t *testing.T) {
	result := Classify(evidence.Field{})
	if result.Label != Ambiguous || result.Action != "abstain_domain" {
		t.Fatalf("expected ambiguous abstaining domain result, got %+v", result)
	}
}

func TestFeatureVectorIncludesImageLevelSignals(t *testing.T) {
	field := evidence.Field{
		Width: 8, Height: 8,
		Luminance: make([]float64, 64), Edge: make([]float64, 64), LocalContrast: make([]float64, 64), Flatness: make([]float64, 64), ChromaMagnitude: make([]float64, 64), Orientation: make([]float64, 64),
	}
	for i := range field.Flatness {
		field.Flatness[i] = 1
	}
	features := FeatureVector(field)
	if len(features) != len(FeatureNames) || len(features) != 15 {
		t.Fatalf("expected 15 aligned features, got %d names/%d values", len(FeatureNames), len(features))
	}
	if features[8] != 1 || features[9] != 1 {
		t.Fatalf("expected blank and low-info fractions to be high, got blank=%v low_info=%v", features[8], features[9])
	}
}
