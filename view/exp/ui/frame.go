package ui

import (
	"image"
	"image/draw"

	"golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
)

// frame is a shell widget which paints frame on its margin space.
type Frame struct {
	*widget.Padder

	// color to draw frame.
	ThemeColor theme.Color
}

func NewFrame(margin unit.Value, color theme.Color, inner node.Node) *Frame {
	padder := widget.NewPadder(widget.AxisBoth, margin, inner)
	f := &Frame{
		Padder:     padder,
		ThemeColor: color,
	}
	f.Wrapper = f
	return f
}

// implements node.Node interface.
// Paint flikers the screen so deprecated.
func (w *Frame) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	if err := w.Padder.PaintBase(ctx, origin); err != nil {
		return err
	}

	t := ctx.Theme
	c := w.ThemeColor.Uniform(t)
	wr := w.Rect.Add(origin)
	for _, r := range imageutil.Border(wr, t.Pixels(w.Margin).Round()) {
		draw.Draw(ctx.Dst, r, c, image.Point{}, draw.Src)
	}
	return nil
}

// implements node.Node interface.
// Paint flikers the screen so deprecated.
// func (w *Frame) Paint(ctx *node.PaintContext, origin image.Point) error {
// 	if err := w.Padder.Paint(ctx, origin); err != nil {
// 		return err
// 	}
//
// 	t := ctx.Theme
// 	c := w.ThemeColor.Color(t)
// 	wr := w.Rect.Add(origin)
// 	for _, r := range imageutil.Border(wr, t.Pixels(w.Margin).Round()) {
// 		ctx.Drawer.DrawUniform(ctx.Src2Dst, c, r, screen.Over, nil)
// 	}
// 	return nil
// }
