package ui

import (
	"image"
	"testing"

	"golang.org/x/exp/shiny/widget/node"

	"local/erago/view/exp/theme"
)

var wideScreenSize = image.Point{512, 256}

func buildImageView(file string) *ImageView {
	v := NewImageView(file)
	Theme := &theme.Default
	v.Measure(Theme, wideScreenSize.X, wideScreenSize.Y)
	v.Rect = image.Rectangle{Max: v.MeasuredSize}
	v.Layout(Theme)
	return v
}

var nothingImageView = buildImageView("nothing")

func TestImageViewNothing(t *testing.T) {
	v := nothingImageView
	Theme := &theme.Default
	rgba := image.NewRGBA(image.Rectangle{Max: wideScreenSize})

	v.Mark(node.MarkNeedsPaintBase)
	if err := v.PaintBase(&node.PaintBaseContext{
		Theme: Theme,
		Dst:   rgba,
	}, image.ZP); err != nil {
		t.Errorf("PaintBase: %v", err)
	}
	saveImage("_img_error_test.png", rgba)
}

func TestImageViewExist(t *testing.T) {
	v := buildImageView("./testimage/_img_error_test.png")
	Theme := &theme.Default
	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: screenSize}
	v.Layout(Theme)
	rgba := image.NewRGBA(v.Rect)

	v.Mark(node.MarkNeedsPaintBase)
	if err := v.PaintBase(&node.PaintBaseContext{
		Theme: Theme,
		Dst:   rgba,
	}, image.ZP); err != nil {
		t.Errorf("PaintBase: %v", err)
	}
	saveImage("_img_test.png", rgba)
}

func BenchmarkImageViewPaintBase(b *testing.B) {
	v := nothingImageView
	Theme := &theme.Default
	v.Measure(Theme, screenSize.X, screenSize.Y)
	v.Rect = image.Rectangle{Max: screenSize}
	v.Layout(Theme)
	rgba := image.NewRGBA(v.Rect)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.PaintBase(&node.PaintBaseContext{
			Theme: Theme,
			Dst:   rgba,
		}, image.Point{}); err != nil {
			b.Fatal(err)
		}
		v.Mark(node.MarkNeedsPaintBase)
	}
}

func BenchmarkImageViewPaintBaseRealSize(b *testing.B) {
	v := nothingImageView
	Theme := &theme.Default
	v.Measure(Theme, realScreenSize.X, realScreenSize.Y)
	v.Rect = image.Rectangle{Max: realScreenSize}
	v.Layout(Theme)
	rgba := image.NewRGBA(v.Rect)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.PaintBase(&node.PaintBaseContext{
			Theme: Theme,
			Dst:   rgba,
		}, image.Point{}); err != nil {
			b.Fatal(err)
		}
		v.Mark(node.MarkNeedsPaintBase)
	}
}
