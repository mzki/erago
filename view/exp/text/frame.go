// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Original source is golang.org/exp/shiny/text/text.go in 2016.06.
// Modification points of this file are:
//
// 1. Pragraph is treated as a line which ends with hard return.
//
// 2. Box is abstract data which contains text, image, ... etc.
//
// 3. Using rune width as a unit of layout, which measures singlebyte character is 1 and multibyte is 2.
//
// 4. Editing and viewing this frame is done by other struct data.
//

// package text provides contents holder cotaining plain-text, clickable-text-button, ... etc.
package text

import (
	"fmt"
	"sync"

	"golang.org/x/image/font"
)

// The size of frame buffer must not exceeds maxTextLen, in byte.
const maxTextLen = 0x7fffffff

// represents view contents, which contains text and image.
//
// referenced to golang.org/exp/shiny/text/text.go
type Frame struct {
	// These slices hold the Frame's Paragraphs, Lines and Boxes, indexed by
	// fields such as Paragraph.firstL and Box.next.
	//
	// Their contents are not necessarily in layout order. Each slice is
	// obviously backed by an array, but a Frame's list of children
	// (Paragraphs) forms a doubly-linked list, not an array list, so that
	// insertion has lower algorithmic complexity. Similarly for a Paragraph's
	// list of children (Lines) and a Line's list of children (Boxes).
	//
	// The 0'th index into each slice is a special case. Otherwise, each
	// element is either in use (forming a double linked list with its
	// siblings) or in a free list (forming a single linked list; the prev
	// field is -1).
	//
	// A zero firstFoo field means that the parent holds a single, implicit
	// (lazily allocated), empty-but-not-nil *Foo child. Every Frame contains
	// at least one Paragraph. Similarly, every Paragraph contains at least one
	// Line, and every Line contains at least one Box.
	//
	// A zero next or prev field means that there is no such sibling (for an
	// in-use Paragraph, Line or Box) or no such next free element (if in the
	// free list).

	paragraphs []Paragraph
	lines      []Line
	boxes      []Box

	// max range of paragraphs and its byte size.
	maxParagraphs, maxParagraphBytes int32

	// max width in runewidth.
	maxRuneWidth int

	// freeX is the index of the first X (Paragraph, Line or Box) in the
	// respective free list. Zero means that there is no such free element.
	freeP, freeL, freeB int32
	firstP              int32

	text []byte

	// len is the total length of the Frame's current textual content, in
	// bytes. It can be smaller then len(text), since that []byte can contain
	// 'holes' of deleted content.
	//
	// Like the paragraphs, lines and boxes slice-typed fields above, the text
	// []byte does not necessarily hold the textual content in layout order.
	// Instead, it holds the content in edit (insertion) order, with occasional
	// compactions. Again, the algorithmic complexity of insertions matters.
	text_len int32

	// TODO: it is held in View, remove it?
	face       font.Face
	faceHeight int32

	cachedHeightPlus1         int32
	cachedLineCountPlus1      int32
	cachedParagraphCountPlus1 int32

	// To prevent confliction of editing and viewing,
	// it is used by Editor or View only, not Frame.
	mu *sync.RWMutex

	// To notify frame's relayouting for editor.
	editor *Editor
}

// default size of max of rune width in Frame.
const DefaultMaxRuneWidth = 80

func NewFrame(opt *FrameOptions) *Frame {
	if opt == nil {
		opt = &defaultFrameOptions
	}
	if opt.MaxParagraphs < 0 {
		panic("paragraph count must be a positive number")
	}
	f := &Frame{
		maxParagraphs:     opt.MaxParagraphs,
		maxParagraphBytes: opt.MaxParagraphBytes,
		maxRuneWidth:      DefaultMaxRuneWidth,
		mu:                new(sync.RWMutex),
		// images: //TODO
	}
	f.initialize()
	return f
}

var defaultFrameOptions = FrameOptions{
	MaxParagraphs:     100,
	MaxParagraphBytes: 400,
}

type FrameOptions struct {
	// The max range of Frame's contents.
	MaxParagraphs, MaxParagraphBytes int32
}

