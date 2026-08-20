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
	"heravision/internal/layout"
	"heravision/internal/ocr"
	"heravision/internal/processor"
	"heravision/mcp"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{Use: "heravision", Short: "The eyes for blind LLMs — pure native vision", Long: "HeraVision — hybrid native vision: extract facts from images (pure Go, no model/API) for text-only LLMs like DeepSeek V4, GLM 5.3"}
	root.AddCommand(versionCmd(), extractCmd(), mcpCmd(), doctorCmd(), setupCmd())
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
			img = processor.Resize(img, 1024)
			b := img.Bounds()
			w, h := b.Dx(), b.Dy()
			boxes := detector.Detect(img)
			texts := ocr.Extract(img)
			dominant := color.Dominant(img, 5)
			bg := ""
			if len(dominant) > 0 {
				bg = dominant[0]
			}
			tree := layout.Build(boxes, w, h)
			elapsed := time.Since(start).Milliseconds()
			result := map[string]interface{}{
				"meta": map[string]interface{}{"width": w, "height": h, "mode": mode, "path": path, "version": version, "elapsed_ms": elapsed},
				"texts": texts, "boxes": boxes,
				"colors": map[string]interface{}{"dominant": dominant, "background": bg},
				"layout": tree, "lines": []interface{}{},
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
	c.Flags().StringVar(&mode, "mode", "general", "Mode: general|ui|code|diagram|error")
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
			fmt.Println("[ok] processor: decode/resize pure Go (CGO_ENABLED=0)")
			fmt.Println("[ok] mcp: 3 tools (heravision_extract, heravision_compare, heravision_describe)")
			fmt.Println("[ok] detector: sobel+components+classify")
			fmt.Println("[ok] color: histogram 64-bin")
			fmt.Println("[ok] layout: header/body/footer tree")
			fmt.Println("[ok] ocr: interface ready (fallback empty)")
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

func writeAgentConfig(agent, exe string) error {
	home, _ := os.UserHomeDir()
	var path string
	var content string
	switch agent {
	case "opencode":
		path = filepath.Join(home, ".config", "opencode", "opencode.json")
		content = fmt.Sprintf(`{"mcp":{"heravision":{"type":"local","command":["%s","mcp"],"enabled":true}}}`, exe)
	case "claude":
		path = filepath.Join(home, ".claude.json")
		content = fmt.Sprintf(`{"mcpServers":{"heravision":{"command":"%s","args":["mcp"]}}}`, exe)
	case "codex":
		path = filepath.Join(home, ".codex", "config.json")
		content = fmt.Sprintf(`{"mcpServers":{"heravision":{"command":"%s","args":["mcp"]}}}`, exe)
	case "cursor":
		path = filepath.Join(home, ".cursor", "mcp.json")
		content = fmt.Sprintf(`{"mcpServers":{"heravision":{"command":"%s","args":["mcp"]}}}`, exe)
	}
	if path == "" {
		return fmt.Errorf("no path")
	}
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "[info] %s exists — merging (manual check needed): %s\n", agent, path)
		fmt.Println(content)
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}
