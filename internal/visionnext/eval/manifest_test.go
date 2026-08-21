package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestValidate(t *testing.T) {
	m := Manifest{
		Name:    "smoke",
		Samples: []Sample{{ID: "a", ImagePath: "testdata/a.png", Split: "test"}},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
}

func TestManifestRejectsDuplicateIDs(t *testing.T) {
	m := Manifest{
		Name: "bad",
		Samples: []Sample{
			{ID: "a", ImagePath: "a.png"},
			{ID: "a", ImagePath: "b.png"},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("duplicate sample ids were accepted")
	}
}

func TestCreateRunWritesMetadataAndManifest(t *testing.T) {
	root := t.TempDir()
	manifest := Manifest{Name: "smoke", Samples: []Sample{{ID: "a", ImagePath: "a.png"}}}
	files, err := CreateRun(root, RunMetadata{
		RunID: "run-001", Engine: "b1", StartedAt: time.Unix(1, 0).UTC(), Threads: 2,
	}, manifest)
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	for _, path := range []string{files.Metadata, files.Manifest} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
	data, err := os.ReadFile(files.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("manifest artifact is invalid JSON: %v", err)
	}
	if got.Name != "smoke" || len(got.Samples) != 1 {
		t.Fatalf("unexpected manifest artifact: %+v", got)
	}
	if filepath.Base(files.Root) != "run-001" {
		t.Fatalf("unexpected run root: %s", files.Root)
	}
}