// SetFace sets the font face for measuring text.
func (f *Frame) SetFace(face font.Face) {
	if !f.initialized() {
		f.initialize()
	}
	// TODO assert that given font must be monospace.

	f.face = face
	if face == nil {
		f.faceHeight = 0
	} else {
		// We round up the ascent and descent separately, instead of asking for
		// the metrics' height, since we quantize the baseline to the integer
		// pixel grid. For example, if ascent and descent were both 3.2 pixels,
		// then the naive height would be 6.4, which rounds up to 7, but we
		// should really provide 8 pixels (= ceil(3.2) + ceil(3.2)) between
		// each line to avoid overlap.
		//
		// TODO: is a font.Metrics.Height actually useful in practice??
		//
		// TODO: is it the font face's responsibility to track line spacing, as
		// in "double line spacing", or does that belong somewhere else, since
		// it doesn't affect the face's glyph masks?
		m := face.Metrics()
		f.faceHeight = int32(m.Ascent.Ceil() + m.Descent.Ceil())
	}
	if f.text_len != 0 {
		// TODO
		// f.relayout()
	}
}

// SetMaxRuneWidth sets the target maximum width of a Line of content, as a
// runewidth which is measured as: singlebyte: 1, multibyte:2. Contents will be broken
// so that a Line's width is less than or equal to this maximum width.
//
// If passed argument is less than or equal 1, it treats as default size 80.
func (f *Frame) SetMaxRuneWidth(rwidth int) {
	if rwidth <= 1 {
		rwidth = DefaultMaxRuneWidth
	}
	if !f.initialized() {
		f.initialize()
	}
	if f.maxRuneWidth == rwidth {
		return
	}
	f.maxRuneWidth = rwidth
	f.relayout()
}

func (f *Frame) relayout() {
	for p := f.firstP; p != 0; p = f.paragraphs[p].next {
		l := f.mergeIntoOneLine(p)
		layoutOneLine(f, l)
		f.paragraphs[p].invalidateCaches()
	}
	f.invalidateCaches()

	if e := f.editor; e != nil {
		e.invalidateLasts()
	}
}

// mergeIntoOneLine merges all of Lines in a Paragraph into a single Line, and
// compacts its empty and otherwise joinable Boxes. It returns the index of
// that Line.
func (f *Frame) mergeIntoOneLine(p int32) (l int32) {
	firstL := f.paragraphs[p].firstL
	ll := &f.lines[firstL]
	b0 := ll.firstB
	bb0 := f.boxes[b0]
	for {
		// Try to join sibling boxes and extract last box.
		if b1 := bb0.nextIndex(); b1 != 0 {
			bb1 := f.boxes[b1]
			if !f.joinBoxes(b0, b1, bb0, bb1) {
				b0, bb0 = b1, bb1
			}
			continue
		}

		if ll.next == 0 {
			f.paragraphs[p].invalidateCaches()
			f.lines[firstL].invalidateCaches()
			return firstL
		}

		// Unbreak the Line.
		nextLL := &f.lines[ll.next]
		b1 := nextLL.firstB
		bb1 := f.boxes[b1]
		bb0.setNextIndex(b1)
		bb1.setPrevIndex(b0)

		toFree := ll.next
		ll.next = nextLL.next
		// There's no need to fix up f.lines[ll.next].prev since it will just
		// be freed later in the loop.
		f.freeLine(toFree)
	}
}

// joinBoxes joins two adjacent Boxes if the Box.j field of the first one
// equals the Box.i field of the second, or at least one of them is empty. It
// returns whether they were joined. If they were joined, the second of the two
// Boxes is freed.
func (f *Frame) joinBoxes(b0, b1 int32, bb0, bb1 Box) bool {
	tb0, ok := bb0.(*textBox)
	if !ok {
		return false
	}
	tb1, ok := bb1.(*textBox)
	if !ok {
		return false
	}

	switch {
	case tb0.i == tb0.j:
		// The first Box is empty. Replace its i/j, color and so on with the second one's.
		tb0.setText(*tb1)
	case tb1.i == tb1.j:
		// The second box is empty. Drop it.
	case tb0.j == tb1.i && tb0.color == tb1.color:
		// The two non-empty Boxes are joinable.
		tb0.j = tb1.j
		tb0.runewidth += tb1.runewidth
	default:
		return false
	}
	tb0.next = tb1.next
	if tb0.next != 0 {
		f.boxes[tb0.next].setPrevIndex(b0)
	}
	f.freeBox(b1)
	return true
}

