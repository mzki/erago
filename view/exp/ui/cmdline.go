// command line interface.

package ui

import (
	"image"
	"io/ioutil"

	"golang.org/x/exp/shiny/text"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
)

// CommandLine is a leaf widget which can edit its text content and submit it.
type CommandLine struct {
	node.LeafEmbed

	sender CmdSender

	f *text.Frame
	c *text.Caret
}

// construct new CommandLine widget with custom theme.
func NewCommandLine(sender CmdSender) *CommandLine {
	f := &text.Frame{}
	c := f.NewCaret()

	cl := &CommandLine{f: f, c: c, sender: sender}
	cl.Wrapper = cl
	return cl
}

// finalization. after this editing text does not work.
// It is exported for unexpected panic.
// It is called automatically onLifeCycelEvent.
func (cl CommandLine) Close() {
	cl.c.Close()
}

func (cl *CommandLine) OnLifecycleEvent(e lifecycle.Event) {
	cl.LeafEmbed.OnLifecycleEvent(e)
	if e.To == lifecycle.StageDead {
		cl.Close()
	}
}

// implements node.Node interface.
func (cl *CommandLine) Measure(t *theme.Theme, widthHint, heightHint int) {
	face := t.AcquireFontFace(theme.FontFaceOptions{})
	defer t.ReleaseFontFace(theme.FontFaceOptions{}, face)
	metrics := face.Metrics()

	cl.MeasuredSize.X = widthHint
	cl.MeasuredSize.Y = metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	cl.f.SetFace(face)
	cl.f.SetMaxWidth(fixed.I(cl.MeasuredSize.X))
}

// implements node.Node interface.
func (cl *CommandLine) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	cl.LeafEmbed.PaintBase(ctx, origin)

	drawPoint := cl.Rect.Min.Add(origin)
	face := ctx.Theme.AcquireFontFace(theme.FontFaceOptions{})
	defer ctx.Theme.ReleaseFontFace(theme.FontFaceOptions{}, face)
	drawPoint.Y += face.Metrics().Ascent.Ceil()

	// TODO draw only last line of text.
	d := font.Drawer{
		Dot:  fixed.P(drawPoint.X, drawPoint.Y),
		Dst:  ctx.Dst,
		Src:  ctx.Theme.Palette.Foreground(),
		Face: face,
	}

	// extract last line
	f := cl.f
	var lastP *text.Paragraph
	for lastP = f.FirstParagraph(); ; {
		next_p := lastP.Next(f)
		if next_p == nil {
			break
		}
		lastP = next_p
	}
	var lastL *text.Line
	for lastL = lastP.FirstLine(f); ; {
		next_l := lastL.Next(f)
		if next_l == nil {
			break
		}
		lastL = next_l
	}

	for b := lastL.FirstBox(f); b != nil; b = b.Next(f) {
		d.DrawBytes(b.TrimmedText(f))
	}
	return nil
}

// implements node.Node interface.
func (cl *CommandLine) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	switch ev := ev.(type) {
	case key.Event:
		if ev.Direction == key.DirPress {
			switch ev.Code {
			case key.CodeDeleteBackspace:
				cl.Delete(1)
				cl.Mark(node.MarkNeedsPaintBase)
				return node.Handled
			case key.CodeReturnEnter:
				cmd := cl.Confirm()
				cl.Mark(node.MarkNeedsPaintBase)
				if s := cl.sender; s != nil {
					s.SendCommand(cmd)
				}
				return node.Handled
			}
			if r := ev.Rune; r > 0 {
				cl.Append(string(r))
				cl.Mark(node.MarkNeedsPaintBase)
				if s := cl.sender; s != nil {
					s.SendRawCommand(r)
				}
				return node.Handled
			}
		}
	}
	return node.NotHandled
}

// append text at end of command line. and position is moved to end.
func (cl *CommandLine) Append(s string) {
	cl.Mark(node.MarkNeedsPaintBase)
	_, err := cl.c.Seek(0, text.SeekEnd)
	if err != nil {
		panic("cmdline: can not seek to end of text. : " + err.Error())
	}
	cl.c.WriteString(s)
}

// insert text at current positon
func (cl *CommandLine) InsertString(s string) {
	cl.Mark(node.MarkNeedsPaintBase)
	cl.c.WriteString(s)
}

// confirm current text as command and return it.
// all of the command line's text is cleared after confirming.
func (cl *CommandLine) Confirm() string {
	if _, err := cl.c.Seek(0, text.SeekSet); err != nil {
		panic("cmdline: can not seek to begin: " + err.Error())
	}

	read, err := ioutil.ReadAll(cl.c)
	if err != nil {
		panic("cmdline: read all byte from text.Frame: " + err.Error())
	}

	cl.c.Delete(text.Backwards, len(read))
	return string(read)
}

// delete n runes backwards. It returns deleted number of runes.
func (cl *CommandLine) Delete(n_rune int) int {
	cl.Mark(node.MarkNeedsPaintBase)
	_, err := cl.c.Seek(0, text.SeekEnd)
	if err != nil {
		panic("cmdline: can not seek to end of text. : " + err.Error())
	}
	d_rune, _ := cl.c.DeleteRunes(text.Backwards, n_rune)
	return d_rune
}

func (cl *CommandLine) MoveCurosr(n_rune int) {
	// TODO: move cursor by seeking caret.
	cl.Mark(node.MarkNeedsPaintBase)
}
