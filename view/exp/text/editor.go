package text

import (
	"errors"
	"fmt"
	"image/color"
	"path"
	"strings"
	"unicode/utf8"

	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/width"
)

// Alignment of current line.
// For convenience to not import attribute package, it is same as attribute.Alignment.
type Alignment int8

const (
	AlignmentLeft   = Alignment(attr.AlignmentLeft)
	AlignmentCenter = Alignment(attr.AlignmentCenter)
	AlignmentRight  = Alignment(attr.AlignmentRight)
)

func (a Alignment) String() string {
	switch a {
	case AlignmentLeft:
		return "left"
	case AlignmentCenter:
		return "center"
	case AlignmentRight:
		return "right"
	default:
		return "unknown"
	}
}

// Editor edits Frame's contents.
// It only appends new content into Frame's last.
//
// It is constructed by Frame.Editor(),
// not NewEditor() or Editor{}
//
// Multiple Editors do not share their states.
type Editor struct {
	frame *Frame

	// current attribute for text.
	// it is used to write text.
	align Alignment
	Color color.RGBA

	// cache for index of last paragraph, line and box to append text.
	lastP, lastL, lastB int32

	// count for outputting new line
	newLineCount int
}

// return Editor to edit frame content.
func (f *Frame) Editor() *Editor {
	if f.editor == nil {
		f.editor = newEditor(f)
	}
	return f.editor
}

func newEditor(f *Frame) *Editor {
	e := &Editor{
		frame: f,
		align: AlignmentLeft,
		Color: ResetColor,
	}
	e.invalidateLasts()
	return e
}

// close this editor. after this reference for Frame is invalid,
// so calling any Editor's method will occur panic for nil reference.
func (e *Editor) Close() {
	e.frame.mu.Lock()
	e.frame.editor = nil
	e.frame.mu.Unlock()
	e.frame = nil
}

// set Alignment to the editor so that writing text is aligned.
func (e *Editor) SetAlignment(a Alignment) {
	e.align = a

	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	if e.isInvalidated() {
		e.setLasts()
	}
	pp := &e.frame.paragraphs[e.lastP]
	pp.align = a
}

// get current Alignment of Paragraph.
func (e *Editor) GetAlignment() Alignment {
	// The frame is not changed so not need mutex lock.
	return e.align
}

func (e *Editor) isInvalidated() bool {
	return e.lastP == 0 || e.lastL == 0 || e.lastB == 0
}

func (e *Editor) invalidateLasts() {
	e.lastP = 0
	e.lastL = 0
	e.lastB = 0
}

func (e *Editor) setLasts() {
	f := e.frame
	e.lastP = f.lastParagraphIndex()
	e.lastL = f.paragraphs[e.lastP].lastLineIndex(f)
	e.lastB = f.lines[e.lastL].lastBoxIndex(f)
}

// It returns runewidth in current editing Paragraph and Line, which must be last of Frame.
// Returned width 0 indicates that this Paragraph/Line have empty contents,
// otherwise have some content in that width.
func (e *Editor) CurrentRuneWidth() int {
	f := e.frame
	f.mu.Lock()
	defer f.mu.Unlock()

	return e.currentLine().RuneWidth(f)
}

// It returns count for outputting new line.
// return 0 at initial state.
func (e *Editor) NewLineCount() int {
	// The frame is not changed so not need mutex lock.
	return e.newLineCount
}

// get current editing paragraph.
func (e *Editor) currentParagragh() *Paragraph {
	if e.isInvalidated() {
		e.setLasts()
	}
	return &e.frame.paragraphs[e.lastP]
}

// get current editing line.
func (e *Editor) currentLine() *Line {
	if e.isInvalidated() {
		e.setLasts()
	}
	return &e.frame.lines[e.lastL]
}

// write text into end of frame's text.
func (e *Editor) WriteText(s string) (n int, err error) {
	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	pcount := e.frame.ParagraphCount()
	for len(s) > 0 {
		pcount += 1
		si := 1 + strings.Index(s, "\n")
		if si == 0 {
			si = len(s)
			pcount -= 1 // decrease so that paragraph count is not changed
		}
		if err = e.appendText(s[:si], &textBox{}); err != nil {
			return
		}

		n += si
		s = s[si:]
	}

	e.truncateParagraphs(pcount)
	return
}

// exceeding paragraphs are truncated.
func (e *Editor) truncateParagraphs(pcount int) {
	maxPCount := int(e.frame.maxParagraphs)
	if over_n := pcount - maxPCount; over_n > 0 {
		e.deleteFirstParagraphs(over_n)
	}
}