func (f *Frame) initialized() bool {
	return len(f.paragraphs) > 0
}

func (f *Frame) initialize() {
	// The first valid Paragraph, Line and Box all have index 1. The 0'th index
	// of each slice is a special case.
	f.paragraphs = make([]Paragraph, 2, 16)
	f.lines = make([]Line, 2, 16)
	f.boxes = make([]Box, 2, 16)

	f.text = make([]byte, 0, 1024)

	f.firstP = 1
	f.paragraphs[1].firstL = 1
	f.lines[1].firstB = 1

	f.boxes[0] = &textBox{} // not accesssed but set default to avoid nil.
	f.boxes[1] = &textBox{} // set default to avoid nil reference.
}

// newParagraph returns the index of an empty Paragraph, and whether or not the
// underlying memory has been re-allocated. Re-allocation means that any
// existing *Paragraph pointers become invalid.
func (f *Frame) newParagraph() (p int32, realloc bool) {
	if f.freeP != 0 {
		p := f.freeP
		pp := &f.paragraphs[p]
		f.freeP = pp.next
		*pp = Paragraph{}
		return p, false
	}
	realloc = len(f.paragraphs) == cap(f.paragraphs)
	f.paragraphs = append(f.paragraphs, Paragraph{})
	return int32(len(f.paragraphs) - 1), realloc
}

// newLine returns the index of an empty Line, and whether or not the
// underlying memory has been re-allocated. Re-allocation means that any
// existing *Line pointers become invalid.
func (f *Frame) newLine() (l int32, realloc bool) {
	if f.freeL != 0 {
		l := f.freeL
		ll := &f.lines[l]
		f.freeL = ll.next
		*ll = Line{}
		return l, false
	}
	realloc = len(f.lines) == cap(f.lines)
	f.lines = append(f.lines, Line{})
	return int32(len(f.lines) - 1), realloc
}

// newBox returns the index of an given box, and whether or not the underlying
// memory has been re-allocated. Re-allocation means that any existing *Box
// pointers become invalid.
func (f *Frame) addBox(new_box Box) (b int32, realloc bool) {
	if f.freeB != 0 {
		b := f.freeB
		f.freeB = f.boxes[b].nextIndex()
		f.boxes[b] = new_box
		return b, false
	}
	realloc = len(f.boxes) == cap(f.boxes)
	f.boxes = append(f.boxes, new_box)
	return int32(len(f.boxes) - 1), realloc
}

func (f *Frame) freeParagraph(p int32) {
	f.paragraphs[p] = Paragraph{next: f.freeP, prev: -1}
	f.freeP = p
	// TODO: run a compaction if the free-list is too large?
}

func (f *Frame) freeLine(l int32) {
	f.lines[l] = Line{next: f.freeL, prev: -1}
	f.freeL = l
	// TODO: run a compaction if the free-list is too large?
}

func (f *Frame) freeBox(b int32) {
	bb := f.boxes[b]
	bb.setNextIndex(f.freeB)
	bb.setPrevIndex(-1)
	f.freeB = b
	// TODO: run a compaction if the free-list is too large?
}

func (f *Frame) lastParagraphIndex() int32 {
	for p := f.firstP; ; {
		if next := f.paragraphs[p].next; next != 0 {
			p = next
			continue
		}
		return p
	}
}

// returns the first paragraph of this frame.
func (f *Frame) FirstParagraph() *Paragraph {
	if !f.initialized() {
		f.initialize()
	}
	return &f.paragraphs[f.firstP]
}

func (f *Frame) paragraph(p int32) *Paragraph {
	if p == 0 {
		return nil
	}
	return &f.paragraphs[p]
}

// return line indexed l. return nil if not found.
func (f *Frame) line(l int32) *Line {
	if l == 0 {
		return nil
	}
	return &f.lines[l]
}

