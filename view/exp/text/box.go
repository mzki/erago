package text

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Box is abstract content. It is used in Frame internally,
// so you need not to implements it.
// The exported methods can be used to inspect it.
type Box interface {
	RuneWidth(*Frame) int     // return box's width in runewidth.
	LineCountHint(*Frame) int // return box's line count hint, which is used optionally.

	// return next Box, if not found return nil
	Next(*Frame) Box
	Prev(*Frame) Box

	// return next index, if not found retuen 0
	nextIndex() int32
	prevIndex() int32

	setNextIndex(int32)
	setPrevIndex(int32)

	// draw its contents into font.Drawer.Dst.
	Draw(*font.Drawer, *View)
}

// implements Box interface
type baseBox struct {
	next, prev int32
}

func (b baseBox) Width(*Frame) int         { return 0 }
func (b baseBox) RuneWidth(*Frame) int     { return 0 }
func (b baseBox) LineCountHint(*Frame) int { return 1 }

func (b baseBox) Next(f *Frame) Box {
	return f.box(b.next)
}

func (b baseBox) Prev(f *Frame) Box {
	return f.box(b.prev)
}

func (b baseBox) nextIndex() int32      { return b.next }
func (b baseBox) prevIndex() int32      { return b.prev }
func (b *baseBox) setNextIndex(i int32) { b.next = i }
func (b *baseBox) setPrevIndex(i int32) { b.prev = i }

func (b baseBox) Draw(d *font.Drawer, v *View) {
	// do nothing
}

// IJBox is containing index of text: i, j.
// Indeed, it has some texts.
type ijBox interface {
	Box

	I() int32
	J() int32
	setI(int32)
	setJ(int32)

	// set text using TextBox's i, j, and so on.
	setText(tbox textBox)
}

// TextBox holds text and can show it for user.
type TextBox interface {
	ijBox

	Text(*Frame) string
	Bytes(*Frame) []byte
	FgColor() color.RGBA
}

// normal text. Only TextBox can be split with its contents.
type textBox struct {
	baseBox

	// rune width is total width of its string,
	// which measures as multibyte: 2, singlebyte: 1.
	runewidth int

	i, j int32

	// color represents 32bit RGBA, having 8bits for each of
	// Red, Green, Blue, Alpha.
	color color.RGBA
}

// return RuneWidth of its text.
func (tb textBox) RuneWidth(f *Frame) int { return tb.runewidth }

// implements IJBox interface.
func (tb textBox) I() int32      { return tb.i }
func (tb textBox) J() int32      { return tb.j }
func (tb *textBox) setI(i int32) { tb.i = i }
func (tb *textBox) setJ(j int32) { tb.j = j }

// implements TextBox interface.
func (tb textBox) Text(f *Frame) string {
	return string(f.text[tb.i:tb.j])
}

// implements TextBox interface.
func (tb *textBox) setText(tbox textBox) {
	tb.i = tbox.i
	tb.j = tbox.j
	tb.color = tbox.color
	tb.runewidth = tbox.runewidth
}

// return text content in bytes.
func (tb textBox) Bytes(f *Frame) []byte {
	return f.text[tb.i:tb.j]
}

// retuen bytes trimmed last of "\n" or "\r".
func (tb textBox) TrimmedBytes(f *Frame) []byte {
	s := f.text[tb.i:tb.j]
	if tb.next == 0 {
		// in utf8 (or ascii), a byte smaller than or eqaul to white space is non-character, then trimming it.
		for len(s) > 0 && s[len(s)-1] <= ' ' {
			s = s[:len(s)-1]
		}
	}
	return s
}

func (tb *textBox) FgColor() color.RGBA { return tb.color }

// implements Box interface.
func (tb *textBox) Draw(d *font.Drawer, v *View) {
	if tb.color.A == 0 {
		// set default color to avoid complete transparent.
		tb.color = DefaultForeColor
	}
	d.Src.(*image.Uniform).C = tb.color
	d.DrawBytes(tb.TrimmedBytes(v.frame))
}

// label box is same as textBox but non-splitable,
// which means labelBox is always move to entire its content.
type labelBox struct {
	textBox
}

type ButtonBox interface {
	TextBox
	Command() string
}

// clickable button which emits a command,
// is represented as like: [cmd] caption.
type buttonBox struct {
	textBox
	cmd string
}

func (bb *buttonBox) Draw(d *font.Drawer, v *View) {
	if bb.color.A == 0 {
		bb.color = DefaultButtonColor
	}

	x0 := d.Dot.X.Floor()

	d.Src.(*image.Uniform).C = bb.color
	d.DrawBytes(bb.Bytes(v.frame))

	x1 := d.Dot.X.Ceil()
	y0 := (d.Dot.Y - v.faceAscent).Floor()
	y1 := (d.Dot.Y + v.faceDescent).Ceil()

	bbPos := image.Rect(x0, y0, x1, y1)

	// TODO: easily find position is what? Rectanle? runewidth based?
	v.buttons = append(v.buttons, clickableButton{
		cmd:      bb.cmd,
		position: bbPos,
	})
}

