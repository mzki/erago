package ui

import (
	"image"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"

	customT "github.com/mzki/erago/view/exp/theme"
)

// Edge is four edges in square shape, upper, bottom, left, and right.
type Edge uint8

const (
	EdgeNone Edge = iota
	EdgeTop
	EdgeBottom
	EdgeLeft
	EdgeRight
)

// whether Edge axis is Horizontal?
func (e Edge) Horizontal() bool { return e != EdgeNone && e == EdgeTop || e == EdgeBottom }

// whether Edge axis is Vertical?
func (e Edge) Vertical() bool { return e != EdgeNone && e == EdgeLeft || e == EdgeRight }

// FixedSplit splits itself to 2 widgets. its splitting line is
// same axis as Edge and BorderSize away from Edge.
// The size of Node close to Edge is fixed by BorderSize and
// another Node's one is chaged by FixedSplit's size.
// These sizes along with Edge are always expanded or shrinked
// to fit FixedSplit's size.
type FixedSplit struct {
	node.ContainerEmbed

	Edge       Edge
	BorderSize unit.Value
}

func NewFixedSplit(edge Edge, borderSize unit.Value, fixedNode, another node.Node) *FixedSplit {
	if fixedNode == nil || another == nil {
		panic("FixedSplit: 2 widgets must not be nil")
	}

	fsp := &FixedSplit{Edge: edge, BorderSize: borderSize}
	fsp.Wrapper = fsp
	// fixedNode must be place at first.
	fsp.ContainerEmbed.Insert(fixedNode, nil)
	fsp.ContainerEmbed.Insert(another, nil)
	return fsp
}

// can not insert new child after build.
func (fsp *FixedSplit) Insert(c, nextSibling node.Node) {
	panic("FixedSplit: inserting new child node is not allowed")
}

func (fsp *FixedSplit) fixedSize(t *theme.Theme) int {
	return customT.Pixels(t, fsp.BorderSize).Ceil() // TODO: or Round?
}

// implements node.Node interface
func (fsp *FixedSplit) Measure(t *theme.Theme, widthHint, heightHint int) {
	if fsp.Edge == EdgeNone {
		fsp.ContainerEmbed.Measure(t, widthHint, heightHint)
		return
	}
	if widthHint < 0 {
		widthHint = 0
	}
	if heightHint < 0 {
		heightHint = 0
	}
	// its size depends on parent's measuring.
	fsp.MeasuredSize = image.Point{widthHint, heightHint}

	fixedPixels := fsp.fixedSize(t)
	if fsp.Edge.Horizontal() {
		fsp.FirstChild.Wrapper.Measure(t, widthHint, fixedPixels)
		heightHint -= fixedPixels
		if heightHint < 0 {
			heightHint = 0
		}
	} else {
		fsp.FirstChild.Wrapper.Measure(t, fixedPixels, heightHint)
		widthHint -= fixedPixels
		if widthHint < 0 {
			widthHint = 0
		}
	}
	fsp.LastChild.Wrapper.Measure(t, widthHint, heightHint)
}

// implements node.Node interface
func (fsp *FixedSplit) Layout(t *theme.Theme) {
	fixedSize := fsp.fixedSize(t)
	fsSize := fsp.Rect.Size()

	// fixedRect is rect of FirstChild and restRect is LastChild.
	fixedRect := image.Rectangle{Max: fsSize}
	restRect := fixedRect

	switch fsp.Edge {
	case EdgeTop:
		borderY := fixedSize
		if maxY := fsSize.Y; maxY < borderY {
			borderY = maxY
		}
		fixedRect.Max.Y = borderY
		restRect.Min.Y = borderY

	case EdgeBottom:
		borderY := fsSize.Y - fixedSize
		if borderY < 0 {
			borderY = 0
		}
		fixedRect.Min.Y = borderY
		restRect.Max.Y = borderY

	case EdgeLeft:
		borderX := fixedSize
		if maxX := fsSize.X; maxX < borderX {
			borderX = maxX
		}
		fixedRect.Max.X = borderX
		restRect.Min.X = borderX

	case EdgeRight:
		borderX := fsSize.X - fixedSize
		if borderX < 0 {
			borderX = 0
		}
		fixedRect.Min.X = borderX
		restRect.Max.X = borderX

	default:
		panic("FixedSplit: unkown Edge")
	}

	if c := fsp.FirstChild; c != nil {
		c.Rect = fixedRect
		c.Wrapper.Layout(t)
	}
	if c := fsp.LastChild; c != nil {
		c.Rect = restRect
		c.Wrapper.Layout(t)
	}
}

func (fsp *FixedSplit) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	origin = origin.Add(fsp.Rect.Min)
	// Iterate backwards. Later children have priority over earlier children,
	// as later ones are usually drawn over earlier ones.
	for c := fsp.LastChild; c != nil; c = c.PrevSibling {
		if c.Wrapper.OnInputEvent(ev, origin) == node.Handled {
			return node.Handled
		}
	}
	return node.NotHandled
}
