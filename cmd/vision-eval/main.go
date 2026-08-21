package main

import (
	"flag"
	"fmt"
	"log"

	"heravision/internal/config"
	"heravision/internal/visionnext/eval"
)

func main() {
	manifest := flag.String("manifest", "experiments/manifests/coco128-verified.json", "dataset manifest JSON")
	output := flag.String("output", "experiments/runs/latest", "output directory")
	mode := flag.String("mode", "general", "legacy/B0 mode")
	maxSide := flag.Int("max-side", 512, "B1 canonical maximum side")
	regionFilter := flag.String("region-filter", "", "optional trained B1 region filter JSON")
	regionFilterThreshold := flag.Float64("region-filter-threshold", 0.95, "trained region filter threshold")
	scaleStable := flag.Bool("scale-stable", false, "use multi-scale stable boundary-aware proposals")
	relationPrune := flag.Bool("relation-prune", false, "prune distant relation candidate pairs")
	scaleMinSupport := flag.Int("scale-min-support", 0, "minimum distinct-scale support for stable proposals; 0 uses default")
	scaleExtraFraction := flag.Float64("scale-extra-fraction", 0, "optional extra stable scale fraction above 1.0")
	answerMinScore := flag.Float64("answer-min-score", 0.65, "minimum evidence score for generic answered status")
	relationTouchingSafe := flag.Bool("relation-touching-safe", false, "suppress containment-derived touching relations")
	semanticModel := flag.String("semantic-model", "", "optional broad semantic model JSON")
	legacyMaxPixels := flag.Int64("legacy-max-pixels", 24_000_000, "explicit B0 decode pixel budget")
	flag.Parse()

	legacyConfig := config.Default()
	legacyConfig.MaxPixels = *legacyMaxPixels
	summary, err := eval.RunManifest(*manifest, *output, eval.RunOptions{
		Mode:                  *mode,
		MaxSide:               *maxSide,
		LegacyConfig:          legacyConfig,
		RegionFilterPath:      *regionFilter,
		RegionFilterThreshold: *regionFilterThreshold,
		ScaleStable:           *scaleStable,
		RelationPrune:         *relationPrune,
		ScaleMinSupport:       *scaleMinSupport,
		ScaleExtraFraction:    *scaleExtraFraction,
		AnswerMinScore:        *answerMinScore,
		RelationTouchingSafe:  *relationTouchingSafe,
		SemanticModelPath:     *semanticModel,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("manifest=%s samples=%d completed=%d errors=%d b0_regions=%d b1_regions=%d mean_coverage=%.4f mean_iou=%.4f mean_b0_ms=%.2f mean_b1_ms=%.2f\n",
		summary.Manifest, summary.Samples, summary.Completed, summary.Errors,
		summary.B0Regions, summary.B1Regions, summary.MeanCoverage, summary.MeanIoU,
		summary.MeanB0MS, summary.MeanB1MS,
	)
}
