package schema

import "fmt"

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type EvidenceRef struct {
	Kind     string  `json:"kind"`
	Stage    string  `json:"stage"`
	Scale    float64 `json:"scale,omitempty"`
	Value    float64 `json:"value,omitempty"`
	RegionID string  `json:"region_id,omitempty"`
	Geometry []Point `json:"geometry,omitempty"`
	Note     string  `json:"note,omitempty"`
}

type Features struct {
	AreaRatio        float64            `json:"area_ratio,omitempty"`
	AspectRatio      float64            `json:"aspect_ratio,omitempty"`
	Compactness      float64            `json:"compactness,omitempty"`
	Solidity         float64            `json:"solidity,omitempty"`
	BoundaryStrength float64            `json:"boundary_strength,omitempty"`
	ScaleStability   float64            `json:"scale_stability,omitempty"`
	Color            []float64          `json:"color,omitempty"`
	Texture          []float64          `json:"texture,omitempty"`
	Extra            map[string]float64 `json:"extra,omitempty"`
}

type Region struct {
	ID         string        `json:"id"`
	BBox       Rect          `json:"bbox"`
	Area       int           `json:"area"`
	PolygonRef string        `json:"polygon_ref,omitempty"`
	Features   Features      `json:"features,omitempty"`
	Evidence   []EvidenceRef `json:"evidence,omitempty"`
}

type Hypothesis struct {
	ID          string        `json:"id"`
	RegionIDs   []string      `json:"region_ids"`
	Label       string        `json:"label"`
	Score       float64       `json:"score"`
	Uncertainty float64       `json:"uncertainty"`
	Evidence    []EvidenceRef `json:"evidence,omitempty"`
	Status      string        `json:"status,omitempty"` // candidate | accepted | rejected | unknown
}

type Node struct {
	ID          string       `json:"id"`
	Region      Region       `json:"region"`
	Hypotheses  []Hypothesis `json:"hypotheses,omitempty"`
	Uncertainty float64      `json:"uncertainty"`
}

type Relation struct {
	From      string        `json:"from"`
	To        string        `json:"to"`
	Predicate string        `json:"predicate"`
	Status    string        `json:"status"` // visible | inferred
	Score     float64       `json:"score"`
	Evidence  []EvidenceRef `json:"evidence,omitempty"`
}

type Warning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	RegionID string `json:"region_id,omitempty"`
}

type Provenance struct {
	EngineVersion string `json:"engine_version"`
	SourcePath    string `json:"source_path,omitempty"`
	ImageWidth    int    `json:"image_width"`
	ImageHeight   int    `json:"image_height"`
	Mode          string `json:"mode"`
}

type SceneGraph struct {
	Nodes      []Node     `json:"nodes"`
	Edges      []Relation `json:"edges"`
	Warnings   []Warning  `json:"warnings,omitempty"`
	Provenance Provenance `json:"provenance"`
}

type Answer struct {
	Text       string        `json:"text"`
	Status     string        `json:"status"` // answered | abstain | insufficient_evidence
	Confidence float64       `json:"confidence"`
	Evidence   []EvidenceRef `json:"evidence,omitempty"`
	Warnings   []Warning     `json:"warnings,omitempty"`
}

func (g SceneGraph) Validate() error {
	seen := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.ID == "" {
			return fmt.Errorf("node id is empty")
		}
		if _, exists := seen[n.ID]; exists {
			return fmt.Errorf("duplicate node id %q", n.ID)
		}
		seen[n.ID] = struct{}{}
		if n.Region.ID == "" {
			return fmt.Errorf("node %q has empty region id", n.ID)
		}
	}
	for _, e := range g.Edges {
		if _, ok := seen[e.From]; !ok {
			return fmt.Errorf("edge references unknown from node %q", e.From)
		}
		if _, ok := seen[e.To]; !ok {
			return fmt.Errorf("edge references unknown to node %q", e.To)
		}
		if e.Status != "visible" && e.Status != "inferred" {
			return fmt.Errorf("edge %s->%s has invalid status %q", e.From, e.To, e.Status)
		}
	}
	return nil
}
