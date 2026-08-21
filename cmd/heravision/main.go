package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"heravision/internal/buildinfo"
	"heravision/internal/config"
	"heravision/internal/facts"
	"heravision/mcp"
)

func main() {
	root := &cobra.Command{
		Use:   "heravision",
		Short: "UI structure extractor for blind LLMs — pure native vision",
		Long:  "HeraVision extracts UI structure facts from images (element boxes, colors, layout) for text-only LLMs. Pure Go, offline, no API. Text content is not OCR-read.",
	}
	root.AddCommand(versionCmd(), extractCmd(), compareCmd(), mcpCmd(), doctorCmd(), setupCmd(), benchCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Printf("heravision %s\n", buildinfo.Version) },
	}
}

func extractCmd() *cobra.Command {
	var mode string
	var asJSON bool
	var cfgPath string
	c := &cobra.Command{
		Use:   "extract <image>",
		Short: "Extract structured structure facts from image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgFile, err := config.Load(cfgPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[warn] %v — using defaults\n", err)
			}
			r, err := facts.Extract(args[0], mode, buildinfo.Version, cfg)
			if err != nil {
				return err
			}
			if cfgFile != "" {
				fmt.Fprintf(os.Stderr, "[info] config: %s\n", cfgFile)
			}
			j, _ := json.MarshalIndent(r, "", "  ")
			if asJSON {
				fmt.Println(string(j))
				return nil
			}
			fmt.Printf("%s\n```json\n%s\n```\n", r.Markdown, string(j))
			return nil
		},
	}
	c.Flags().StringVar(&mode, "mode", "general", "Mode: general|ui|code|diagram|error|blur")
	c.Flags().BoolVar(&asJSON, "json", false, "Output JSON only")
	c.Flags().StringVar(&cfgPath, "config", "", "Path to heravision.json config")
	return c
}

func compareCmd() *cobra.Command {
	var cfgPath string
	c := &cobra.Command{
		Use:   "compare <a> <b>",
		Short: "Compare two images and return structural diff",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.Load(cfgPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[warn] %v — using defaults\n", err)
			}
			res, err := facts.Compare(args[0], args[1], cfg)
			if err != nil {
				return err
			}
			j, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(j))
			return nil
		},
	}
	c.Flags().StringVar(&cfgPath, "config", "", "Path to heravision.json config")
	return c
}

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run MCP stdio server",
		RunE:  func(cmd *cobra.Command, args []string) error { return mcp.Serve() },
	}
}

func doctorCmd() *cobra.Command {
	var cfgPath string
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Check heravision setup",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("heravision %s\n", buildinfo.Version)
			exe, _ := os.Executable()
			if exe != "" {
				fmt.Printf("[ok] exe: %s\n", exe)
			}
			cfg, cfgFile, err := config.Load(cfgPath)
			if err != nil {
				fmt.Printf("[warn] config: %v (defaults in use)\n", err)
			} else if cfgFile != "" {
				fmt.Printf("[ok] config: %s (max_side=%d, canny=%d/%d)\n", cfgFile, cfg.MaxSide, cfg.Detector.CannyLow, cfg.Detector.CannyHigh)
			} else {
				fmt.Println("[info] config: none found — defaults in use (see heravision.json.example)")
			}
			if fi, err := os.Stat("testdata"); err == nil && fi.IsDir() {
				fmt.Println("[ok] testdata: present")
			} else {
				fmt.Println("[warn] testdata: missing")
			}
			fmt.Println("[ok] processor: decode jpg/png/webp + EXIF auto-rotate + decode limits")
			fmt.Println("[ok] detector: Sobel+Canny hysteresis+NMS, morph close, 8-connect, classify v3, box color")
			fmt.Println("[ok] color: Lab k-means + dE merge + background border")
			fmt.Println("[ok] facts: page_type classifier, grid detection, captions, reading order, diff summary")
			fmt.Println("[ok] diagram: mermaid chain graph (mode diagram)")
			fmt.Println("[warn] ocr: heuristic shape placeholders only — real OCR engine not bundled yet (roadmap)")
			fmt.Printf("[info] targets: binary <12MB core, RAM <80MB, latency <300ms\n")
		},
	}
	c.Flags().StringVar(&cfgPath, "config", "", "Path to heravision.json config")
	return c
}

func setupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "setup",
		Short: "Setup MCP config for agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, _ := cmd.Flags().GetString("agent")
			all, _ := cmd.Flags().GetBool("all")
			if all {
				agent = "all"
			}
			if agent == "" {
				fmt.Fprintln(os.Stderr, "usage: heravision setup --agent opencode|claude|codex|cursor|all")
				fmt.Println(`Add manually:`)
				fmt.Println(`  opencode.json: {"mcp":{"heravision":{"type":"local","command":["heravision","mcp"],"enabled":true}}}`)
				fmt.Println(`  claude.json:   {"mcpServers":{"heravision":{"command":"heravision","args":["mcp"]}}}`)
				return nil
			}
			exe, _ := os.Executable()
			if exe == "" {
				exe = "heravision"
			}
			targets := []string{}
			switch agent {
			case "opencode":
				targets = []string{"opencode"}
			case "claude":
				targets = []string{"claude"}
			case "codex":
				targets = []string{"codex"}
			case "cursor":
				targets = []string{"cursor"}
			case "all":
				targets = []string{"opencode", "claude", "codex", "cursor"}
			default:
				return fmt.Errorf("unknown agent: %s", agent)
			}
			for _, t := range targets {
				if err := writeAgentConfig(t, exe); err != nil {
					fmt.Fprintf(os.Stderr, "[warn] %s: %v\n", t, err)
				} else {
					fmt.Fprintf(os.Stderr, "[ok] %s configured\n", t)
				}
			}
			return nil
		},
	}
	c.Flags().String("agent", "", "Agent: opencode|claude|codex|cursor")
	c.Flags().Bool("all", false, "Setup all agents")
	return c
}

func benchCmd() *cobra.Command {
	var cfgPath string
	c := &cobra.Command{
		Use:   "bench [image]",
		Short: "Benchmark extract latency",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, _ := cmd.Flags().GetInt("n")
			mode, _ := cmd.Flags().GetString("mode")
			path := "testdata/ui.png"
			if len(args) > 0 {
				path = args[0]
			}
			cfg, _, err := config.Load(cfgPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[warn] %v — using defaults\n", err)
			}
			start := time.Now()
			r, err := facts.Extract(path, mode, buildinfo.Version, cfg)
			if err != nil {
				return err
			}
			warmup := time.Since(start)
			var total time.Duration
			for i := 0; i < n; i++ {
				t0 := time.Now()
				_, err := facts.Extract(path, mode, buildinfo.Version, cfg)
				if err != nil {
					return err
				}
				total += time.Since(t0)
			}
			avg := float64(total.Microseconds()) / float64(n) / 1000
			fmt.Printf("bench %s [%s] x%d: avg %.2fms, boxes %d, first-run %v (incl. warmup)\n", path, mode, n, avg, len(r.Boxes), warmup)
			if avg > 300 {
				fmt.Fprintln(os.Stderr, "[warn] avg >300ms target exceeded")
			}
			return nil
		},
	}
	c.Flags().Int("n", 10, "iterations")
	c.Flags().String("mode", "general", "Mode: general|ui|code|diagram|error|blur")
	c.Flags().StringVar(&cfgPath, "config", "", "Path to heravision.json config")
	return c
}

func writeAgentConfig(agent, exe string) error {
	home, _ := os.UserHomeDir()
	var path string
	var content string
	switch agent {
	case "opencode":
		path = filepath.Join(home, ".config", "opencode", "opencode.json")
		b, _ := json.Marshal(map[string]interface{}{"mcp": map[string]interface{}{"heravision": map[string]interface{}{"type": "local", "command": []string{exe, "mcp"}, "enabled": true}}})
		content = string(b)
	case "claude":
		path = filepath.Join(home, ".claude.json")
		b, _ := json.Marshal(map[string]interface{}{"mcpServers": map[string]interface{}{"heravision": map[string]interface{}{"command": exe, "args": []string{"mcp"}}}})
		content = string(b)
	case "codex":
		path = filepath.Join(home, ".codex", "config.json")
		b, _ := json.Marshal(map[string]interface{}{"mcpServers": map[string]interface{}{"heravision": map[string]interface{}{"command": exe, "args": []string{"mcp"}}}})
		content = string(b)
	case "cursor":
		path = filepath.Join(home, ".cursor", "mcp.json")
		b, _ := json.Marshal(map[string]interface{}{"mcpServers": map[string]interface{}{"heravision": map[string]interface{}{"command": exe, "args": []string{"mcp"}}}})
		content = string(b)
	}
	if path == "" {
		return fmt.Errorf("no path")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var existing map[string]interface{}
		if json.Unmarshal(data, &existing) == nil {
			var patch map[string]interface{}
			if json.Unmarshal([]byte(content), &patch) == nil {
				for k, v := range patch {
					if em, ok := existing[k].(map[string]interface{}); ok {
						if pm, ok := v.(map[string]interface{}); ok {
							for pk, pv := range pm {
								em[pk] = pv
							}
							existing[k] = em
							continue
						}
					}
					existing[k] = v
				}
				merged, err := json.MarshalIndent(existing, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(path, merged, 0644); err == nil {
					fmt.Fprintf(os.Stderr, "[ok] %s merged: %s\n", agent, path)
					return nil
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[info] %s exists — manual merge needed: %s\n", agent, path)
		fmt.Println(content)
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[ok] %s created: %s\n", agent, path)
	return nil
}
