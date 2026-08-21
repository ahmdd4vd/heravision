package eval

import "heravision/internal/visionnext/schema"

type PairResult struct {
	SampleID string       `json:"sample_id"`
	B0       Observation  `json:"b0"`
	B1       Observation  `json:"b1"`
	Geometry MatchSummary `json:"geometry"`
	Status   string       `json:"status"`
	Error    string       `json:"error,omitempty"`
}

func EvaluatePair(sample Sample, opts RunOptions) PairResult {
	out := PairResult{SampleID: sample.ID, Status: "ok"}
	b0, err := RunB0(sample.ImagePath, opts)
	if err != nil {
		out.Status = "error"
		out.Error = err.Error()
		return out
	}
	if err := b0.Graph.Validate(); err != nil {
		out.Status = "error"
		out.Error = "b0 graph validation: " + err.Error()
		return out
	}
	b1, err := RunB1(sample.ImagePath, opts)
	if err != nil {
		out.Status = "error"
		out.Error = err.Error()
		return out
	}
	if err := b1.Graph.Validate(); err != nil {
		out.Status = "error"
		out.Error = "b1 graph validation: " + err.Error()
		return out
	}
	out.B0 = b0
	out.B1 = b1
	out.Geometry = MatchRegions(graphRegions(b0.Graph), graphRegions(b1.Graph), 0.5)
	return out
}

func graphRegions(g schema.SceneGraph) []schema.Region {
	regions := make([]schema.Region, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		regions = append(regions, n.Region)
	}
	return regions
}
