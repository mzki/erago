package ui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/exp/shiny/widget/node"

	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text"
	"github.com/mzki/erago/view/exp/theme"
)

var screenSize = image.Pt(256, 256)
var realScreenSize = image.Pt(1800, 1024)

const Jugem = `
寿限無、寿限無
五劫の擦り切れ
海砂利水魚の
水行末 雲来末 風来末
食う寝る処に住む処
藪ら柑子の藪柑子
パイポパイポ パイポのシューリンガン
シューリンガンのグーリンダイ
グーリンダイのポンポコピーのポンポコナーの
長久命の長助
`

// implements screen.EventDeque interface
type eventQueueStub struct{}

func (Q eventQueueStub) Send(event interface{}) {}

func (Q eventQueueStub) SendFirst(event interface{}) {}

func (Q eventQueueStub) NextEvent() interface{} { return nil }

func buildMultipleView() *MultipleView {
	presenter := NewEragoPresenter(eventQueueStub{})
	mv := NewMultipleView(presenter, DefaultTextViewOptions)
	Theme := &theme.Default
	mv.Measure(Theme, screenSize.X, screenSize.Y)
	mv.Rect = image.Rectangle{Max: screenSize}
	mv.Layout(Theme)
	return mv
}

func TestMultipleView(t *testing.T) {
	mv := buildMultipleView()
	defer mv.Close()
	Theme := &theme.Default

	const Text = "てすとだよーん"
	mv.Print(Text)
	if got := text.String(mv.viewManager.currentView().frame); strings.Compare(got, Text) != 0 {
		t.Errorf("diffrent frame content, got: %s, expect: %s", got, Text)
	}

	mv.Print(Jugem)
	mv.viewManager.currentView().Mark(node.MarkNeedsPaintBase)

	rgba := image.NewRGBA(image.Rectangle{Max: screenSize})
	if err := mv.PaintBase(&node.PaintBaseContext{
		Theme: Theme,
		Dst:   rgba,
	}, image.ZP); err != nil {
		t.Errorf("PaintBase: %v", err)
	}
	saveImage("_mv_test.png", rgba)
}

const imageDirectory = "./testimage"

func saveImage(name string, m image.Image) error {
	fp, err := os.Create(filepath.Join(imageDirectory, name))
	if err != nil {
		return err
	}
	defer fp.Close()

	return png.Encode(fp, m)
}

func TestMultipleViewSetLayout(t *testing.T) {
	mv := buildMultipleView()
	defer mv.Close()

	mv.Print(Jugem)

	if err := mv.SetLayout(
		attr.NewFlowHorizontal(
			attr.WithParentValue(attr.NewSingleText("left"), 1),
			attr.WithParentValue(attr.NewSingleText("right"), 1),
		),
	); err != nil {
		t.Fatal(err)
	}

	if vs := mv.viewManager.textViews; len(vs) != 2 {
		t.Fatalf("different view number, got: %d, expect: %d", len(vs), 2)
	}

	if got := mv.viewManager.currentView().name; got != "right" {
		t.Fatalf("different current view name, got: %v, expect: %v", got, "right")
	}

	mv.Print(Jugem)

	Theme := &theme.Default
	mv.Measure(Theme, screenSize.X, screenSize.Y)
	mv.Rect = image.Rectangle{Max: screenSize}
	mv.Layout(Theme)

	if err := mv.SetCurrentView("right"); err != nil {
		t.Error("can not change current view")
	}
	mv.Print("text to right")
}

func TestMultipleViewSetSingleLayout(t *testing.T) {
	mv := buildMultipleView()
	defer mv.Close()

	mv.Print(Jugem)

	newName := "new one"
	if err := mv.SetSingleLayout(newName); err != nil {
		t.Fatal(err)
	}

	if vs := mv.viewManager.textViews; len(vs) != 1 {
		t.Fatalf("different view number, got: %d, expect: %d", len(vs), 1)
	}

	if got := mv.viewManager.currentView().name; got != newName {
		t.Fatalf("different current view name, got: %v, expect: %v", got, newName)
	}

	mv.Print(Jugem)

	if err := mv.SetCurrentView(newName); err != nil {
		t.Error("can not change current view")
	}
	mv.Print(Jugem)

	// relayout by current view name itself.
	if err := mv.SetSingleLayout(newName); err != nil {
		t.Fatal(err)
	}

	if vs := mv.viewManager.textViews; len(vs) != 1 {
		t.Fatalf("different view number, got: %d, expect: %d", len(vs), 1)
	}

	if got := mv.viewManager.currentView().name; got != newName {
		t.Fatalf("different current view name, got: %v, expect: %v", got, newName)
	}

	mv.Print(Jugem)
}

func BenchmarkMultipleViewPrint(b *testing.B) {
	mv := buildMultipleView()
	defer mv.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mv.Print(Jugem)
	}
}

func BenchmarkMultipleViewSetLayout(b *testing.B) {
	mv := buildMultipleView()
	defer mv.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := mv.SetLayout(benchLayoutData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMultipleViewPaintBase(b *testing.B) {
	mv := buildMultipleView()
	defer mv.Close()
	Theme := &theme.Default

	mv.Print(Jugem)
	rgba := image.NewRGBA(image.Rectangle{Max: screenSize})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mv.viewManager.currentView().Mark(node.MarkNeedsPaintBase)
		if err := mv.PaintBase(&node.PaintBaseContext{
			Theme: Theme,
			Dst:   rgba,
		}, image.ZP); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMultipleViewPaintBaseRealScreenSize(b *testing.B) {
	mv := buildMultipleView()
	defer mv.Close()
	Theme := &theme.Default

	mv.Measure(Theme, realScreenSize.X, realScreenSize.Y)
	mv.Rect = image.Rectangle{Max: realScreenSize}
	mv.Layout(Theme)
	for i := 0; i < 10; i++ {
		mv.Print(Jugem)
	}

	rgba := image.NewRGBA(image.Rectangle{Max: realScreenSize})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mv.viewManager.currentView().Mark(node.MarkNeedsPaintBase)
		if err := mv.PaintBase(&node.PaintBaseContext{
			Theme: Theme,
			Dst:   rgba,
		}, image.ZP); err != nil {
			b.Fatal(err)
		}
	}
}
