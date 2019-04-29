package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"golang.org/x/exp/shiny/widget/node"

	"github.com/mzki/erago/view/exp/text"
	"github.com/mzki/erago/view/exp/theme"
)

func buildTextView() *TextView {
	presenter := NewEragoPresenter(eventQueueStub{})
	v := NewTextView("default", presenter)
	Theme := &theme.Default
	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: v.MeasuredSize}
	v.Layout(Theme)
	return v
}

func TestTextView(t *testing.T) {
	v := buildTextView()
	defer v.Close()
	Theme := &theme.Default

	const Text = "てすとだよーん"
	v.Print(Text)
	if got := text.String(v.frame); strings.Compare(got, Text) != 0 {
		t.Errorf("diffrent frame content, got: %s, expect: %s", got, Text)
	}

	v.Printer.e.SetAlignment(text.AlignmentRight)
	v.Print(Jugem)

	rgba := image.NewRGBA(image.Rectangle{Max: screenSize})
	v.Mark(node.MarkNeedsPaintBase)
	if err := v.PaintBase(&node.PaintBaseContext{
		Theme: Theme,
		Dst:   rgba,
	}, image.ZP); err != nil {
		t.Errorf("PaintBase: %v", err)
	}
	saveImage("_v_test.png", rgba)
}

func TestTextViewMeasureLayout(t *testing.T) {
	v := buildTextView()
	defer v.Close()
	Theme := &theme.Default

	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: screenSize}
	v.Layout(Theme)

	v.NewPage()
	if vText := fmt.Sprint(v); vText == "" {
		t.Error("in layouted view, NewPage but empty content")
	}
}

func BenchmarkTextViewPrint(b *testing.B) {
	v := buildTextView()
	defer v.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Print(Jugem)
	}
}
