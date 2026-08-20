package layout

import "heravision/internal/detector"

type Node struct {
	Type     string `json:"type"`
	X        int    `json:"x,omitempty"`
	Y        int    `json:"y,omitempty"`
	W        int    `json:"w,omitempty"`
	H        int    `json:"h,omitempty"`
	Children []Node `json:"children,omitempty"`
}

func Build(boxes []detector.Box, imgW, imgH int) Node {
	if len(boxes) == 0 {
		return Node{Type: "root", Children: []Node{{Type: "body", X: 0, Y: 0, W: imgW, H: imgH}}}
	}
	headerEnd := imgH / 5
	footerStart := imgH * 4 / 5
	var header, body, footer []Node
	for _, b := range boxes {
		n := Node{Type: b.Type, X: b.X, Y: b.Y, W: b.W, H: b.H}
		cy := b.Y + b.H/2
		switch {
		case cy < headerEnd:
			header = append(header, n)
		case cy > footerStart:
			footer = append(footer, n)
		default:
			body = append(body, n)
		}
	}
	children := []Node{}
	if len(header) > 0 {
		children = append(children, Node{Type: "header", Children: header})
	}
	children = append(children, Node{Type: "body", Children: body})
	if len(footer) > 0 {
		children = append(children, Node{Type: "footer", Children: footer})
	}
	return Node{Type: "root", W: imgW, H: imgH, Children: children}
}
