package facts

import (
	"strings"
	"testing"

	"heravision/internal/config"
	"heravision/internal/detector"
)

func TestColorNameBasic(t *testing.T) {
	cases := map[string]string{
		"#FF0000": "red",
		"#000000": "black",
		"#FFFFFF": "white",
		"#3B82F6": "blue",
		"#808080": "gray",
		"#BDBBB9": "light gray",
		"#8998A4": "gray",
	}
	for hex, want := range cases {
		if got := colorName(hex); got != want {
			t.Errorf("colorName(%s)=%s want %s", hex, got, want)
		}
	}
}

func TestBuildCaption(t *testing.T) {
	bx := detector.Box{Type: "button", X: 700, Y: 800, W: 200, H: 40, Color: "#3B82F6"}
	cap := buildCaption(bx, 1000, 1000)
	for _, want := range []string{"blue", "button", "bottom-right"} {
		if !strings.Contains(cap, want) {
			t.Errorf("caption %q missing %q", cap, want)
		}
	}
}

func TestClassifyPageLogin(t *testing.T) {
	boxes := []detector.Box{
		{Type: "card", X: 350, Y: 250, W: 300, H: 40},
		{Type: "input", X: 380, Y: 380, W: 240, H: 36},
		{Type: "input", X: 380, Y: 440, W: 240, H: 36},
		{Type: "button", X: 420, Y: 520, W: 160, H: 44},
	}
	pt, conf := classifyPage(boxes, "#FFFFFF", 1000, 800, nil)
	if pt != "login" {
		t.Fatalf("expected login got %s (%.2f)", pt, conf)
	}
	if conf < 0.45 {
		t.Fatalf("confidence too low: %.2f", conf)
	}
}

func TestClassifyPageDashboard(t *testing.T) {
	grids := []Grid{{Cell: "card", Cols: 3, Rows: 2, Count: 6}}
	boxes := []detector.Box{
		{Type: "card", X: 10, Y: 10, W: 900, H: 60},
		{Type: "card", X: 10, Y: 100, W: 290, H: 180},
		{Type: "card", X: 310, Y: 100, W: 290, H: 180},
		{Type: "card", X: 610, Y: 100, W: 290, H: 180},
		{Type: "card", X: 10, Y: 300, W: 290, H: 180},
		{Type: "card", X: 310, Y: 300, W: 290, H: 180},
		{Type: "card", X: 610, Y: 300, W: 290, H: 180},
	}
	pt, _ := classifyPage(boxes, "#FFFFFF", 920, 500, grids)
	if pt != "dashboard" {
		t.Fatalf("expected dashboard got %s", pt)
	}
}

func TestClassifyPageTerminal(t *testing.T) {
	var boxes []detector.Box
	for i := 0; i < 5; i++ {
		boxes = append(boxes, detector.Box{Type: "text_block", X: 20, Y: 30 + i*40, W: 600, H: 24})
	}
	pt, _ := classifyPage(boxes, "#111111", 640, 400, nil)
	if pt != "terminal" {
		t.Fatalf("expected terminal got %s", pt)
	}
}

func TestDetectGrids3x2(t *testing.T) {
	var boxes []detector.Box
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			boxes = append(boxes, detector.Box{
				Type: "card",
				X:    10 + col*120,
				Y:    10 + row*80,
				W:    100,
				H:    60,
			})
		}
	}
	grids := detectGrids(boxes)
	if len(grids) == 0 {
		t.Fatal("expected grid detection")
	}
	g := grids[0]
	if g.Cell != "card" || g.Cols != 3 || g.Rows != 2 {
		t.Fatalf("expected card 3x2 got %+v", g)
	}
}

func TestReadingOrderAssigned(t *testing.T) {
	r, err := Extract("../../testdata/ui.png", "ui", "test", config.Default())
	if err != nil {
		t.Fatal(err)
	}
	for i, bx := range r.Boxes {
		if bx.Order != i+1 {
			t.Fatalf("order must be sequential: box %d has order %d", i, bx.Order)
		}
		if bx.Caption == "" {
			t.Fatalf("caption must be filled for box %d", i)
		}
	}
	if r.PageType == "" {
		t.Fatal("page_type must never be empty")
	}
}

func TestDiffSummaryMoved(t *testing.T) {
	res := &CompareResult{
		Diff: Diff{
			Moved:        []Move{{From: detector.Box{Type: "button", X: 48, Y: 28}, To: detector.Box{Type: "button", X: 52, Y: 30}}},
			ColorChanged: []Move{{From: detector.Box{Type: "button", Color: "#0000FF"}, To: detector.Box{Type: "button", Color: "#FF0000"}}},
			Added:        []detector.Box{{Type: "card", X: 5, Y: 5}},
		},
	}
	s := buildDiffSummary(res)
	for _, want := range []string{"1 element(s) moved", "changed color", "1 added"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary %q missing %q", s, want)
		}
	}
	if s2 := buildDiffSummary(&CompareResult{}); s2 != "no structural changes detected" {
		t.Errorf("empty diff summary wrong: %s", s2)
	}
}
