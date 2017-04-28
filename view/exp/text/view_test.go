package text

import (
	"image"
	"image/png"
	"os"
	"strings"
	"testing"

	"local/erago/view/exp/theme"
)

func TestView(t *testing.T) {
	const (
		W = 400
		H = 300
	)

	f := NewFrame(nil)
	editor := f.Editor()
	defer editor.Close()
	_, err := editor.WriteText(strings.Repeat(Jugem, 4))
	if err != nil {
		t.Fatal(err)
	}
	v := f.View()
	v.SetSize(image.Point{W, H})

	face := theme.NewDefaultFace(nil)
	v.SetFace(face)

	rgba := image.NewRGBA(image.Rect(0, 0, W, H))
	v.Draw(rgba, image.ZP)
	// if err := saveImage(rgba); err != nil {
	// 	t.Error(err)
	// }
}

func drawViewState(v *View) error {
	rgba := image.NewRGBA(image.Rectangle{Max: v.size})
	v.Draw(rgba, image.ZP)
	return saveImage(rgba)
}

func saveImage(m image.Image) error {
	fp, err := os.Create("_test.png")
	if err != nil {
		return err
	}
	defer fp.Close()

	return png.Encode(fp, m)
}

func TestViewStartPAndL(t *testing.T) {
	f := NewFrame(nil)
	editor := f.Editor()
	defer editor.Close()
	editor.WriteText(strings.Repeat(Jugem, 10))

	v := f.View()
	v.SetSize(image.Point{256, 128})
	v.SetFace(theme.NewDefaultFace(nil))

	startP, startL := startPAndL(f, v.maxLines)
	v.startP, v.startL = startP, startL
	lineCount := func() int {
		lcount := 0
		extractLinesFromStartPAndL(v, func(l *Line) {
			lcount += 1
		})
		return lcount
	}

	lcount := lineCount()
	if expect := int(v.maxLines); lcount != expect {
		t.Errorf("diffrent line count, got: %d, expect: %d", lcount, expect)
	}

	// after change size
	v.SetSize(image.Pt(64, 128))
	startP, startL = startPAndL(f, v.maxLines)
	v.startP, v.startL = startP, startL
	lcount = lineCount()
	if expect := int(v.maxLines); lcount != expect {
		t.Errorf("diffrent line count, got: %d, expect: %d", lcount, expect)
	}

	// after write new text
	editor.WriteText(Jugem)
	startP, startL = startPAndL(f, v.maxLines)
	v.startP, v.startL = startP, startL
	lcount = lineCount()
	if expect := int(v.maxLines); lcount != expect {
		t.Errorf("diffrent line count, got: %d, expect: %d", lcount, expect)
	}
}

func TestDraw(t *testing.T) {
	f := NewFrame(nil)
	editor := f.Editor()
	defer editor.Close()

	v := f.View()
	v.SetSize(image.Point{64, 256})
	v.SetFace(theme.NewDefaultFace(nil))

	const (
		showIndex   = 38
		repeatTimes = 100
	)
	for i := 0; i < repeatTimes; i++ {
		editor.WriteText(Jugem)
		editor.WriteLabel("Label-Text-YEY!!")
		editor.WriteButton("[] command", "0")
		editor.WriteLine("=")
		if i == showIndex {
			// drawViewState(v)
		}
	}
}

func TestViewScroll(t *testing.T) {
	f := NewFrame(nil)

	text := strings.Repeat(Jugem, 4)
	text_nline := 2 + strings.Count(text, "\n")
	editor := f.Editor()
	defer editor.Close()
	_, err := editor.WriteText(text)
	if err != nil {
		t.Fatal(err)
	}

	v := f.View()
	v.SetFace(theme.NewDefaultFace(nil))

	const H = 400
	v.SetSize(image.Point{800, H})
	view_nline := int(v.toLineCount(H))

	startP := int32(text_nline - view_nline)
	if startP != v.startP {
		t.Logf("view line: %d, text line: %d", view_nline, text_nline)
		t.Fatalf("startP is different: got %d, expect %d", startP, v.startP)
	}

	v.Scroll(int(v.faceHeight * 2)) // scroll up 2 lines

	// appending Paragraph is slice's order such as
	// paragraphs[0] is 0'th line and paragraphs[1] is 1st line.
	if scrolledP := startP - 2; scrolledP != v.startP {
		t.Fatalf("startP is different: got %d, expect %d", v.startP, scrolledP)
	}
}

func BenchmarkDraw(b *testing.B) {
	const (
		W = 1000
		H = 800
	)

	f := NewFrame(nil)
	e := f.Editor()
	defer e.Close()
	e.WriteText(strings.Repeat(Jugem, 4))
	e.WriteButton("[] button text", "0")
	e.WriteLine("=")

	rgba := image.NewRGBA(image.Rect(0, 0, W, H))

	v := f.View()
	v.SetFace(theme.NewDefaultFace(nil))
	v.SetSize(rgba.Bounds().Size())
	// drawViewState(v)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Draw(rgba, image.ZP)
	}
}
