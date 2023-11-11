package ui

import (
	"image"

	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
)

// Sheet is same as widget.Sheet, but avoid some bugs cases, such as empty rect.
type Sheet struct {
	*widget.Sheet
}

func NewSheet(inner node.Node) *Sheet {
	s := &Sheet{widget.NewSheet(inner)}
	s.Wrapper = s
	return s
}

func (s *Sheet) Paint(ctx *node.PaintContext, origin image.Point) error {
	// skip insufficient paint space, which causes invalid argument on buffer creation in original Sheet.
	if r := s.Wrappee().Rect; r.Dx() == 0 || r.Dy() == 0 {
		s.Marks.UnmarkNeedsPaint()
		return nil
	}
	return s.Sheet.Paint(ctx, origin)
}