func (f *Frame) box(b int32) Box {
	if b == 0 {
		return nil
	}
	return f.boxes[b]
}

func (f *Frame) invalidateCaches() {
	f.cachedHeightPlus1 = 0
	f.cachedLineCountPlus1 = 0
	f.cachedParagraphCountPlus1 = 0
}

// Height returns the height in pixels of this Frame.
func (f *Frame) Height() int {
	if !f.initialized() {
		f.initialize()
	}
	if f.cachedHeightPlus1 <= 0 {
		h := 1
		for p := f.firstP; p != 0; p = f.paragraphs[p].next {
			h += f.paragraphs[p].Height(f)
		}
		f.cachedHeightPlus1 = int32(h)
	}
	return int(f.cachedHeightPlus1 - 1)
}

// LineCount returns the number of Lines in this Frame.
//
// This count includes any soft returns inserted to wrap text to the maxWidth.
func (f *Frame) LineCount() int {
	if !f.initialized() {
		f.initialize()
	}
	if f.cachedLineCountPlus1 <= 0 {
		n := 1
		for p := f.firstP; p != 0; p = f.paragraphs[p].next {
			n += 1
		}
		f.cachedLineCountPlus1 = int32(n)
	}
	return int(f.cachedLineCountPlus1 - 1)
}

// ParagraphCount returns the number of Paragraphs in this Frame.
//
// This count excludes any soft returns inserted to wrap text to the maxWidth.
func (f *Frame) ParagraphCount() int {
	if !f.initialized() {
		f.initialize()
	}
	if f.cachedParagraphCountPlus1 <= 0 {
		n := 1
		for p := f.firstP; p != 0; p = f.paragraphs[p].next {
			n++
		}
		f.cachedParagraphCountPlus1 = int32(n)
	}
	return int(f.cachedParagraphCountPlus1 - 1)
}

// Len returns the number of bytes in the Frame's text.
func (f *Frame) Len() int {
	// We would normally check f.initialized() at the start of each exported
	// method of a Frame, but that is not necessary here. The Frame's text's
	// length does not depend on its Paragraphs, Lines and Boxes.
	return int(f.text_len)
}

// deletedLen returns the number of deleted bytes in the Frame's text.
func (f *Frame) deletedLen() int32 {
	return int32(len(f.text)) - f.text_len
}

func (f *Frame) compactText() {
	// f.text contains f.len live bytes and len(f.text) - f.len deleted bytes.
	// After the compaction, the new f.text slice's capacity should be at least
	// f.len, to hold all of the live bytes, but also be below len(f.text) to
	// allow total memory use to decrease. The actual value used (halfway
	// between them) is arbitrary. A lower value means less up-front memory
	// consumption but a lower threshold for re-allocating the f.text slice
	// upon further writes, such as a paste immediately after a cut. A higher
	// value means the opposite.

	newText := make([]byte, 0, f.text_len+f.deletedLen()/2)

	for p := f.firstP; p != 0; {
		pp := &f.paragraphs[p]
		for l := pp.firstL; l != 0; {
			ll := &f.lines[l]

			for b := ll.firstB; b != 0; {
				bb := f.boxes[b]
				if tbb, ok := bb.(TextBox); ok {
					i := int32(len(newText))
					newText = append(newText, tbb.Bytes(f)...)
					tbb.setI(i)
					tbb.setJ(int32(len(newText)))
				}
				b = bb.nextIndex()
			}

			l = ll.next
		}

		p = pp.next
	}
	if len(newText) != int(f.text_len) {
		panic(
			fmt.Sprintf("Frame.compactText: invalid state. textlen: %d, but newTextLen: %d", int(f.text_len), len(newText)),
		)
	}
	f.text = newText
}

// Paragraph holds Lines of text.
//
// Here, Paragraph is treated as actual Line.
// Its internal Lines are treated as layouted lines,
// which are adjust to Frame's max width.
type Paragraph struct {
	firstL, next, prev   int32
	cachedHeightPlus1    int32
	cachedLineCountPlus1 int32

	align Alignment
}

