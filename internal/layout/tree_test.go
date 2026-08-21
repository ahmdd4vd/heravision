package layout

import (
	"testing"

	"heravision/internal/detector"
)

func TestHeaderBodyFooterSplit(t *testing.T) {
	boxes := []detector.Box{
		{Type: "button", X: 10, Y: 5, W: 40, H: 10},
		{Type: "card", X: 10, Y: 45, W: 60, H: 20},
		{Type: "card", X: 10, Y: 88, W: 60, H: 8},
	}
	tree := Build(boxes, 200, 100)
	if tree.Type != "root" {
		t.Fatalf("expected root got %s", tree.Type)
	}
	var kinds []string
	for _, ch := range tree.Children {
		kinds = append(kinds, ch.Type)
	}
	if len(kinds) != 3 || kinds[0] != "header" || kinds[1] != "body" || kinds[2] != "footer" {
		t.Fatalf("expected header/body/footer got %v", kinds)
	}
}

func TestEmptyBoxesFallbackBody(t *testing.T) {
	tree := Build(nil, 300, 200)
	if len(tree.Children) != 1 || tree.Children[0].Type != "body" {
		t.Fatalf("expected single body fallback got %+v", tree.Children)
	}
}
