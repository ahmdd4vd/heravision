package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"heravision/internal/color"
	"heravision/internal/detector"
	"heravision/internal/diagram"
	"heravision/internal/layout"
	"heravision/internal/ocr"
	"heravision/internal/processor"
	"heravision/mcp"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{Use: "heravision", Short: "The eyes for blind LLMs — pure native vision", Long: "HeraVision — hybrid native vision: extract facts from images (pure Go, no model/API) for text-only LLMs like DeepSeek V4, GLM 5.3"}
	root.AddCommand(versionCmd(), extractCmd(), mcpCmd(), doctorCmd(), setupCmd(), benchCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Print version", Run: func(cmd *cobra.Command, args []string) { fmt.Printf("heravision %s\n", version) }}
}

func extractCmd() *cobra.Command {
	var mode string
	var asJSON bool
	c := &cobra.Command{
		Use: "extract <image>", Short: "Extract structured facts from image", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			start := time.Now()
			img, _, err := processor.Decode(path)
			if err != nil {
				return err
			}
				img = processor.FixOrientation(img)
			img = processor.Preprocess(img, mode)
			img = processor.Resize(img, 1024)
			b := img.Bounds()
			w, h := b.Dx(), b.Dy()
			boxes := detector.Detect(img)
			texts := ocr.Extract(img)
			dominant := color.Dominant(img, 5)
			bg := color.Background(img)
			if bg == "" && len(dominant) > 0 {
				bg = dominant[0]
			}
			tree := layout.Build(boxes, w, h)
			var mermaid string
			if mode == "diagram" {
				mermaid = diagram.ToMermaid(boxes)
			}
			elapsed := time.Since(start).Milliseconds()
			result := map[string]interface{}{
				"meta": map[string]interface{}{"width": w, "height": h, "mode": mode, "path": path, "version": version, "elapsed_ms": elapsed},
				"texts": texts, "boxes": boxes,
				"colors": map[string]interface{}{"dominant": dominant, "background": bg},
				"layout": tree, "lines": []interface{}{},
				"mermaid": mermaid,
				"markdown": fmt.Sprintf("## Image Facts\n- Size: %dx%d\n- Texts: %d\n- Boxes: %d\n- Colors: %v", w, h, len(texts), len(boxes), dominant),
			}
			if asJSON {
				j, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(j))
				return nil
			}
			j, _ := json.MarshalIndent(result, "", "  ")
			fmt.Printf("## HeraVision Extract (%s)\n- Size: %dx%d\n- Texts: %d\n- Boxes: %d\n- Colors: %v\n- Elapsed: %dms\n\n```json\n%s\n```\n", mode, w, h, len(texts), len(boxes), dominant, elapsed, string(j))
			return nil
		},
	}
	c.Flags().StringVar(&mode, "mode", "general", "Mode: general|ui|code|diagram|error|blur")
	c.Flags().BoolVar(&asJSON, "json", false, "Output JSON only")
	return c
}

func mcpCmd() *cobra.Command {
	return &cobra.Command{Use: "mcp", Short: "Run MCP stdio server", RunE: func(cmd *cobra.Command, args []string) error { return mcp.Serve() }}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use: "doctor", Short: "Check heravision setup",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("heravision %s\n", version)
			fmt.Println("[ok] binary: runnable")
			exe, _ := os.Executable()
			if exe != "" {
				fmt.Printf("[ok] exe: %s\n", exe)
			}
			fi, err := os.Stat("testdata")
			if err == nil && fi.IsDir() {
				fmt.Println("[ok] testdata: present")
			} else {
				fmt.Println("[warn] testdata: missing")
			}
			fmt.Println("[ok] processor: decode/resize/preprocess B++ (CLAHE+Unsharp+SR, CGO_ENABLED=0)")
			fmt.Println("[ok] mcp: 3 tools (heravision_extract, heravision_compare, heravision_describe) + mermaid")
			fmt.Println("[ok] detector: Canny hysteresis 8-connect morph classify v2")
			fmt.Println("[ok] color: Lab k-means 5 ΔE merge + bg border")
			fmt.Println("[ok] layout: whitespace rows/cols")
			fmt.Println("[ok] ocr: heuristic + WASM stub + SR")
		},
	}
}

func setupCmd() *cobra.Command {
	c := &cobra.Command{
		Use: "setup", Short: "Setup MCP config for agents",
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
	c := &cobra.Command{
		Use: "bench [image]", Short: "Benchmark extract latency",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, _ := cmd.Flags().GetInt("n")
			path := "testdata/ui.png"
			if len(args) > 0 {
				path = args[0]
			}
			img, _, err := processor.Decode(path)
			if err != nil {
				return err
			}
			img = processor.FixOrientation(img)
			img = processor.Resize(img, 1024)
			var total time.Duration
			var boxes int
			for i := 0; i < n; i++ {
				start := time.Now()
				b := detector.Detect(img)
				_ = color.Dominant(img, 5)
				_ = ocr.Extract(img)
				boxes = len(b)
				total += time.Since(start)
			}
			avg := float64(total.Microseconds()) / float64(n) / 1000
			fmt.Printf("bench %s x%d: avg %.2fms, boxes %d, total %v\n", path, n, avg, boxes, total)
			if avg > 100 {
				fmt.Fprintln(os.Stderr, "[warn] avg >100ms target exceeded (without OCR)")
			}
			return nil
		},
	}
	c.Flags().Int("n", 10, "iterations")
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
	_ = os.MkdirAll(dir, 0755)
	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		var existing map[string]interface{}
		if json.Unmarshal(data, &existing) == nil {
			var patch map[string]interface{}
			_ = json.Unmarshal([]byte(content), &patch)
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
			merged, _ := json.MarshalIndent(existing, "", "  ")
			if err := os.WriteFile(path, merged, 0644); err == nil {
				fmt.Fprintf(os.Stderr, "[ok] %s merged: %s\n", agent, path)
				return nil
			}
		}
		fmt.Fprintf(os.Stderr, "[info] %s exists — manual merge needed: %s\n", agent, path)
		fmt.Println(content)
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}
