package text

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	gfont "golang.org/x/image/font"

	"local/erago/view/exp/theme"
	gwidth "local/erago/width"
)

func TestWrite(t *testing.T) {
	const (
		line1 = "First, Red Onion\n"
		line2 = "Second, Blue Calot\n"
		box3  = "Third,"
		box4  = "Yellow Pickle"
	)
	f := NewFrame(nil)
	e := f.Editor()
	defer e.Close()

	nbyte, err := e.WriteText(line1)
	if err != nil {
		t.Fatal(err)
	}
	if len := len(line1); nbyte != len {
		t.Errorf("different wrote bytes and passed text: bytes %d, text: %d", nbyte, len)
	}

	e.WriteText(line2)
	e.WriteText(box3)
	e.WriteText(box4)

	for _, test := range []struct {
		delN   int
		expect string
	}{
		{0, line1 + line2 + box3 + box4},
		{1, line1 + line2},
		{2, line1},
		{2, ""},
		{1, ""},
	} {
		e.DeleteLastParagraphs(test.delN)
		if read := string(readAllBytes(f)); read != test.expect {
			t.Errorf("different wrote contents,\n  got: %s,\n  expect: %s", read, test.expect)
		}
	}
}

func TestWriteText(t *testing.T) {
	f := NewFrame(nil)
	f.SetMaxRuneWidth(30)
	e := f.Editor()
	defer e.Close()

	nBytes, _ := e.WriteText("abc")
	if rw := e.currentLine().RuneWidth(f); rw != nBytes {
		t.Errorf("different current line width and writing text's width, got: %v, expect: %v", rw, nBytes)
	}
}

func TestWriteLine(t *testing.T) {
	f := NewFrame(nil)
	f.SetMaxRuneWidth(30)
	e := f.Editor()
	defer e.Close()

	e.WriteLine("=")
	if pcount := f.ParagraphCount(); pcount != 2 {
		t.Fatalf("invalid Paragraph count, expect: %d, got: %d", 2, pcount)
	}

	e.DeleteLastParagraphs(2) // delete all
	if pcount := f.ParagraphCount(); pcount != 1 {
		t.Fatalf("invalid Paragraph count, expect: %d, got: %d", 1, pcount)
	}

	e.WriteText("aa")
	e.WriteLine("=")
	if pcount := f.ParagraphCount(); pcount != 3 {
		t.Fatalf("invalid Paragraph count, expect: %d, got: %d", 3, pcount)
	}
}

func TestDeleteFirstParagraph(t *testing.T) {
	const (
		line1 = "aaaa\n"
		line2 = "iii\n"
		line3 = "uu\n"
		line4 = "e"
	)
	f := NewFrame(nil)
	e := f.Editor()
	defer e.Close()

	for _, s := range []string{
		line1, line2, line3, line4,
	} {
		_, err := e.WriteText(s)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, expect := range []int32{
		2, 3, 4,
	} {
		e.deleteFirstParagraphs(1)
		if f.firstP != expect {
			t.Errorf("different firstP, got: %d, expect: %v", f.firstP, expect)
			t.Log(string(readAllBytes(f)))
		}
	}
}

func readAllBytes(f *Frame) []byte {
	buf := make([]byte, 0, f.Len())
	for p := f.FirstParagraph(); p != nil; p = p.Next(f) {
		for l := p.FirstLine(f); l != nil; l = l.Next(f) {
			for b := l.FirstBox(f); b != nil; b = b.Next(f) {
				if tbox, ok := b.(TextBox); ok {
					buf = append(buf, tbox.Text(f)...)
				}
			}
		}
	}
	return buf
}

func TestTextLen(t *testing.T) {
	f := NewFrame(&FrameOptions{
		MaxParagraphs:     2,
		MaxParagraphBytes: 100,
	})
	e := f.Editor()
	defer e.Close()

	checkTextLen := func(expect int32) {
		if f.text_len != expect {
			t.Errorf("different text length, got: %d, expect: %d", f.text_len, expect)
		}
		t.Logf("frame content:\n%#v", String(f))
	}

	const bytes10 = "abcdefghij"

	e.WriteText(bytes10 + "\n")
	e.WriteText(bytes10)
	checkTextLen(21)

	e.WriteText("\n") // first paragraph is removed.
	checkTextLen(11)

	e.WriteButton(bytes10+"klmno", "10")
	e.WriteText("\n") // first paragraph is removed.
	checkTextLen(16)

	f.compactText()

	checkTextLen(16)

	e.WriteLine("=") // => \n========\n
	checkTextLen(0)

	e.WriteText(bytes10) // => ======\n(bytes10)
	e.WriteLine("=")     // => (bytes10)\n========\n
	checkTextLen(0)
}

func TestWriteManyTimes(t *testing.T) {
	f := NewFrame(&FrameOptions{
		MaxParagraphs:     10,
		MaxParagraphBytes: 100,
	})
	e := f.Editor()
	defer e.Close()

	for i := 0; i < 100; i++ {
		e.WriteText(Jugem)
		e.WriteButton("button", "10")
		e.WriteLine("=")
	}
}

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

func BenchmarkExtractFrameText(b *testing.B) {
	f := NewFrame(nil)
	e := f.Editor()
	defer e.Close()
	e.WriteText(strings.Repeat(Jugem, 4))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readAllBytes(f)
	}
}

func BenchmarkWriteSmallText(b *testing.B) {
	f := NewFrame(nil)
	text := "こんにちは、世界\n"
	e := f.Editor()
	defer e.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.WriteText(text)
	}
}

func BenchmarkWriteLargeText(b *testing.B) {
	f := NewFrame(nil)
	text := strings.Repeat(Jugem, 4)
	e := f.Editor()
	defer e.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.WriteText(text)
	}
}

func BenchmarkWriteLargeWidthText(b *testing.B) {
	f := NewFrame(nil)
	text := strings.Replace(strings.Repeat(Jugem, 4), "\n", "", -1)
	e := f.Editor()
	defer e.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.WriteText(text)
	}
}

func BenchmarkWriteAndDeleteText(b *testing.B) {
	f := NewFrame(nil)
	text := Jugem
	e := f.Editor()
	defer e.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.WriteText(text)
		e.DeleteLastParagraphs(9)
	}
}

func BenchmarkMeasureRuneWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runewidth.StringWidth(Jugem)
	}
}

func BenchmarkMeasureGoWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gwidth.StringWidth(Jugem)
	}
}

func BenchmarkMeasureFontAdvance(b *testing.B) {
	face := theme.NewDefaultFace(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gfont.MeasureString(face, Jugem)
	}
}
