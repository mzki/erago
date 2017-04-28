package ui

import (
	"image"
	"strings"
	"testing"

	"golang.org/x/exp/shiny/widget/node"

	attr "local/erago/attribute"
	"local/erago/view/exp/text"
	"local/erago/view/exp/theme"
)

func buildTabView() *TabView {
	presenter := NewEragoPresenter(eventQueueStub{})
	v := NewTabView(presenter)
	Theme := &theme.Default
	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: screenSize}
	v.Layout(Theme)
	return v
}

func TestTabView(t *testing.T) {
	v := buildTabView()
	Theme := &theme.Default

	const Text = "てすとだよーん"
	v.Print(Text)
	if got := text.String(v.content.viewManager.currentView().frame); strings.Compare(got, Text) != 0 {
		t.Errorf("diffrent frame content, got: %s, expect: %s", got, Text)
	}

	v.Print(Jugem)

	rgba := image.NewRGBA(image.Rectangle{Max: screenSize})
	if err := v.PaintBase(&node.PaintBaseContext{
		Theme: Theme,
		Dst:   rgba,
	}, image.ZP); err != nil {
		t.Errorf("PaintBase: %v", err)
	}
	saveImage("_tabview_test.png", rgba)
}

func TestTabViewSetLayout(t *testing.T) {
	v := buildTabView()
	v.Print(Jugem)

	if err := v.SetLayout(
		attr.NewFlowHorizontal(
			attr.WithParentValue(attr.NewSingleText("left"), 1),
			attr.WithParentValue(attr.NewSingleText("right"), 1),
		),
	); err != nil {
		t.Fatal(err)
	}

	if vs := v.content.viewManager.textViews; len(vs) != 2 {
		t.Fatalf("different view number, got: %d, expect: %d", len(vs), 2)
	}

	if got := v.content.viewManager.currentView().name; got != "right" {
		t.Fatalf("different current view name, got: %v, expect: %v", got, "right")
	}

	v.Print(Jugem)

	Theme := &theme.Default
	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: screenSize}
	v.Layout(Theme)

	if err := v.SetCurrentView("right"); err != nil {
		t.Error("can not change current view")
	}
	v.Print("text to right")
}