func (bb *buttonBox) Command() string { return bb.cmd }

type LineBox interface {
	Text(*Frame) string
	Symbol() string
	FgColor() color.RGBA
}

// text line as like: ==========================
type lineBox struct {
	baseBox
	color     color.RGBA
	symbol    string
	runewidth int
}

// lineBox always place at end of Line.
func (l lineBox) nextIndex() int32 { return 0 }

func (l lineBox) RuneWidth(f *Frame) int {
	return f.maxRuneWidth // lineBox always lays in full of frame width.
}

func (l *lineBox) Draw(d *font.Drawer, v *View) {
	if l.color.A == 0 {
		l.color = DefaultForeColor
	}
	d.Src.(*image.Uniform).C = l.color

	var repeat int
	numerator := fixed.I(v.size.X) - d.Dot.X
	denominator := int26_6_Mul(fixed.I(l.runewidth), v.faceSingleWidth)
	if numerator < 0 || denominator == 0 {
		repeat = 0
	} else {
		repeat = int26_6_Div(numerator, denominator).Floor()
	}
	d.DrawString(strings.Repeat(l.symbol, repeat))
}

func (l *lineBox) Text(f *Frame) string {
	n := f.maxRuneWidth / l.runewidth
	txtLine := strings.Repeat(l.symbol, n)
	return txtLine
}

func (l *lineBox) Symbol() string      { return l.symbol }
func (l *lineBox) FgColor() color.RGBA { return l.color }

// ImageBox holds image source and TextBox facility.
type ImageBox interface {
	// ImageBox's text content may be a debug information.
	// RuneWidth and LineCountHint shows image size in text scale.
	TextBox
	// SourceImage returns source path for image content. The byte
	// content of image is not included.
	SourceImage() string
}

// image box. It often has LineCountHint() > 2. This means
// it should be treated as special case.
type imageBox struct {
	labelBox

	src           string // source image path
	dstWidthInRW  int    // in rune width
	dstHeightInLC int    // in line count
	// TODO: Add more image option, which may be in attribute package?
}

func (box *imageBox) RuneWidth(*Frame) int     { return box.dstWidthInRW }
func (box *imageBox) LineCountHint(*Frame) int { return box.dstHeightInLC }
func (box *imageBox) SourceImage() string      { return box.src }

func (box *imageBox) Draw(d *font.Drawer, v *View) {
	toImagePoint := func(fp fixed.Point26_6) image.Point {
		return image.Point{
			X: fp.X.Round(),
			Y: fp.Y.Round(),
		}
	}

	// d.Dot points text baseline but image should starts with left top of blank space.
	leftTopDot := d.Dot.Add(fixed.Point26_6{X: 0, Y: -v.faceAscent})
	srcImg, fpSize, _ := v.GetImage(box.src, box.dstWidthInRW, box.dstHeightInLC)
	r := image.Rectangle{
		Min: toImagePoint(leftTopDot),
		Max: toImagePoint(leftTopDot.Add(fpSize)),
	}
	draw.Draw(d.Dst, r, srcImg, image.Point{}, draw.Over)

	// advance d.Dot.X by width of image drawn region
	d.Dot.X += fpSize.X
}

// space box is alsmost same as labelBox, is non splitable,
// but does not contain any content.
// It is used for spacing with runeWidth and lineCount=1 for the area and
// used for complex layout. The difference from lableBox with multile space
// character " " is that there is no drawing at the area while the labelBox
// with space draws space characters, means filling the area by font color.
// For example, suppose [Space] is space box and others are text box then
// view area will be shown as like below.
// >> Some text...
// >> [Space]Indented Some text...
// >> Some text...[Space]Trailing text...
type SpaceBox interface {
	TextBox

	SpaceRuneWidth() int
}

type spaceBox struct {
	textBox
	spaceRuneWidth int
}

func (sb *spaceBox) RuneWidth(*Frame) int { return sb.SpaceRuneWidth() }
func (sb *spaceBox) SpaceRuneWidth() int  { return sb.spaceRuneWidth }

func (sb *spaceBox) Draw(f *font.Drawer, v *View) {
	// space box write no content.
	// just move toward right so that next box can write next to this box.
	advanceX := int26_6_Mul(fixed.I(sb.RuneWidth(v.frame)), v.faceSingleWidth)
	f.Dot.X += advanceX
}
