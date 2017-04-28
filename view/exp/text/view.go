package text

import (
	"errors"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"local/erago/view/exp/theme"
)

// View is view of Frame, which controls
// how to show Frame to user and comminucates any user event for changing view state.
//
// View is constructed by Frame.NewView(), not NewView() or View{}.
type View struct {
	frame *Frame

	// size of drawing area.
	size image.Point

	face            font.Face
	faceHeight      fixed.Int26_6 // cache for pixel height of a character.
	faceAscent      fixed.Int26_6 // cache for pixel height of a ascent.
	faceDescent     fixed.Int26_6 // cache for pixel height of a descent.
	faceSingleWidth fixed.Int26_6 // cache for pixel width of a single byte character.

	// TODO: use faceHeight + inter-line-spacing instead of faceHeight?
	// lineHeight int32

	// start index of view pragraphs and lines.
	// startL is valid in lines of Paragraph startP.
	startP, startL int32

	// maximum start index of view paragraphs and lines.
	// The next of these may be break showing at least maxLines.
	maxStartP, maxStartL int32

	maxLines     int32 // max count of view lines.
	maxRuneWidth int   // max width in runewidth.

	buttons         []clickableButton // conserves clickable button list.
	highlightButton clickableButton   // conserves a clickable button to highlight.

	accumulatedMoveY float64 // accumulation of mouse moving Y.
}

// clickableButton holds its command and position to indicate clicking it.
type clickableButton struct {
	cmd      string
	position image.Rectangle
}

// return View to view frame content.
func (f *Frame) View() *View {
	return newView(f)
}

func newView(f *Frame) *View {
	return &View{
		frame: f,
	}
}

// Set font face which must be monospace.
// if font is not monospace return error.
// It changes Frame layout.
func (v *View) SetFace(face font.Face) error {
	// Assert that face is type monospace? and containing Japanese?
	x_adv, _ := face.GlyphAdvance('X')
	if i_adv, _ := face.GlyphAdvance('I'); x_adv != i_adv {
		return errors.New("Frame.View: font must be monospace")
	}

	ja_adv, ok := face.GlyphAdvance('あ')
	if !ok {
		return errors.New("Frame.View: font must support japanese")
	}
	// since rounding float point, testring with approximatry equal.
	if xx, ja := x_adv.Round()*2, ja_adv.Round(); xx > ja+1 || xx < ja-1 {
		return errors.New("Frame.View: width of singlebyte character must be half of that of multibyte character")
	}

	m := face.Metrics()
	v.face = face
	v.faceAscent = m.Ascent
	v.faceDescent = m.Descent
	v.faceHeight = v.faceAscent + v.faceDescent
	v.faceSingleWidth = x_adv

	v.relayout()
	return nil
}

// It returns max runewidth in the max view width,
// indicating how many charancters are in view's width.
// Note that single-byte character is counted as 1 and multibyte is typically as 2.
func (v *View) RuneWidth() int {
	return v.maxRuneWidth
}

// It returns a count of visuallized lines in the max view height.
func (v *View) LineCount() int {
	return int(v.maxLines)
}

// Set View size in pixel.
// It changes Frame layout.
func (v *View) SetSize(size image.Point) {
	v.size = size
	v.relayout()
}

func (v *View) relayout() {
	v.maxLines = v.toLineCount(v.size.Y)
	v.maxRuneWidth = v.toRuneWidth(v.size.X)

	v.frame.mu.Lock()
	defer v.frame.mu.Unlock()
	v.frame.SetMaxRuneWidth(v.maxRuneWidth)

	startP, startL := startPAndL(v.frame, v.maxLines)
	v.startP, v.startL = startP, startL
	v.maxStartP, v.maxStartL = startP, startL
}

// reference to fixed.Point26_6.Mul().
func int26_6_Mul(x, y fixed.Int26_6) fixed.Int26_6 {
	return x * y / 64
}

// reference to fixed.Point26_6.Div().
func int26_6_Div(x, y fixed.Int26_6) fixed.Int26_6 {
	return x * 64 / y
}

// convert pixel size to line count included in.
func (v *View) toLineCount(px int) int32 {
	if v.faceHeight == 0 {
		return 0
	}
	x := int26_6_Div(fixed.I(px), v.faceHeight).Floor()
	return int32(x)
}

// convert pixel size to runewidth included in.
func (v *View) toRuneWidth(px int) int {
	if v.faceSingleWidth == 0 {
		return 0
	}
	return int26_6_Div(fixed.I(px), v.faceSingleWidth).Floor()
}

// find command on position p in pixel.
// because stored commands are in phisical screen coordinate space,
// it requires point p in phisical screen coordinate space.
// return command and whether command is found?
func (v *View) FindCommand(p image.Point) (string, bool) {
	if cmd, found := v.findCommand(p); found {
		return cmd.cmd, true
	}
	return "", false
}

