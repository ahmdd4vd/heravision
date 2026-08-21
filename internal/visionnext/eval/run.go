package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DatasetSummary struct {
	Manifest     string    `json:"manifest"`
	Samples      int       `json:"samples"`
	Completed    int       `json:"completed"`
	Errors       int       `json:"errors"`
	B0Regions    int       `json:"b0_regions"`
	B1Regions    int       `json:"b1_regions"`
	MeanCoverage float64   `json:"mean_coverage_b0_to_b1"`
	MeanIoU      float64   `json:"mean_iou_matched"`
	MeanB0MS     float64   `json:"mean_b0_ms"`
	MeanB1MS     float64   `json:"mean_b1_ms"`
	GeneratedAt  time.Time `json:"generated_at"`
}

func RunManifest(manifestPath, outputDir string, opts RunOptions) (DatasetSummary, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return DatasetSummary{}, err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return DatasetSummary{}, fmt.Errorf("create output directory: %w", err)
	}
	predictionsPath := filepath.Join(outputDir, "predictions.jsonl")
	file, err := os.Create(predictionsPath)
	if err != nil {
		return DatasetSummary{}, fmt.Errorf("create predictions: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	summary := DatasetSummary{Manifest: manifestPath, Samples: len(manifest.Samples), GeneratedAt: time.Now().UTC()}
	var coverageSum, iouSum float64
	var matchedCount int
	for _, sample := range manifest.Samples {
		result := EvaluatePair(sample, opts)
		if err := encoder.Encode(result); err != nil {
			return DatasetSummary{}, fmt.Errorf("write prediction %s: %w", sample.ID, err)
		}
		if result.Status != "ok" {
			summary.Errors++
			continue
		}
		summary.Completed++
		summary.B0Regions += len(result.B0.Graph.Nodes)
		summary.B1Regions += len(result.B1.Graph.Nodes)
		summary.MeanB0MS += float64(result.B0.ElapsedMS)
		summary.MeanB1MS += float64(result.B1.ElapsedMS)
		coverageSum += result.Geometry.CoverageA
		for _, match := range result.Geometry.Matches {
			iouSum += match.IoU
			matchedCount++
		}
	}
	if err := writer.Flush(); err != nil {
		return DatasetSummary{}, fmt.Errorf("flush predictions: %w", err)
	}
	if summary.Completed > 0 {
		summary.MeanCoverage = coverageSum / float64(summary.Completed)
		summary.MeanB0MS /= float64(summary.Completed)
		summary.MeanB1MS /= float64(summary.Completed)
	}
	if matchedCount > 0 {
		summary.MeanIoU = iouSum / float64(matchedCount)
	}
	if err := writeJSON(filepath.Join(outputDir, "summary.json"), summary); err != nil {
		return DatasetSummary{}, err
	}
	return summary, nil
}
