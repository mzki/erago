package ui

import (
	"image"
	"testing"

	"local/erago/view/exp/theme"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
)

func TestFixedSplit(t *testing.T) {
	fixed := NewFixedSplit(EdgeLeft, unit.Pixels(100),
		widget.NewSpace(),
		widget.NewSpace(),
	)
	Theme := &theme.Default

	for _, test := range []struct {
		Edge Edge
		Rect image.Rectangle
	}{
		{
			EdgeLeft,
			image.Rectangle{
				Min: image.Point{},
				Max: image.Point{100, screenSize.Y},
			},
		},
		{
			EdgeTop,
			image.Rectangle{
				Min: image.Point{},
				Max: image.Point{screenSize.X, 100},
			},
		},
		{
			EdgeBottom,
			image.Rectangle{
				Min: image.Point{0, screenSize.Y - 100},
				Max: screenSize,
			},
		},
		{
			EdgeRight,
			image.Rectangle{
				Min: image.Point{screenSize.X - 100, 0},
				Max: screenSize,
			},
		},
	} {
		fixed.Edge = test.Edge
		fixed.Measure(Theme, screenSize.X, screenSize.Y)
		fixed.Rect = image.Rectangle{Max: screenSize}
		fixed.Layout(Theme)
		if cRect := fixed.FirstChild.Rect; cRect != test.Rect {
			t.Errorf("diffrent fixed Rectangle. got: %v, expect: %v", cRect, test.Rect)
		}
	}
}