// write text as label into end of frame's text.
// The lable text is truncated if including "\n"
// It is not splitable to remain continuity of content for showing screen.
func (e *Editor) WriteLabel(text string) (n int, err error) {
	i := strings.Index(text, "\n")
	if i == -1 {
		i = len(text)
	}

	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	err = e.appendText(text[:i], &labelBox{})
	n = i
	return
}

// write text as button into end of frame's text.
// button text does not accept "\n".
// It is not splitable as same as label.
func (e *Editor) WriteButton(text, cmd string) (n int, err error) {
	i := strings.Index(text, "\n")
	if i == -1 {
		i = len(text)
	}

	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	err = e.appendText(text[:i], &buttonBox{cmd: cmd})
	n = i
	return
}

// WriteImage writes image into text frame with size of widthInRW x heightInLC.
// widthInRW stands for width in Rune Width and stands for height in Line Count.
func (e *Editor) WriteImage(imgFile string, widthInRW, heightInLC int) (err error) {
	if widthInRW <= 0 {
		return fmt.Errorf("widthInRW must be greater than zero, but %d", widthInRW)
	}

	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	imgText := fmt.Sprintf("<img src=%q width=%d height=%d >", path.Base(imgFile), widthInRW, heightInLC)
	err = e.appendText(imgText, &imageBox{
		src:           imgFile,
		dstWidthInRW:  widthInRW,
		dstHeightInLC: heightInLC,
	})
	return
}

// It appends text into frame's text buffer
// then returns whether text is appended?
// Passed ijbox is used to hold appended text.
//
// Given text should end with "\n" or contain no "\n".
// If given text ends with "\n" then new line is allocated and
// editor's lastL is changed.
func (e *Editor) appendText(s string, ijbox ijBox) error {
	if rest := maxTextLen - len(e.frame.text); len(s) > rest {
		return errors.New("Editor: insufficient space for writing text")
	}

	f := e.frame
	if e.isInvalidated() {
		e.setLasts()
	}

	// append text buffer
	length := int32(len(s))
	i := int32(len(f.text))
	j := i + length
	f.text_len += length
	f.text = append(f.text, s...)

	{ // add a box object
		ijbox.setText(textBox{
			color:     e.Color,
			i:         i,
			j:         j,
			runewidth: width.StringWidth(s),
		})
		e.appendBox(ijbox)
	}

	// layouting this line if currentRuneWidth exceeds frame.runewidth,
	if rest := f.maxRuneWidth - e.currentLine().RuneWidth(f); rest < 0 {
		layoutOneLine(f, e.lastL)
		e.setLasts()
		e.currentParagragh().invalidateCaches()
	}

	// Hard return. Layout last appended paragraph then,
	// insert new paragraph to after the last paragraph.
	if s[length-1] == '\n' {
		e.appendParagraph()
	}
	return nil
}

// append new Box bb into after last box.
// after that e.lastB is updated.
// it assumes e.lastB is valid.
func (e *Editor) appendBox(bb Box) {
	f := e.frame
	newB, _ := f.addBox(bb)

	// join previous and new boxes
	lastBB, newBB := f.boxes[e.lastB], f.boxes[newB]
	joined := f.joinBoxes(e.lastB, newB, lastBB, newBB)
	if !joined {
		// link newB and lastB
		newBB.setPrevIndex(e.lastB)
		lastBB.setNextIndex(newB)
		e.lastB = newB
	}
}

// append new paragraph into after last paragraph.
// it assumes e.lastP is valid before running this.
// after appending, lastP, lastL and lastB is valid.
func (e *Editor) appendParagraph() {
	f := e.frame

	// new paragraph is inserted to after last paragraph,
	// so not needing to consider that next paragraph is
	// already exist?
	newP, _ := e.newParagraph()
	newPP := &f.paragraphs[newP]
	newPP.prev = e.lastP
	f.paragraphs[e.lastP].next = newP
	e.lastP = newP

	newL, _ := f.newLine()
	newLL := &f.lines[newL]
	newPP.firstL = newL
	e.lastL = newL

	newB, _ := f.addBox(&textBox{})
	newLL.firstB = newB
	e.lastB = newB

	e.newLineCount++

	f.invalidateCaches()
}

// allocalte new paragraph which contains editor's align.
func (e Editor) newParagraph() (int32, bool) {
	p, realloc := e.frame.newParagraph()
	e.frame.paragraphs[p].align = e.align
	return p, realloc
}

