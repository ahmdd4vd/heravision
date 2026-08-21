package schema

import "testing"

func TestSceneGraphValidate(t *testing.T) {
	g := SceneGraph{
		Nodes: []Node{
			{ID: "n1", Region: Region{ID: "r1"}},
			{ID: "n2", Region: Region{ID: "r2"}},
		},
		Edges: []Relation{{
			From: "n1", To: "n2", Predicate: "above", Status: "visible", Score: 0.9,
		}},
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("valid graph rejected: %v", err)
	}
}

func TestSceneGraphValidateRejectsUnknownEdge(t *testing.T) {
	g := SceneGraph{
		Nodes: []Node{{ID: "n1", Region: Region{ID: "r1"}}},
		Edges: []Relation{{From: "n1", To: "n9", Predicate: "near", Status: "inferred"}},
	}
	if err := g.Validate(); err == nil {
		t.Fatal("graph with unknown edge target was accepted")
	}
}

func TestSceneGraphValidateRejectsInvalidStatus(t *testing.T) {
	g := SceneGraph{
		Nodes: []Node{
			{ID: "n1", Region: Region{ID: "r1"}},
			{ID: "n2", Region: Region{ID: "r2"}},
		},
		Edges: []Relation{{From: "n1", To: "n2", Predicate: "holding", Status: "guess"}},
	}
	if err := g.Validate(); err == nil {
		t.Fatal("graph with invalid relation status was accepted")
	}
}
