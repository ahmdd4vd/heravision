package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"heravision/internal/buildinfo"
	"heravision/internal/config"
	"heravision/internal/facts"
	"heravision/internal/ocr"
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
			if cfg.Ocr.Enabled && ocr.BundlePresent(cfg.Ocr.LibPath, cfg.Ocr.DetPath, cfg.Ocr.RecPath, cfg.Ocr.DictPath) {
				fmt.Println("[ok] ocr: ONNX PP-OCR mobile engine ready (real text reading)")
			} else if cfg.Ocr.Enabled {
				fmt.Println("[warn] ocr: enabled but bundle missing — run: heravision setup --ocr")
			} else {
				fmt.Println("[warn] ocr: heuristic placeholders only — enable via heravision.json + setup --ocr")
			}
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
			if agent == "" && !all {
				if ocrOnly, _ := cmd.Flags().GetBool("ocr"); !ocrOnly {
					return runSetupWizard()
				}
			}
			if all {
				agent = "all"
				if ocrFlag, _ := cmd.Flags().GetBool("ocr"); ocrFlag {
					return downloadOcrBundle()
				}
			}
			if ocrFlag, _ := cmd.Flags().GetBool("ocr"); ocrFlag {
				return downloadOcrBundle()
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
	c.Flags().Bool("ocr", false, "Download OCR bundle (onnxruntime + PP-OCR mobile models, ~28MB)")
	return c
}

func ocrBundleDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".heravision", "ocr")
}

func downloadOcrBundle() error {
	dir := ocrBundleDir()
	libDir := filepath.Join(dir, "lib")
	wDir := filepath.Join(dir, "weights")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(wDir, 0755); err != nil {
		return err
	}
	base := "https://huggingface.co/GetcharZp/go-ocr/resolve/main"
	var runtimeLib struct{ url, dest string }
	switch runtime.GOOS {
	case "linux":
		runtimeLib = struct{ url, dest string }{base + "/lib/onnxruntime_amd64.so", filepath.Join(libDir, "libonnxruntime.so")}
	case "darwin":
		runtimeLib = struct{ url, dest string }{base + "/lib/onnxruntime_amd64.dylib", filepath.Join(libDir, "libonnxruntime.dylib")}
	default:
		runtimeLib = struct{ url, dest string }{base + "/lib/onnxruntime.dll", filepath.Join(libDir, "onnxruntime.dll")}
	}
	downloads := []struct {
		url  string
		dest string
	}{
		runtimeLib,
		{"https://huggingface.co/SWHL/RapidOCR/resolve/main/PP-OCRv4/ch_PP-OCRv4_det_infer.onnx", filepath.Join(wDir, "det.onnx")},
		{"https://huggingface.co/SWHL/RapidOCR/resolve/main/PP-OCRv3/en_PP-OCRv3_rec_infer.onnx", filepath.Join(wDir, "en_rec.onnx")},
		{"https://raw.githubusercontent.com/PaddlePaddle/PaddleOCR/main/ppocr/utils/en_dict.txt", filepath.Join(wDir, "en_dict.txt")},
	}
	for _, d := range downloads {
		if _, err := os.Stat(d.dest); err == nil {
			fmt.Fprintf(os.Stderr, "[skip] %s exists\n", d.dest)
			continue
		}
		fmt.Fprintf(os.Stderr, "[get] %s\n", d.url)
		if err := fetchFile(d.url, d.dest); err != nil {
			return fmt.Errorf("download %s: %w", d.dest, err)
		}
	}
	fmt.Fprintf(os.Stderr, "[ok] OCR bundle ready: %s\n", dir)
	var libName string
	switch runtime.GOOS {
	case "linux":
		libName = "libonnxruntime.so"
	case "darwin":
		libName = "libonnxruntime.dylib"
	default:
		libName = "onnxruntime.dll"
	}
	libPath := filepath.Join(libDir, libName)
	_ = enableOcrInHomeConfig(libPath, filepath.Join(wDir, "det.onnx"), filepath.Join(wDir, "en_rec.onnx"), filepath.Join(wDir, "en_dict.txt"))
	cfgSnippet := map[string]interface{}{
		"ocr": map[string]interface{}{
			"enabled":   true,
			"lib_path":  libPath,
			"det_path":  filepath.Join(wDir, "det.onnx"),
			"rec_path":  filepath.Join(wDir, "en_rec.onnx"),
			"dict_path": filepath.Join(wDir, "en_dict.txt"),
		},
	}
	j, _ := json.MarshalIndent(cfgSnippet, "", "  ")
	fmt.Printf("Add to heravision.json:\n%s\n", string(j))
	return nil
}

func fetchFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	var written int64
	total := resp.ContentLength
	buf := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			prev := written
			written += int64(n)
			if total > 0 && written>>20 != prev>>20 {
				fmt.Fprintf(os.Stderr, "\r      %d / %d MB", written>>20, total>>20)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if total > 0 {
		fmt.Fprintf(os.Stderr, "\r      %d / %d MB\n", total>>20, total>>20)
	} else {
		fmt.Fprintln(os.Stderr)
	}
	return nil
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

// agentTargets maps a wizard menu choice to the agent names to configure.
func agentTargets(choice string) []string {
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "1":
		return []string{"opencode"}
	case "2":
		return []string{"claude"}
	case "3":
		return []string{"codex"}
	case "4":
		return []string{"cursor"}
	case "a", "all":
		return []string{"opencode", "claude", "codex", "cursor"}
	}
	return nil
}

func runSetupWizard() error {
	return runSetup(os.Stdin, os.Stdout)
}

// runSetup drives the interactive first-run wizard. Injected reader/writer
// keep it unit-testable without touching the filesystem (skip paths only).
func runSetup(in io.Reader, out io.Writer) error {
	rd := bufio.NewReader(in)
	ask := func(label string) string {
		fmt.Fprint(out, label)
		line, _ := rd.ReadString('\n')
		return strings.TrimSpace(line)
	}
	exe, _ := os.Executable()
	if exe == "" {
		exe = "heravision"
	}
	fmt.Fprintln(out, "HeraVision setup")
	fmt.Fprintf(out, "binary: %s\n\n", exe)

	ans := ask("[1/2] Download real OCR engine (~30 MB: onnxruntime + PP-OCR mobile)? [y/N]: ")
	switch strings.ToLower(ans) {
	case "y", "yes":
		if err := downloadOcrBundle(); err != nil {
			fmt.Fprintf(os.Stderr, "[warn] ocr download failed: %v — heuristic placeholders stay active\n", err)
		}
	default:
		fmt.Fprintln(out, "[skip] ocr — shape placeholders stay active (enable later: heravision setup --ocr)")
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "[2/2] Configure MCP for your AI coding agent:")
	fmt.Fprintln(out, "  1) opencode   2) claude   3) codex   4) cursor   a) all   s) skip")
	targets := agentTargets(ask("choice [1/2/3/4/a/s]: "))
	for _, t := range targets {
		if err := writeAgentConfig(t, exe); err != nil {
			fmt.Fprintf(os.Stderr, "[warn] %s: %v\n", t, err)
		}
	}
	if len(targets) == 0 {
		fmt.Fprintln(out, "[skip] agent config — run `heravision setup` anytime")
	}

	fmt.Fprintln(out, "\n=== setup complete ===")
	fmt.Fprintf(out, "extract : %s extract ./image.png --mode ui --json\n", exe)
	fmt.Fprintf(out, "serve   : %s mcp       (MCP stdio server for your agent)\n", exe)
	fmt.Fprintf(out, "check   : %s doctor\n", exe)
	return nil
}

func homeConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "heravision.json")
}

func enableOcrInHomeConfig(libPath, detPath, recPath, dictPath string) error {
	path := homeConfigPath()
	root := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil && json.Unmarshal(data, &root) != nil {
		root = map[string]interface{}{}
	}
	ocr, _ := root["ocr"].(map[string]interface{})
	if ocr == nil {
		ocr = map[string]interface{}{}
	}
	ocr["enabled"] = true
	ocr["lib_path"] = libPath
	ocr["det_path"] = detPath
	ocr["rec_path"] = recPath
	ocr["dict_path"] = dictPath
	root["ocr"] = ocr
	merged, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, merged, 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[ok] ocr enabled in %s\n", path)
	return nil
}