// Write line into end of frame.
// The line must place end of frame's line.
// That is, after this, writing frame's line moves to next.
func (e *Editor) WriteLine(sym string) (n int, err error) {
	f := e.frame
	f.mu.Lock()
	defer f.mu.Unlock()

	if e.isInvalidated() {
		e.setLasts()
	}
	// force move to next if current line has any content.
	if e.currentLine().RuneWidth(f) > 0 {
		e.appendParagraph()
	}

	e.appendBox(&lineBox{
		color:     e.Color,
		symbol:    sym,
		runewidth: width.StringWidth(sym),
	})

	// must end paragraph with lineBox.
	e.appendParagraph()

	// exceeding paragraphs are truncated
	e.truncateParagraphs(f.ParagraphCount())
	return 0, nil
}

// delete n first paragraphs of Frame
func (e *Editor) deleteFirstParagraphs(n int) {
	if maxn := e.frame.ParagraphCount(); maxn < n {
		n = maxn
	}

	f := e.frame
	to_remove := f.firstP
	for i := 0; i < n; i++ {
		next_p := e.deleteParagraph(to_remove)
		f.firstP = next_p
		to_remove = next_p
	}
}

// delete n-1 last paragraphs of Frame and clear contents of 1 current line.
// if n is less than or equal to 0, do nothing..
func (e *Editor) DeleteLastParagraphs(n int) {
	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()

	if maxn := e.frame.ParagraphCount(); maxn < n {
		n = maxn
	}
	if n <= 0 {
		return
	}

	to_remove := e.frame.lastParagraphIndex()
	for i := 0; i < n; i++ {
		prev_p := e.frame.paragraphs[to_remove].prev
		_ = e.deleteParagraph(to_remove)
		to_remove = prev_p
	}

	// at least one paragraph keeps aliving so that current lastP is not changed when only one paragraph is deleted.
	// Example: ">" indicate current lastP,
	//	"line1"
	// >"line2"
	//
	// which is changed after DeleteLastParagraphs()
	//	"line1"
	// >""
	//
	if to_remove != 0 {
		e.setLasts()
		e.appendParagraph()
	}
}

// delete all contents in the Frame.
func (e *Editor) DeleteAll() {
	e.frame.mu.Lock()
	defer e.frame.mu.Unlock()
	maxp := int(e.frame.maxParagraphs)
	e.deleteFirstParagraphs(maxp)
}

// delete paragraph at frame's index removeP
// and return next index of removed paragraph.
// frame's firstP is valid after this.
func (e *Editor) deleteParagraph(removeP int32) int32 {
	f := e.frame

	// link prev and next of remove paragraph.
	removePP := &f.paragraphs[removeP]
	nextP := removePP.next // it will be returned.
	prevP := removePP.prev
	if prevP > 0 {
		f.paragraphs[prevP].next = nextP
	}
	if nextP > 0 {
		f.paragraphs[nextP].prev = prevP
	}

	// set frame's firstP index to avoid missing first Paragraph by remove it.
	if removeP == f.firstP {
		f.firstP = nextP
	}

	{ // remove the Paragraph, its lines and its boxes.
		var remove_len int32 = 0
		for l := removePP.firstL; l != 0; {
			ll := &f.lines[l]

			for b := ll.firstB; b != 0; {
				box := f.boxes[b]
				if ijbox, ok := box.(ijBox); ok {
					remove_len += ijbox.J() - ijbox.I()
				}
				nextB := box.nextIndex()
				f.freeBox(b)
				b = nextB
			}

			nextL := ll.next
			f.freeLine(l)
			l = nextL
		}
		f.freeParagraph(removeP)
		f.text_len -= remove_len
	}

	f.invalidateCaches()
	e.invalidateLasts()

	// allocate empty paragraph to remain at least one Paragraph, Line and Box.
	if nextP == 0 && prevP == 0 {
		newP, _ := e.newParagraph()
		newPP := &f.paragraphs[newP]

		newL, _ := f.newLine()
		newLL := &f.lines[newL]
		newPP.firstL = newL

		newB, _ := f.addBox(&textBox{})
		newLL.firstB = newB

		f.firstP = newP
		nextP = newP // it will be returned
	}

	// Compact c.f.text if it's large enough and the fraction of deleted text
	// is above some threshold. The actual threshold value (25%) is arbitrary.
	// A lower value means more frequent compactions, so less memory on average
	// but more CPU. A higher value means the opposite.
	if len(f.text) > 4096 && int32(len(f.text)/4) < f.deletedLen() {
		f.compactText()
	}
	return nextP
}

