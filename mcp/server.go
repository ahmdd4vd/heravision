package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"heravision/internal/color"
	"heravision/internal/detector"
	"heravision/internal/layout"
	"heravision/internal/ocr"
	"heravision/internal/processor"
)

func NewServer() *server.MCPServer {
	s := server.NewMCPServer("heravision", "0.1.0",
		server.WithToolCapabilities(true),
	)
	s.AddTool(mcp.NewTool("heravision_extract",
		mcp.WithDescription("Extract structured facts from image — texts, boxes, colors, layout. Use for any image when you are text-only and need to see."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Absolute or relative path to image (png/jpg/webp)")),
		mcp.WithString("mode", mcp.Description("Mode: general, ui, code, diagram, error"), mcp.DefaultString("general")),
	), handleExtract)
	s.AddTool(mcp.NewTool("heravision_compare",
		mcp.WithDescription("Compare two images and return diff facts — added/removed/moved boxes and texts."),
		mcp.WithString("path_a", mcp.Required(), mcp.Description("Path to first image")),
		mcp.WithString("path_b", mcp.Required(), mcp.Description("Path to second image")),
	), handleCompare)
	s.AddTool(mcp.NewTool("heravision_describe",
		mcp.WithDescription("Describe image as markdown — alias to extract but returns markdown only."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to image")),
		mcp.WithString("mode", mcp.Description("Mode: general, ui, code, diagram, error"), mcp.DefaultString("general")),
	), handleDescribe)
	return s
}

func Serve() error {
	s := NewServer()
	return server.ServeStdio(s)
}

func handleExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing path: %v", err)), nil
	}
	mode := req.GetString("mode", "general")
	img, _, err := processor.Decode(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("decode failed: %v", err)), nil
	}
	img = processor.FixOrientation(img)
	img = processor.Resize(img, 1024)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	boxes := detector.Detect(img)
	texts := ocr.Extract(img)
	dominant := color.Dominant(img, 5)
	bg := dominant[0]
	if len(dominant) == 0 {
		dominant = []string{"#FFFFFF"}
		bg = "#FFFFFF"
	}
	tree := layout.Build(boxes, w, h)
	result := map[string]interface{}{
		"meta": map[string]interface{}{"width": w, "height": h, "mode": mode, "path": path},
		"texts": texts, "boxes": boxes,
		"colors": map[string]interface{}{"dominant": dominant, "background": bg},
		"layout": tree, "lines": []interface{}{},
	}
	md := fmt.Sprintf("## HeraVision Extract (%s)\n- Size: %dx%d\n- Mode: %s\n- Texts: %d\n- Boxes: %d\n- Colors: %v\n- Layout: %s/%d header, %d body\n", mode, w, h, mode, len(texts), len(boxes), dominant, tree.Type, len(tree.Children), len(boxes))
	j, _ := json.MarshalIndent(result, "", "  ")
	combined := fmt.Sprintf("%s\n```json\n%s\n```\n", md, string(j))
	return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: combined}}}, nil
}

func handleCompare(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := req.RequireString("path_a")
	if err != nil {
		return mcp.NewToolResultError("missing path_a"), nil
	}
	bs, err := req.RequireString("path_b")
	if err != nil {
		return mcp.NewToolResultError("missing path_b"), nil
	}
	if _, err := os.Stat(a); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path_a not found: %s", a)), nil
	}
	if _, err := os.Stat(bs); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path_b not found: %s", bs)), nil
	}
	imgA, _, err := processor.Decode(a)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("decode a: %v", err)), nil
	}
	imgB, _, err := processor.Decode(bs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("decode b: %v", err)), nil
	}
	imgA = processor.Resize(imgA, 512)
	imgB = processor.Resize(imgB, 512)
	boxesA := detector.Detect(imgA)
	boxesB := detector.Detect(imgB)
	added, removed, moved := diffBoxes(boxesA, boxesB)
	diff := map[string]interface{}{
		"path_a": a, "path_b": bs,
		"diff": map[string]interface{}{"added": added, "removed": removed, "moved": moved, "color_changed": []interface{}{}},
		"counts": map[string]int{"a_boxes": len(boxesA), "b_boxes": len(boxesB)},
	}
	j, _ := json.MarshalIndent(diff, "", "  ")
	return mcp.NewToolResultText(string(j)), nil
}

func diffBoxes(a, b []detector.Box) ([]detector.Box, []detector.Box, []map[string]interface{}) {
	used := make([]bool, len(b))
	var added []detector.Box
	var removed []detector.Box
	var moved []map[string]interface{}
	if added == nil {
		added = []detector.Box{}
	}
	if removed == nil {
		removed = []detector.Box{}
	}
	if moved == nil {
		moved = []map[string]interface{}{}
	}
	for _, ba := range a {
		found := -1
		for j, bb := range b {
			if used[j] {
				continue
			}
			if iou(ba, bb) > 0.5 {
				found = j
				break
			}
		}
		if found == -1 {
			removed = append(removed, ba)
		} else {
			used[found] = true
			bb := b[found]
			if abs(ba.X-bb.X) > 5 || abs(ba.Y-bb.Y) > 5 {
				moved = append(moved, map[string]interface{}{"from": ba, "to": bb})
			}
		}
	}
	for j, bb := range b {
		if !used[j] {
			added = append(added, bb)
		}
	}
	return added, removed, moved
}

func iou(a, b detector.Box) float64 {
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.W, b.X+b.W)
	y2 := min(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	inter := float64((x2 - x1) * (y2 - y1))
	union := float64(a.W*a.H+b.W*b.H) - inter
	if union == 0 {
		return 0
	}
	return inter / union
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func handleDescribe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleExtract(ctx, req)
}
