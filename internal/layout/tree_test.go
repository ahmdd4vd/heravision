package layout

import (
	"testing"

	"heravision/internal/detector"
)

func TestSidebarAndStackedContent(t *testing.T) {
	boxes := []detector.Box{
		{Type: "card", X: 10, Y: 10, W: 100, H: 300, Order: 1},
		{Type: "button", X: 200, Y: 10, W: 200, H: 140, Order: 2},
		{Type: "input", X: 200, Y: 170, W: 200, H: 140, Order: 3},
	}
	tree := Build(boxes, 500, 400)
	if tree.Type != "root" {
		t.Fatalf("expected root got %s", tree.Type)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("expected sidebar + row group, got %+v", tree.Children)
	}
	sidebar := tree.Children[0]
	if sidebar.Type != "card" || sidebar.Order != 1 {
		t.Fatalf("expected sidebar leaf card#1 got %+v", sidebar)
	}
	row := tree.Children[1]
	if row.Type != "col" || len(row.Children) != 2 {
		t.Fatalf("expected col with 2 children got %+v", row)
	}
	if row.Children[0].Type != "button" || row.Children[1].Type != "input" {
		t.Fatalf("col children wrong: %+v", row.Children)
	}
}

func TestRowsSplitTopDown(t *testing.T) {
	boxes := []detector.Box{
		{Type: "card", X: 50, Y: 20, W: 300, H: 60, Order: 1},
		{Type: "card", X: 50, Y: 200, W: 300, H: 60, Order: 2},
	}
	tree := Build(boxes, 400, 300)
	if len(tree.Children) != 2 {
		t.Fatalf("expected two top-level rows got %+v", tree.Children)
	}
	if tree.Children[0].Order != 1 || tree.Children[1].Order != 2 {
		t.Fatalf("reading order must propagate to leaves: %+v", tree.Children)
	}
}

func TestEmptyBoxesFallbackRegion(t *testing.T) {
	tree := Build(nil, 300, 200)
	if len(tree.Children) != 1 || tree.Children[0].Type != "region" {
		t.Fatalf("expected single region fallback got %+v", tree.Children)
	}
}

func TestInseparableBoxesFlatRegion(t *testing.T) {
	boxes := []detector.Box{
		{Type: "icon", X: 10, Y: 10, W: 30, H: 30, Order: 1},
		{Type: "icon", X: 14, Y: 14, W: 30, H: 30, Order: 2},
	}
	tree := Build(boxes, 100, 100)
	if len(tree.Children) != 1 || tree.Children[0].Type != "region" {
		t.Fatalf("overlapping boxes must collapse to flat region got %+v", tree.Children)
	}
	if len(tree.Children[0].Children) != 2 {
		t.Fatalf("region must contain both leaves got %+v", tree.Children[0].Children)
	}
}
