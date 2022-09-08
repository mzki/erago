package ui

import (
	"image"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/mouse"

	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text"
)

// View is interface of text.Frame and its Printer.
// Any text.Frame is treated through View.
// View is identified by uniq name.
type TextView struct {
	node.LeafEmbed
	// implements erago/uiadapter.Printer interface.
	*Printer
	closed bool // closed Printer?

	frame *text.Frame

	paddingCache int

	// identifer for this.
	name string

	sender *EragoPresenter

	focused bool
}

//
func NewTextView(name string, sender *EragoPresenter) *TextView {
	if sender == nil {
		panic("nil sender is not allowed")
	}
	f := text.NewFrame(nil)
	view := &TextView{
		name:    name,
		frame:   f,
		Printer: NewPrinter(f),
		sender:  sender,
	}
	view.Wrapper = view
	view.Focus()
	return view
}

// close to explicitly finalize this.
func (v *TextView) Close() {
	if v.closed {
		return
	}
	v.closed = true
	v.Printer.e.Close()
}

func (v *TextView) padding(t *theme.Theme) int {
	return t.Pixels(unit.Chs(1)).Round()
}

// implements node.Node interface
func (v *TextView) Measure(t *theme.Theme, widthHint, heightHint int) {
	if widthHint < 0 {
		widthHint = 0
	}
	if heightHint < 0 {
		heightHint = 0
	}
	// its size depends on parent's measuring.
	v.MeasuredSize = image.Point{widthHint, heightHint}
}

// implements node.Node interface
func (v *TextView) Layout(t *theme.Theme) {
	vSize := v.Rect.Size()
	padding := v.padding(t)
	v.paddingCache = padding
	vSize.X -= 2 * padding
	if vSize.X < 0 {
		vSize.X = 0
	}

	face := t.AcquireFontFace(theme.FontFaceOptions{})
	defer t.ReleaseFontFace(theme.FontFaceOptions{}, face)
	v.v.SetFace(face)
	v.v.SetSize(vSize)
}

// implements node.Node interface
// if it has no MarkPaintBase then do nothing.
func (v *TextView) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	// TODO: ditinguish paint and paintBase then this code is OK.
	// if !v.Marks.NeedsPaintBase() {
	// 	return nil
	// }
	v.Marks.UnmarkNeedsPaintBase()

	dst := ctx.Dst
	vRect := v.Rect.Add(origin)

	t := ctx.Theme
	draw.Draw(dst, vRect, theme.Background.Uniform(t), image.Point{}, draw.Src)

	dstOrigin := vRect.Min
	dstOrigin.X += v.paddingCache
	v.v.Draw(dst, dstOrigin)
	return nil
}

// Focused means this view is current view on viewManager, not lifecycle.Focused.
//
func (v *TextView) Focus() {
	v.focused = true
}
func (v *TextView) Unfocus() {
	v.focused = false
}

// implements node.Node interface
// if it has no focus then do nothing and return NotHandled.
func (v *TextView) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	if !v.focused {
		return node.NotHandled
	}

	switch ev := ev.(type) {
	case mouse.Event:
		// check scroll
		var inputWaiting = v.sender.InputWaiting()
		var scrollLine = 0
		switch ev.Button {
		case mouse.ButtonWheelUp:
			scrollLine = 1
		case mouse.ButtonWheelDown:
			scrollLine = -1
		}
		if scrollLine != 0 && inputWaiting {
			v.v.ScrollLine(scrollLine)
			v.v.UnhighlightCommand()
			v.Mark(node.MarkNeedsPaintBase)
			return node.Handled
		}

		// check highlight on pointed button
		var p = image.Point{round(ev.X), round(ev.Y)}
		var changed = false
		if v.sender.CommandWaiting() && v.HighlightCommand(p, origin) {
			changed = true
		} else {
			changed = changed || v.v.UnhighlightCommand()
		}
		if changed {
			v.Mark(node.MarkNeedsPaintBase)
		}

	case gesture.Event:
		switch ev.Type {
		case gesture.TypeTap:
			if ev.DoublePress { // exactly tap once.
				return node.NotHandled
			}
			if !v.sender.InputWaiting() {
				return node.NotHandled
			}
			var p = image.Point{round(ev.CurrentPos.X), round(ev.CurrentPos.Y)}
			if cmd, found := v.FindCommand(p, origin); found {
				v.sender.SendCommand(cmd)
			} else {
				v.sender.SendCommand("")
			}
			return node.Handled

		case gesture.TypeIsDoublePress:
			v.sender.SendControlSkippingWait(true)
			return node.Handled
		}
	}
	return node.NotHandled
}

// find clicakble Command at the postion.
// Return command and command found.
func (v *TextView) FindCommand(at, origin image.Point) (string, bool) {
	// text.View requires a point in phisical screen coordinate space
	// to find underlying command.
	return v.v.FindCommand(at)
}

// highlight Command at the postion.
// Return highlighted command is found.
func (v *TextView) HighlightCommand(at, origin image.Point) bool {
	// text.View requires a point in phisical screen coordinate space
	// to find underlying command.
	return v.v.HighlightCommand(at)
}

// implements fmt.Stringer
func (v *TextView) String() string {
	return text.String(v.frame)
}

//
// View.Printer implements uiadapter.Printer interface, but
// View can not Marks asynchronously.
// So send Mark event to eventQ.
//

func (v *TextView) Print(s string) error {
	v.Printer.Print(s)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) PrintLabel(s string) error {
	v.Printer.PrintLabel(s)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) PrintButton(caption string, command string) error {
	v.Printer.PrintButton(caption, command)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) PrintLine(sym string) error {
	v.Printer.PrintLine(sym)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) PrintImage(file string, widthInRW, heightInLC int) error {
	v.Printer.PrintImage(file, widthInRW, heightInLC)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) NewPage() error {
	v.Printer.NewPage()
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) ClearLine(nline int) error {
	v.Printer.ClearLine(nline)
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) ClearLineAll() error {
	v.Printer.ClearLineAll()
	v.sender.Mark(v, node.MarkNeedsPaintBase)
	return nil
}

func (v *TextView) GetColor() (uint32, error) {
	return v.Printer.GetColor(), nil
}

func (v *TextView) SetColor(c uint32) error {
	v.Printer.SetColor(c)
	return nil
}

func (v *TextView) ResetColor() error {
	v.Printer.ResetColor()
	return nil
}

func (v *TextView) GetAlignment() (attr.Alignment, error) {
	return v.Printer.GetAlignment(), nil
}

func (v *TextView) SetAlignment(a attr.Alignment) error {
	v.Printer.SetAlignment(a)
	return nil
}

func (v *TextView) CurrentRuneWidth() (int, error) {
	return v.Printer.CurrentRuneWidth(), nil
}

func (v *TextView) WindowRuneWidth() (int, error) {
	return v.Printer.WindowRuneWidth(), nil
}

func (v *TextView) LineCount() (int, error) {
	return v.Printer.LineCount(), nil
}

func (v *TextView) WindowLineCount() (int, error) {
	return v.Printer.WindowLineCount(), nil
}

func (v *TextView) Sync() error {
	v.sender.sync(v)
	return nil
}