// layoutOneLine inserts a soft return in the Line l if its content measures longer than
// f.maxRuneWidth. This may spill content onto the next line, which will also be laid out,
// and so on recursively.
//
// NOTE: line count is changed in the paragraph, so needing to invalidate paragraph's cache.
func layoutOneLine(f *Frame, l int32) {
	for l != 0 {
		rwidth := 0
		nextL := int32(0) // remains if no break Line.
		ll := &f.lines[l]

		for b := ll.firstB; b != 0; {
			bb := f.boxes[b]
			rw := bb.RuneWidth(f)
			if rest := f.maxRuneWidth - rwidth; rest-rw < 0 {
				// break Line at box bb
				nextL = breakLine(f, l, b, rest)
				break
			}
			rwidth += rw
			b = bb.nextIndex()
		}

		l = nextL
	}
}

// breakLine breaks the Line l at box index b. The b index and rwidth must
// not be at the start or end of the Line. Content to the right of b in the
// Line l will be moved to the start of the next Line, with
// that next Line being created if it didn't already exist.
//
// It returns index of breaked next line.
func breakLine(f *Frame, l, b int32, rwidth int) int32 {
	bb := f.boxes[b]

	if tbox, ok := bb.(*textBox); !ok || rwidth == 0 {
		// Not splitable this Box or the space of placing this Box is nothing,
		// so break at a boundary between previous or next Box and this Box.
		prevBB := bb.Prev(f)
		if prevBB != nil {
			b = bb.prevIndex()
			bb = prevBB
		} else {
			// give up breaking this Box, so break at a boundary between this and next Box.
			//
			// bb is not changed.

			if nextBB := bb.Next(f); nextBB == nil {
				// give up breaking this Line, so not break and return next Line.
				// panic("TODO: First Box can not be breaked, how does treat it?")
				return f.lines[l].next
			}
		}
	} else {
		// Split this Box into two if possible, so that rwidth equals a Box's j end.
		if rwidth <= tbox.runewidth {
			pre, post := splitTextBox(f, tbox, rwidth)
			tbox.setText(*pre) // overwrite tbox by pre
			postB, _ := f.addBox(post)
			nextB := tbox.next
			if nextB != 0 {
				f.boxes[nextB].setPrevIndex(postB)
			}
			tbox.next = postB
			post.next = nextB
			post.prev = b
		}
	}

	// Assert that the break point isn't already at the end of the Line.
	if bb.nextIndex() == 0 {
		panic("Frame.Editor.breakLine: invalid state")
	}

	// Insert a line after this one, if one does'nt exist
	ll := &f.lines[l]
	if ll.next == 0 {
		newL, realloc := f.newLine()
		if realloc {
			ll = &f.lines[l]
		}
		f.lines[newL].prev = l
		ll.next = newL
	}

	// Move the remaining boxes to the next line.
	nextB, nextL := bb.nextIndex(), ll.next
	bb.setNextIndex(0)
	f.boxes[nextB].setPrevIndex(0)
	fb := f.lines[nextL].firstB
	f.lines[nextL].firstB = nextB

	// If the next Line already contained Boxes, append them to the end of the
	// nextB chain, and join the two newly linked Boxes if possible.
	if fb != 0 {
		lb := f.lines[nextL].lastBoxIndex(f)
		lbb := f.boxes[lb]
		fbb := f.boxes[fb]
		lbb.setNextIndex(fb)
		fbb.setPrevIndex(lb)
		f.joinBoxes(lb, fb, lbb, fbb)
	}

	return nextL
}

// given rune width rwidth, split *textBox into
// TextBox(rwidth), TextBox(maxrwidth-rwidth)
func splitTextBox(f *Frame, tb *textBox, rwidth int) (*textBox, *textBox) {
	if tb.i == tb.j {
		return &textBox{}, &textBox{}
	}

	total_rw := 0
	btext := f.text[tb.i:tb.j]
	pivot := tb.i

	for total_rw < rwidth && len(btext) != 0 {
		r, nbyte := utf8.DecodeRune(btext)
		if r == utf8.RuneError {
			panic("Frame: text contains invalid utf8 byte")
		}
		rw := width.RuneWidth(r)
		if rw+total_rw > rwidth {
			break
		}
		total_rw += rw
		pivot += int32(nbyte)
		btext = btext[nbyte:]
	}

	tb0, tb1 := &textBox{}, &textBox{}
	tb0.setText(*tb)
	tb1.setText(*tb)
	tb0.runewidth = total_rw
	tb1.runewidth = tb.runewidth - total_rw
	tb0.j, tb1.i = pivot, pivot
	return tb0, tb1
}
