package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"heravision/internal/buildinfo"
	"heravision/internal/config"
	"heravision/internal/facts"
)

func NewServer() *server.MCPServer {
	s := server.NewMCPServer("heravision", buildinfo.Version,
		server.WithToolCapabilities(true),
	)
	s.AddTool(mcp.NewTool("heravision_extract",
		mcp.WithDescription("Extract UI structure facts from an image: page type guess (login/dashboard/terminal/chat/form), element boxes (button/input/card/image/text_block/icon/checkbox/avatar) with position, size, color, reading order and caption; grids; dominant colors; layout tree; optional mermaid graph. Text is NOT OCR-read - text fields are shape placeholders like [button]. Use when you are a text-only model and need to see layout."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Absolute or relative path to image (png/jpg/webp)")),
		mcp.WithString("mode", mcp.Description("Mode: general, ui, code, diagram, error, blur"), mcp.DefaultString("general")),
	), handleExtract)
	s.AddTool(mcp.NewTool("heravision_compare",
		mcp.WithDescription("Compare two images and return structural diff: added/removed/moved boxes and color changes."),
		mcp.WithString("path_a", mcp.Required(), mcp.Description("Path to first image")),
		mcp.WithString("path_b", mcp.Required(), mcp.Description("Path to second image")),
	), handleCompare)
	s.AddTool(mcp.NewTool("heravision_describe",
		mcp.WithDescription("Describe image structure as markdown only (no JSON block). Alias of extract with markdown output."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to image")),
		mcp.WithString("mode", mcp.Description("Mode: general, ui, code, diagram, error, blur"), mcp.DefaultString("general")),
	), handleDescribe)
	return s
}

func Serve() error {
	s := NewServer()
	return server.ServeStdio(s)
}

func loadCfg() config.Config {
	cfg, path, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warn] config %s: %v\n", path, err)
	}
	return cfg
}

func handleExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing path: %v", err)), nil
	}
	mode := req.GetString("mode", "general")
	r, err := facts.Extract(path, mode, buildinfo.Version, loadCfg())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("extract failed: %v", err)), nil
	}
	j, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal: %v", err)), nil
	}
	combined := fmt.Sprintf("%s\n```json\n%s\n```\n", r.Markdown, string(j))
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
	res, err := facts.Compare(a, bs, loadCfg())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("compare failed: %v", err)), nil
	}
	j, _ := json.MarshalIndent(res, "", "  ")
	return mcp.NewToolResultText(string(j)), nil
}

func handleDescribe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing path: %v", err)), nil
	}
	mode := req.GetString("mode", "general")
	r, err := facts.Extract(path, mode, buildinfo.Version, loadCfg())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("describe failed: %v", err)), nil
	}
	return mcp.NewToolResultText(r.Markdown), nil
}