func (v *View) findCommand(p image.Point) (clickableButton, bool) {
	for _, b := range v.buttons {
		if p.In(b.position) {
			return b, true
		}
	}
	return clickableButton{}, false
}

// Highlight command which is selected on the position p
// because stored commands are in phisical screen coordinate space,
// it requires point p in phisical screen coordinate space.
// in View's coordinate space. return that command is found.
func (v *View) HighlightCommand(p image.Point) bool {
	if cmd, found := v.findCommand(p); found {
		v.highlightButton = cmd
		return true
	}
	return false
}

var emptyClickableButton = clickableButton{}

// Unhighlight all the command in View.
// return highlighted button exist and is Unhighlighted.
func (v *View) UnhighlightCommand() bool {
	exist := v.highlightButton != emptyClickableButton
	v.highlightButton = emptyClickableButton
	return exist
}

var (
	// Material White to draw text.
	DefaultForeColor = theme.DefaultPalette.Foreground().C.(color.RGBA)

	// Material LightBlue 400 to draw button text
	DefaultButtonColor = theme.DefaultPalette.Accent().C.(color.RGBA)

	defaultDrawSource = image.NewUniform(DefaultForeColor)

	// It is used to reset color setting.
	ResetColor = color.RGBA{}
)

var moreParent = image.NewUniform(color.RGBA{A: 0x33})

func (v *View) Draw(m *image.RGBA, origin image.Point) {
	drawRect := image.Rectangle{Max: v.size}.Add(origin)
	if drawRect.Empty() {
		return
	}

	v.drawBase(m, drawRect)

	// TODO: move to node.Paint?
	// highlight focused button.
	if hlButton := v.highlightButton; !hlButton.position.Eq(image.Rectangle{}) {
		hlColor := defaultDrawSource
		hlRect := hlButton.position
		draw.DrawMask(m, hlRect, hlColor, image.Point{}, moreParent, image.Point{}, draw.Over)
	}
}

func (v *View) drawBase(m *image.RGBA, drawRect image.Rectangle) {
	minX := fixed.I(drawRect.Min.X)
	minY := fixed.I(drawRect.Min.Y) + v.faceAscent  // adjusts base line
	maxY := fixed.I(drawRect.Max.Y) - v.faceDescent // adjusts base line

	defaultDrawSource.C = DefaultForeColor
	drawer := &font.Drawer{
		Dot:  fixed.Point26_6{X: minX, Y: minY},
		Face: v.face,
		Dst:  m,
		Src:  defaultDrawSource,
	}

	v.buttons = v.buttons[:0] // clear button list.

	f := v.frame
	f.mu.RLock()
	defer f.mu.RUnlock()

	scrolled := v.startP != v.maxStartP || v.startL != v.maxStartL
	{ // scroll to end of bottom if maxStartP or maxStartL is updated.
		// that case is occured when text content is updated or view size is updated.
		maxStartP, maxStartL := startPAndL(f, v.maxLines)
		if v.maxStartP != maxStartP || v.maxStartL != maxStartL {
			v.startP = maxStartP
			v.startL = maxStartL
			scrolled = false
			v.UnhighlightCommand()
		}
		v.maxStartP = maxStartP
		v.maxStartL = maxStartL
	}

	// draw some lines contained in first paragraph which starts not firstL.
	startPP := f.paragraph(v.startP)
	if startPP == nil {
		panic("Frame.View.Draw(): invalid state, missing startParagraph")
	}
	for l := f.line(v.startL); l != nil; l = l.Next(f) {
		if drawer.Dot.Y >= maxY {
			return
		}
		drawer.Dot.X = minX
		v.drawLine(drawer, l, startPP.align)
	}

	// draw trailing paragraphs which starts its content with firstL.
	for p := startPP.Next(f); p != nil; p = p.Next(f) {
		for l := p.FirstLine(f); l != nil; l = l.Next(f) {
			if drawer.Dot.Y > maxY {
				if !scrolled {
					panic("Frame.View.Draw(): not arriving at end of paragraph, on not scrolled")
				}
				return
			}
			drawer.Dot.X = minX
			v.drawLine(drawer, l, p.align)
		}
	}
}

// draw given line's content with alignment.
// after this, drawer's Dot is advanced with content height and width.
func (v *View) drawLine(drawer *font.Drawer, l *Line, align Alignment) {
	f := v.frame
	// add indent based on alignment.
	indent := fixed.I(0)
	switch align {
	case AlignmentLeft:
		// no add indent
	case AlignmentCenter:
		indentWidth := fixed.I((v.maxRuneWidth - l.RuneWidth(f)) / 2)
		indent = int26_6_Mul(v.faceSingleWidth, indentWidth)
	case AlignmentRight:
		indentWidth := fixed.I(v.maxRuneWidth - l.RuneWidth(f))
		indent = int26_6_Mul(v.faceSingleWidth, indentWidth)
	}
	drawer.Dot.X += indent

	for b := l.FirstBox(f); b != nil; b = b.Next(f) {
		b.Draw(drawer, v)
	}

	// TODO: add inter line space?
	drawer.Dot.Y += v.faceHeight
}

