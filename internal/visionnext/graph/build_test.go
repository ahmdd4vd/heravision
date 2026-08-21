package graph

import (
	"testing"

	"heravision/internal/visionnext/schema"
)

func TestBuildSceneGraph(t *testing.T) {
	regions := []schema.Region{{ID: "r1", BBox: schema.Rect{W: 4, H: 4}, Area: 16}}
	hypotheses := []schema.Hypothesis{{ID: "r1-h0", RegionIDs: []string{"r1"}, Label: "region", Score: 0.7}}
	g := Build(regions, hypotheses, nil, schema.Provenance{EngineVersion: "next-dev", Mode: "pure"})
	if len(g.Nodes) != 1 || len(g.Nodes[0].Hypotheses) != 1 {
		t.Fatalf("hypothesis was not attached: %+v", g)
	}
	if g.Provenance.EngineVersion != "next-dev" {
		t.Fatalf("provenance lost: %+v", g.Provenance)
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("graph should validate: %v", err)
	}
}
