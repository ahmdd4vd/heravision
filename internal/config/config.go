package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Detector struct {
	CannyLow  uint8 `json:"canny_low"`
	CannyHigh uint8 `json:"canny_high"`
	MinArea   int   `json:"min_area"`
}

type Color struct {
	K           int     `json:"k"`
	DeltaEMerge float64 `json:"deltaE_merge"`
}

type Preprocess struct {
	BlurThreshold float64 `json:"blur_threshold"`
}

type Config struct {
	MaxSide    int        `json:"max_side"`
	MaxPixels  int64      `json:"max_pixels"`
	Multiscale bool       `json:"multiscale"`
	Detector   Detector   `json:"detector"`
	Color      Color      `json:"color"`
	Preprocess Preprocess `json:"preprocess"`
}

func Default() Config {
	return Config{
		MaxSide:    1024,
		MaxPixels:  12_000_000,
		Multiscale: true,
		Detector:   Detector{CannyLow: 50, CannyHigh: 150, MinArea: 200},
		Color:      Color{K: 5, DeltaEMerge: 12},
		Preprocess: Preprocess{
			BlurThreshold: 80,
		},
	}
}

func Load(explicit string) (Config, string, error) {
	candidates := []string{}
	if explicit != "" {
		candidates = append(candidates, explicit)
	} else {
		candidates = append(candidates, "heravision.json")
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, "heravision.json"))
		}
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && explicit == "" {
				continue
			}
			if explicit != "" {
				return Default(), "", fmt.Errorf("read config %s: %w", p, err)
			}
			continue
		}
		cfg := Default()
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Default(), p, fmt.Errorf("parse config %s: %w", p, err)
		}
		cfg.normalize()
		return cfg, p, nil
	}
	cfg := Default()
	cfg.normalize()
	return cfg, "", nil
}

func (c *Config) normalize() {
	if c.MaxSide < 64 {
		c.MaxSide = 1024
	}
	if c.MaxSide > 4096 {
		c.MaxSide = 4096
	}
	if c.MaxPixels < 100_000 {
		c.MaxPixels = 12_000_000
	}
	if c.Detector.MinArea < 25 {
		c.Detector.MinArea = 200
	}
	if c.Detector.CannyLow == 0 {
		c.Detector.CannyLow = 50
	}
	if c.Detector.CannyHigh <= c.Detector.CannyLow {
		c.Detector.CannyHigh = c.Detector.CannyLow * 3
	}
	if c.Color.K < 1 {
		c.Color.K = 5
	}
	if c.Color.K > 16 {
		c.Color.K = 16
	}
	if c.Color.DeltaEMerge <= 0 {
		c.Color.DeltaEMerge = 12
	}
	if c.Preprocess.BlurThreshold <= 0 {
		c.Preprocess.BlurThreshold = 80
	}
}
