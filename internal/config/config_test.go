package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultNormalized(t *testing.T) {
	c := Default()
	c.normalize()
	if c.MaxSide != 1024 || c.Detector.CannyLow != 50 || c.Detector.CannyHigh != 150 {
		t.Fatalf("bad defaults: %+v", c)
	}
}

func TestNormalizeClamps(t *testing.T) {
	c := Config{MaxSide: 1, MaxPixels: 0, Detector: Detector{MinArea: 0}, Color: Color{K: 999}, Preprocess: Preprocess{BlurThreshold: -5}}
	c.normalize()
	if c.MaxSide != 1024 || c.MaxPixels != 12_000_000 || c.Detector.MinArea != 200 {
		t.Fatalf("clamp failed: %+v", c)
	}
	if c.Color.K != 16 || c.Preprocess.BlurThreshold != 80 {
		t.Fatalf("clamp failed: %+v", c)
	}
}

func TestLoadExplicitMissingErrors(t *testing.T) {
	if _, _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for explicit missing config")
	}
}

func TestLoadFromFileOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "heravision.json")
	content := []byte(`{"max_side": 512, "detector": {"canny_low": 30}}`)
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := Load(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if path != p {
		t.Fatalf("expected path %s got %s", p, path)
	}
	if cfg.MaxSide != 512 {
		t.Fatalf("max_side override failed: %d", cfg.MaxSide)
	}
	if cfg.Detector.CannyLow != 30 {
		t.Fatalf("canny_low override failed: %d", cfg.Detector.CannyLow)
	}
	if cfg.Detector.CannyHigh != 150 {
		t.Fatalf("unset field must keep default, got %d", cfg.Detector.CannyHigh)
	}
}

func TestLoadNoFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := Load("")
	if err != nil || path != "" {
		t.Fatalf("expected defaults without error, got path=%q err=%v", path, err)
	}
	if cfg.MaxSide != 1024 {
		t.Fatal("expected default max_side")
	}
}
