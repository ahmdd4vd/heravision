package eval

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Manifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Samples     []Sample `json:"samples"`
}

type Sample struct {
	ID             string   `json:"id"`
	ImagePath      string   `json:"image_path"`
	AnnotationPath string   `json:"annotation_path,omitempty"`
	Split          string   `json:"split,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type RunMetadata struct {
	RunID         string    `json:"run_id"`
	GitSHA        string    `json:"git_sha,omitempty"`
	Engine        string    `json:"engine"`
	EngineVersion string    `json:"engine_version,omitempty"`
	Manifest      string    `json:"manifest"`
	StartedAt     time.Time `json:"started_at"`
	CPU           string    `json:"cpu,omitempty"`
	Threads       int       `json:"threads,omitempty"`
	Precision     string    `json:"precision,omitempty"`
}

type RunFiles struct {
	Root        string
	Metadata    string
	Manifest    string
	Predictions string
	Graphs      string
	Metrics     string
	Latency     string
	Failures    string
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("validate manifest: %w", err)
	}
	return m, nil
}

func (m Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("manifest name is empty")
	}
	if len(m.Samples) == 0 {
		return errors.New("manifest has no samples")
	}
	seen := make(map[string]struct{}, len(m.Samples))
	for i, s := range m.Samples {
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("sample %d has empty id", i)
		}
		if _, ok := seen[s.ID]; ok {
			return fmt.Errorf("duplicate sample id %q", s.ID)
		}
		seen[s.ID] = struct{}{}
		if strings.TrimSpace(s.ImagePath) == "" {
			return fmt.Errorf("sample %q has empty image_path", s.ID)
		}
	}
	return nil
}

func CreateRun(root string, meta RunMetadata, manifest Manifest) (RunFiles, error) {
	if err := manifest.Validate(); err != nil {
		return RunFiles{}, err
	}
	if strings.TrimSpace(meta.RunID) == "" {
		return RunFiles{}, errors.New("run id is empty")
	}
	if meta.StartedAt.IsZero() {
		meta.StartedAt = time.Now().UTC()
	}
	dir := filepath.Join(root, meta.RunID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return RunFiles{}, fmt.Errorf("create run directory: %w", err)
	}
	files := RunFiles{
		Root:        dir,
		Metadata:    filepath.Join(dir, "metadata.json"),
		Manifest:    filepath.Join(dir, "manifest.json"),
		Predictions: filepath.Join(dir, "predictions.jsonl"),
		Graphs:      filepath.Join(dir, "graphs.jsonl"),
		Metrics:     filepath.Join(dir, "metrics.json"),
		Latency:     filepath.Join(dir, "latency.csv"),
		Failures:    filepath.Join(dir, "failures.jsonl"),
	}
	if err := writeJSON(files.Metadata, meta); err != nil {
		return RunFiles{}, err
	}
	if err := writeJSON(files.Manifest, manifest); err != nil {
		return RunFiles{}, err
	}
	return files, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
