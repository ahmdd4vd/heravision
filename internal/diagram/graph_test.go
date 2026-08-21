package diagram

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"heravision/internal/detector"
	"heravision/internal/layout"
)

func drawOutlinedRect(img *image.RGBA, x0, y0, w, h int) {
	black := color.RGBA{0, 0, 0, 255}
	for k := 0; k < 2; k++ {
		for x := x0; x < x0+w; x++ {
			img.Set(x, y0+k, black)
			img.Set(x, y0+h-1-k, black)
		}
		for y := y0; y < y0+h; y++ {
			img.Set(x0+k, y, black)
			img.Set(x0+w-1-k, y, black)
		}
	}
}

func TestMermaidGraphWithArrow(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	drawOutlinedRect(img, 30, 80, 80, 40)
	drawOutlinedRect(img, 290, 80, 80, 40)
	black := color.RGBA{0, 0, 0, 255}
	for x := 118; x < 282; x++ {
		img.Set(x, 100, black)
	}
	for d := 0; d < 5; d++ {
		for dy := -d; dy <= d; dy++ {
			img.Set(281-d, 100+dy, black)
		}
	}
	edges := detector.EdgeMap(img, detector.DefaultParams)
	boxes := detector.DetectCfg(img, detector.DefaultParams)
	if len(boxes) < 2 {
		t.Fatalf("fixture must yield 2 boxes got %d", len(boxes))
	}
	out := ToMermaidGraph(boxes, edges, buildTestTree(boxes))
	if !strings.Contains(out, "flowchart TD") {
		t.Fatalf("missing flowchart header: %q", out)
	}
	if !strings.Contains(out, "-->") {
		t.Fatalf("arrow connector must produce edge, got %q", out)
	}
}

func TestMermaidGraphFallbackHierarchy(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	drawOutlinedRect(img, 30, 80, 80, 40)
	drawOutlinedRect(img, 290, 80, 80, 40)
	edges := detector.EdgeMap(img, detector.DefaultParams)
	boxes := detector.DetectCfg(img, detector.DefaultParams)
	out := ToMermaidGraph(boxes, edges, buildTestTree(boxes))
	if strings.Contains(out, "N0 --> N1") {
		return
	}
	if !strings.Contains(out, "H") {
		t.Fatalf("fallback must emit hierarchy nodes, got %q", out)
	}
}

func buildTestTree(boxes []detector.Box) layout.Node {
	children := make([]layout.Node, 0, len(boxes))
	for _, b := range boxes {
		children = append(children, layout.Node{Type: b.Type, Order: b.Order})
	}
	return layout.Node{Type: "root", Children: children}
}

func TestFindConnectorsIgnoresBoxOutlines(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	drawOutlinedRect(img, 30, 80, 80, 40)
	drawOutlinedRect(img, 290, 80, 80, 40)
	edges := detector.EdgeMap(img, detector.DefaultParams)
	boxes := detector.DetectCfg(img, detector.DefaultParams)
	conns := findConnectors(edges, boxes)
	if len(conns) != 0 {
		t.Fatalf("box outlines must not become connectors got %d", len(conns))
	}
}
