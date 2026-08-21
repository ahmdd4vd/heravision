package graph

import (
	"heravision/internal/visionnext/schema"
)

func Build(regions []schema.Region, hypotheses []schema.Hypothesis, edges []schema.Relation, provenance schema.Provenance) schema.SceneGraph {
	hByRegion := make(map[string][]schema.Hypothesis)
	for _, h := range hypotheses {
		for _, id := range h.RegionIDs {
			hByRegion[id] = append(hByRegion[id], h)
		}
	}
	nodes := make([]schema.Node, 0, len(regions))
	for _, r := range regions {
		hs := append([]schema.Hypothesis(nil), hByRegion[r.ID]...)
		uncertainty := 1.0
		for _, h := range hs {
			if h.Score > 1-uncertainty {
				uncertainty = 1 - h.Score
			}
		}
		nodes = append(nodes, schema.Node{ID: r.ID, Region: r, Hypotheses: hs, Uncertainty: uncertainty})
	}
	g := schema.SceneGraph{Nodes: nodes, Edges: edges, Provenance: provenance}
	if err := g.Validate(); err != nil {
		g.Warnings = append(g.Warnings, schema.Warning{Code: "invalid_graph", Message: err.Error()})
	}
	return g
}