// extract start index of paragraph and line, next of which contain tail n lines.
func startPAndL(f *Frame, nlines int32) (int32, int32) {
	startP := f.lastParagraphIndex()
	lcount := int32(0)
	lcountInP := int32(0)

	// find start Paragraph index exceeding or containing eqaul to nlines.
	for p := startP; p != 0; {
		pp := &f.paragraphs[p]
		lcountInP = int32(pp.LineCount(f))
		lcount += lcountInP
		startP = p

		if lcount >= nlines {
			break
		}
		p = pp.prev
	}

	exceedingLines := (lcount - nlines)
	if exceedingLines == 0 {
		return startP, f.paragraphs[startP].firstL
	}

	blankLines := lcountInP - exceedingLines

	// find startL in startP paragraph to fill blank lines.
	startL := f.paragraphs[startP].lastLineIndex(f)
	lc := int32(0)

	for l := startL; l != 0; {
		ll := &f.lines[l]
		lc += 1
		startL = l
		if lc == blankLines {
			break
		}
		l = ll.prev
	}

	return startP, startL
}

// vertiacal scrolling of  view port of Frame contents.
// moveY is size of y move in pixel.
// positive moveY corresponds to scroll up and otherwise scroll down.
// 0 value corresponds to stop scroll.
func (v *View) Scroll(moveY int) {
	if moveY == 0 {
		v.accumulatedMoveY = 0
		return
	}

	// reset accurate moving if direction is different between current moving and accurated.
	if (moveY < 0 && v.accumulatedMoveY > 0) || (moveY > 0 && v.accumulatedMoveY < 0) {
		v.accumulatedMoveY = 0
	}

	mod := math.Mod(float64(moveY), float64(v.faceHeight))
	v.accumulatedMoveY += mod

	step := moveY / int(v.faceHeight)
	if faceH := float64(v.faceHeight); faceH <= math.Abs(v.accumulatedMoveY) {
		stepToAdd := int(v.accumulatedMoveY / faceH)
		step += stepToAdd
		v.accumulatedMoveY -= float64(stepToAdd)
	}
	v.ScrollLine(step)
}

// scroll n lines. positive n corresponds to scroll up, otherwise down.
func (v *View) ScrollLine(n int) {
	f := v.frame
	p := v.startP
	l := v.startL

	switch {
	case n > 0:
		if p == f.firstP && l == f.FirstParagraph().firstL {
			return
		}
		for ; n != 0; n -= 1 {
			ll := &f.lines[l]
			if prevL := ll.prev; prevL != 0 {
				v.startL = prevL
				l = prevL
				continue
			}

			// Since previous line is not found (=0) in this paragraph,
			// then move to previous paragraph,
			prevP := f.paragraphs[p].prev
			if prevP == 0 {
				return // can not scroll because p and l is most first index.
			}
			v.startP, v.startL = prevP, f.paragraphs[prevP].lastLineIndex(f)
			p = prevP
		}
	case n < 0:
		if p == v.maxStartP && l == v.maxStartL {
			return
		}
		for ; n != 0; n += 1 {
			ll := &f.lines[l]
			if nextL := ll.next; nextL != 0 {
				v.startL = nextL
				l = nextL
				continue
			}

			nextP := f.paragraphs[p].next
			if nextP == 0 || p == v.maxStartP {
				return // can not scroll because p and l is most last index.
			}
			v.startP, v.startL = nextP, f.paragraphs[nextP].firstL
			p = nextP
		}
	case n == 0:
		// do nothing
	}
}

// return string cotent that will be shown by Draw().
func DrawingString(v *View) string {
	f := v.frame
	buf := make([]byte, 0, f.text_len)

	appendBoxText := func(l *Line) {
		for b := l.FirstBox(f); b != nil; b = b.Next(f) {
			if tb, ok := b.(TextBox); ok {
				buf = append(buf, tb.Bytes(f)...)
			}
		}
	}
	appendReturnCode := func() {
		buf = append(buf, byte('\n'))
	}

	extractLinesFromStartPAndL(v, func(l *Line) {
		appendBoxText(l)
		appendReturnCode()
	})
	return string(buf)
}

func extractLinesFromStartPAndL(v *View, handler func(l *Line)) {
	f := v.frame
	for l := f.line(v.startL); l != nil; l = l.Next(f) {
		handler(l)
	}
	for p := f.paragraph(v.startP).Next(f); p != nil; p = p.Next(f) {
		for l := p.FirstLine(f); l != nil; l = l.Next(f) {
			handler(l)
		}
	}
}
