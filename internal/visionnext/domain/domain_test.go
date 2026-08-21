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
	if result.Score < 0.58 || result.Confidence == "low" {
		t.Fatalf("expected strong domain evidence, got %+v", result)
	}
}

func TestClassifyEmptyFieldIsAmbiguous(t *testing.T) {
	result := Classify(evidence.Field{})
	if result.Label != Ambiguous {
		t.Fatalf("expected ambiguous empty field, got %+v", result)
	}
}
