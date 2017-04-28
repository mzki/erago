package ui

import (
	"strings"

	attr "local/erago/attribute"
	"local/erago/view/exp/text"
)

// Printer is interface for text.Frame.
// It implements erago.uiadapter.UI.Printer
type Printer struct {
	e *text.Editor

	// NOTE: View is not able to be used concurrently.
	// here, using this is limited to only read its property.
	v *text.View
}

func NewPrinter(f *text.Frame) *Printer {
	return &Printer{
		e: f.Editor(),
		v: f.View(),
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) Print(s string) {
	_, err := p.e.WriteText(s)
	if err != nil {
		panic(err) // TODO return error?
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) PrintLabel(s string) {
	_, err := p.e.WriteLabel(s)
	if err != nil {
		panic(err) // TODO return error?
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) PrintButton(caption, cmd string) {
	_, err := p.e.WriteButton(caption, cmd)
	if err != nil {
		panic(err) // TODO return error?
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) PrintLine(sym string) {
	_, err := p.e.WriteLine(sym)
	if err != nil {
		panic(err)
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) SetColor(c uint32) {
	Color := &p.e.Color
	Color.R = uint8((c & 0x00ff0000) >> 16)
	Color.B = uint8((c & 0x0000ff00) >> 8)
	Color.G = uint8((c & 0x000000ff) >> 0)
	Color.A = 0xff
}

// implemnts erago/uiadapter/UI.
func (p Printer) GetColor() uint32 {
	var c uint32 = 0
	Color := p.e.Color
	c |= (uint32(Color.R) << 16)
	c |= (uint32(Color.B) << 8)
	c |= (uint32(Color.G) << 0)
	return c
}

// implemnts erago/uiadapter/UI.
func (p Printer) ResetColor() {
	p.e.Color = text.ResetColor
}

// implemnts erago/uiadapter/UI.
func (p Printer) SetAlignment(a attr.Alignment) {
	p.e.SetAlignment(text.Alignment(a))
}

// implemnts erago/uiadapter/UI.
func (p Printer) GetAlignment() attr.Alignment {
	return attr.Alignment(p.e.GetAlignment())
}

// implemnts erago/uiadapter/UI.
func (p Printer) NewPage() {
	s := strings.Repeat("\n", p.v.LineCount())
	_, err := p.e.WriteText(s)
	if err != nil {
		panic(err)
	}
}

// implemnts erago/uiadapter/UI.
func (p Printer) ClearLine(nline int) {
	p.e.DeleteLastParagraphs(nline)
}

// implemnts erago/uiadapter/UI.
func (p Printer) ClearLineAll() {
	p.e.DeleteAll()
}

// implemnts erago/uiadapter/UI.
func (p Printer) CurrentRuneWidth() int {
	return p.e.CurrentRuneWidth()
}

// implemnts erago/uiadapter/UI.
func (p Printer) MaxRuneWidth() int {
	return p.v.RuneWidth()
}

// implemnts erago/uiadapter/UI.
func (p Printer) LineCount() int {
	return p.v.LineCount()
}
