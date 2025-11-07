package app

import (
	"image"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"

	"github.com/mzki/erago/app/config"
	"github.com/mzki/erago/uiadapter"
	customT "github.com/mzki/erago/view/exp/theme"
	"github.com/mzki/erago/view/exp/ui"
)

var DefaultAppTextViewOptions = ui.DefaultTextViewOptions

// UI is mixture of widgets in ui package.
type UI struct {
	node.ShellEmbed

	mv *ui.MultipleView

	// These Backgrounds fill each widgets by background color.
	// To change color you can modify XXXBackground.ThemeColor.
	CmdLineBackground *widget.Uniform

	// for workaround
	lastMouseEvent mouse.Event
}

// construct standard UI node tree for the era application.
func NewUI(presenter *ui.EragoPresenter, appConf *config.Config) *UI {
	ui.DefaultTextViewOptions = DefaultAppTextViewOptions
	if appConf.HistoryLineCount > 0 {
		ui.DefaultTextViewOptions.MaxParagraphs = int32(appConf.HistoryLineCount)
	}
	// if appConf.HistoryBytesPerLine > 0 {
	// 	ui.DefaultTextViewOptions.MaxParagraphBytes = int32(appConf.HistoryBytesPerLine)
	// }
	mv := ui.NewMultipleView(presenter, ui.TextViewOptions{
		TextFrameOptions: ui.TextFrameOptions{
			MaxParagraphs:     int32(appConf.HistoryLineCount),
			MaxParagraphBytes: ui.DefaultTextViewOptions.MaxParagraphs,
		},
		ImageCacheSize: appConf.ImageCacheSize,
	})
	bg_cmd := widget.NewUniform(theme.Background,
		widget.NewPadder(widget.AxisHorizontal, unit.Ems(0.5),
			ui.NewCommandLine(presenter),
		),
	)
	fixed := ui.NewFixedSplit(ui.EdgeBottom, customT.Lhs(1),
		ui.NewSheet(bg_cmd),
		ui.NewSheet(mv),
	)

	ui := &UI{
		mv:                mv,
		CmdLineBackground: bg_cmd,
	}
	ui.Wrapper = ui
	ui.Insert(fixed, nil)
	return ui
}

// implement node.Node interface.
func (ui *UI) OnLifecycleEvent(e lifecycle.Event) {
	ui.ShellEmbed.OnLifecycleEvent(e)
	if e.To == lifecycle.StageDead {
		// TODO presenter quitting?.
	}
}

// return erago/uiadapter.UI interface.
func (ui UI) Editor() uiadapter.UI {
	return ui.mv
}

func (ui *UI) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	// Workaround: In case of windows, multidisplay and mouse cursur on sub display which is left or above of primary display,
	// point X and Y in mouse event with DirStep indicates 65522 or something.
	// It seems uint16 underflow is occured (-1) --> (65535).
	// But the event with other direction, X and Y are normal e.g. 0 <= x <= 1920 in 1080p screen.
	// Here performs remembering last normal mouse event and replace abnormal X and Y by last normal one.
	switch testEvent := e.(type) {
	case mouse.Event:
		if testEvent.Direction == mouse.DirNone {
			ui.lastMouseEvent = testEvent
		} else if testEvent.Direction == mouse.DirStep {
			newMouseEvent := testEvent
			newMouseEvent.X = ui.lastMouseEvent.X
			newMouseEvent.Y = ui.lastMouseEvent.Y
			e = newMouseEvent
		}
	}
	return ui.ShellEmbed.OnInputEvent(e, origin)
}
