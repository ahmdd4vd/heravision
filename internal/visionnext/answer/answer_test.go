package answer

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestFromRegionsInsufficientEvidence(t *testing.T) {
	answer := FromRegions(nil)
	if answer.Status != "insufficient_evidence" || answer.Confidence != 0 {
		t.Fatalf("unexpected insufficient answer: %+v", answer)
	}
}

func TestFromRegionsAbstainsOnWeakEvidence(t *testing.T) {
	answer := FromRegions([]schema.Region{{
		ID:       "r-1",
		Area:     10,
		Features: schema.Features{AreaRatio: 0.01, BoundaryStrength: 0.1, ScaleStability: 0.1},
	}})
	if answer.Status != "abstain" {
		t.Fatalf("expected abstain, got %+v", answer)
	}
}

func TestFromRegionsUsesOnlyGenericClaim(t *testing.T) {
	answer := FromRegions([]schema.Region{{
		ID:       "r-1",
		Area:     400,
		Features: schema.Features{AreaRatio: 0.25, BoundaryStrength: 0.9, ScaleStability: 1},
		Evidence: []schema.EvidenceRef{{Kind: "scale-consensus", Stage: "test", RegionID: "r-1", Value: 1}},
	}})
	if answer.Status != "answered" {
		t.Fatalf("expected generic answer, got %+v", answer)
	}
	if answer.Text != "stable visual structure detected" {
		t.Fatalf("unexpected semantic claim: %q", answer.Text)
	}
	if len(answer.Evidence) == 0 {
		t.Fatal("expected provenance evidence")
	}
}