func (p *Paragraph) lastLineIndex(f *Frame) int32 {
	for l := p.firstL; ; {
		if next := f.lines[l].next; next != 0 {
			l = next
			continue
		}
		return l
	}
}

// FirstLine returns the first Line of this Paragraph.
//
// f is the Frame that contains the Paragraph.
func (p *Paragraph) FirstLine(f *Frame) *Line {
	return &f.lines[p.firstL]
}

// Next returns the next Paragraph after this one in the Frame.
//
// f is the Frame that contains the Paragraph.
func (p *Paragraph) Next(f *Frame) *Paragraph {
	return f.paragraph(p.next)
}

func (p *Paragraph) Prev(f *Frame) *Paragraph {
	return f.paragraph(p.prev)
}

func (p *Paragraph) invalidateCaches() {
	p.cachedHeightPlus1 = 0
	p.cachedLineCountPlus1 = 0
}

// Height returns the height in pixels of this Paragraph.
func (p *Paragraph) Height(f *Frame) int {
	if p.cachedHeightPlus1 <= 0 {
		h := 1
		for l := p.firstL; l != 0; l = f.lines[l].next {
			h += f.lines[l].Height(f)
		}
		p.cachedHeightPlus1 = int32(h)
	}
	return int(p.cachedHeightPlus1 - 1)
}

// LineCount returns the number of Lines in this Paragraph.
//
// This count includes any soft returns inserted to wrap text to the maxWidth.
func (p *Paragraph) LineCount(f *Frame) int {
	if p.cachedLineCountPlus1 <= 0 {
		n := 1
		for l := p.firstL; l != 0; l = f.lines[l].next {
			n++
		}
		p.cachedLineCountPlus1 = int32(n)
	}
	return int(p.cachedLineCountPlus1 - 1)
}

// Line holds Boxes of contents.
type Line struct {
	firstB, next, prev int32
	cachedHeightPlus1  int32
}

// return next Line. if not found return nil.
func (l *Line) Next(f *Frame) *Line {
	return f.line(l.next)
}

// return prev Line. if not found return nil.
func (l *Line) Prev(f *Frame) *Line {
	return f.line(l.prev)
}

// return first Box. if not found return nil,
// indicating the line has no contents.
func (l *Line) FirstBox(f *Frame) Box {
	if l.firstB == 0 {
		return nil
	}
	return f.boxes[l.firstB]
}

func (l *Line) lastBoxIndex(f *Frame) int32 {
	for b := l.firstB; ; {
		if next := f.boxes[b].nextIndex(); next != 0 {
			b = next
			continue
		}
		return b
	}
}

func (l *Line) invalidateCaches() {
	l.cachedHeightPlus1 = 0
}

// Height returns the height in pixels of this Line.
func (l *Line) Height(f *Frame) int {
	// TODO: measure the height of each box, if we allow rich text (i.e. more
	// than one Frame-wide font face).
	if f.face == nil {
		return 0
	}
	if l.cachedHeightPlus1 <= 0 {
		l.cachedHeightPlus1 = f.faceHeight + 1
	}
	return int(l.cachedHeightPlus1 - 1)
}

// it returns sum of runewidth of all Boxes in the Line.
func (l *Line) RuneWidth(f *Frame) int {
	rwidth := 0
	for b := l.firstB; b != 0; {
		bb := f.boxes[b]
		rwidth += bb.RuneWidth(f)
		b = bb.nextIndex()
	}
	// TODO: caching total rwidth?
	return rwidth
}

// return string entire content in the given Frame.
func String(f *Frame) string {
	return StringFrom(f, f.firstP)
}

// return string cotent in the given Frame
// in range from startP to end.
func StringFrom(f *Frame, startP int32) string {
	buf := make([]byte, 0, f.text_len)
	for p := f.paragraph(startP); p != nil; p = p.Next(f) {
		for l := p.FirstLine(f); l != nil; l = l.Next(f) {
			for b := l.FirstBox(f); b != nil; b = b.Next(f) {
				if tbox, ok := b.(TextBox); ok {
					buf = append(buf, tbox.Bytes(f)...)
				}
			}
		}
	}
	return string(buf)
}
