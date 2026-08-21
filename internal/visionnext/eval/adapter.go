package eval

import (
	"fmt"
	"time"

	"heravision/internal/config"
	"heravision/internal/facts"
	"heravision/internal/processor"
	"heravision/internal/visionnext/evidence"
	"heravision/internal/visionnext/graph"
	"heravision/internal/visionnext/hypothesis"
	"heravision/internal/visionnext/imageview"
	"heravision/internal/visionnext/learned"
	"heravision/internal/visionnext/region"
	"heravision/internal/visionnext/relation"
	"heravision/internal/visionnext/schema"
)

type RunOptions struct {
	Mode                  string
	EngineVersion         string
	MaxSide               int
	LegacyConfig          config.Config
	RegionFilterPath      string
	RegionFilterThreshold float64
	ScaleStable           bool
}

type Observation struct {
	Engine      string            `json:"engine"`
	ImagePath   string            `json:"image_path"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	ElapsedMS   int64             `json:"elapsed_ms"`
	Graph       schema.SceneGraph `json:"graph"`
	LegacyBoxes int               `json:"legacy_boxes,omitempty"`
}

func RunB0(path string, opts RunOptions) (Observation, error) {
	started := time.Now()
	version := opts.EngineVersion
	if version == "" {
		version = "b0-legacy"
	}
	mode := opts.Mode
	if mode == "" {
		mode = "general"
	}
	cfg := opts.LegacyConfig
	if cfg.MaxSide == 0 {
		cfg = config.Default()
	}
	previousMaxPixels := processor.MaxPixels
	if cfg.MaxPixels > 0 {
		processor.MaxPixels = cfg.MaxPixels
	}
	defer func() { processor.MaxPixels = previousMaxPixels }()
	result, err := facts.Extract(path, mode, version, cfg)
	if err != nil {
		return Observation{}, fmt.Errorf("b0 extract: %w", err)
	}
	regions := make([]schema.Region, 0, len(result.Boxes))
	hyps := make([]schema.Hypothesis, 0, len(result.Boxes))
	for i, b := range result.Boxes {
		id := fmt.Sprintf("b0-r-%04d", i+1)
		regions = append(regions, schema.Region{
			ID:   id,
			BBox: schema.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H},
			Area: b.W * b.H,
			Features: schema.Features{
				AreaRatio:      safeRatio(b.W*b.H, result.Meta.Width*result.Meta.Height),
				AspectRatio:    safeRatio(b.W, b.H),
				ScaleStability: 0,
				Extra:          map[string]float64{"legacy_score": b.Score},
			},
			Evidence: []schema.EvidenceRef{{Kind: "legacy-box", Stage: "b0-detector", RegionID: id, Value: b.Score, Note: b.Type}},
		})
		hyps = append(hyps, schema.Hypothesis{
			ID:          fmt.Sprintf("b0-h-%04d", i+1),
			RegionIDs:   []string{id},
			Label:       b.Type,
			Score:       clamp01(b.Score),
			Uncertainty: 1 - clamp01(b.Score),
			Status:      "candidate",
			Evidence:    []schema.EvidenceRef{{Kind: "legacy-classification", Stage: "b0-detector", RegionID: id, Value: b.Score}},
		})
	}
	edges := relation.Build(regions)
	g := graph.Build(regions, hyps, edges, schema.Provenance{
		EngineVersion: version, SourcePath: path, ImageWidth: result.Meta.Width, ImageHeight: result.Meta.Height, Mode: mode,
	})
	return Observation{Engine: "B0", ImagePath: path, Width: result.Meta.Width, Height: result.Meta.Height, ElapsedMS: time.Since(started).Milliseconds(), Graph: g, LegacyBoxes: len(result.Boxes)}, nil
}

func RunB1(path string, opts RunOptions) (Observation, error) {
	started := time.Now()
	version := opts.EngineVersion
	if version == "" {
		version = "b1-dev"
	}
	maxSide := opts.MaxSide
	if maxSide <= 0 {
		maxSide = 1024
	}
	previousMaxPixels := processor.MaxPixels
	if opts.LegacyConfig.MaxPixels > 0 {
		processor.MaxPixels = opts.LegacyConfig.MaxPixels
	}
	defer func() { processor.MaxPixels = previousMaxPixels }()
	img, _, err := processor.Decode(path)
	if err != nil {
		return Observation{}, fmt.Errorf("b1 decode: %w", err)
	}
	view, err := imageview.FromImage(img, maxSide)
	if err != nil {
		return Observation{}, fmt.Errorf("b1 canonical view: %w", err)
	}
	field := evidence.Compute(view)
	var regions []schema.Region
	if opts.ScaleStable {
		regions, err = region.ProposeStable(img, maxSide, region.StableConfig{Base: region.DefaultConfig()})
		if err != nil {
			return Observation{}, fmt.Errorf("b1 stable proposals: %w", err)
		}
	} else {
		regions = region.Propose(field, region.DefaultConfig())
	}
	if opts.RegionFilterPath != "" {
		filter, filterErr := learned.LoadRegionFilter(opts.RegionFilterPath)
		if filterErr != nil {
			return Observation{}, filterErr
		}
		regions = filter.Apply(regions, view.Width, view.Height, opts.RegionFilterThreshold)
	}
	hyps := hypothesis.Generate(regions, view.Width, view.Height)
	edges := relation.Build(regions)
	g := graph.Build(regions, hyps, edges, schema.Provenance{
		EngineVersion: version, SourcePath: path, ImageWidth: view.Width, ImageHeight: view.Height, Mode: opts.Mode,
	})
	return Observation{Engine: "B1", ImagePath: path, Width: view.Width, Height: view.Height, ElapsedMS: time.Since(started).Milliseconds(), Graph: g}, nil
}

func safeRatio(a, b int) float64 {
	if b <= 0 {
		return 0
	}
	return float64(a) / float64(b)
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
