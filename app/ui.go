package app

import (
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/lifecycle"

	"local/erago/uiadapter"
	customT "local/erago/view/exp/theme"
	"local/erago/view/exp/ui"
)

// UI is mixture of widgets in ui package.
type UI struct {
	node.ShellEmbed

	mv *ui.MultipleView

	// These Backgrounds fill each widgets by background color.
	// To change color you can modify XXXBackground.ThemeColor.
	CmdLineBackground *widget.Uniform
}

// construct standard UI node tree for the era application.
func NewUI(presenter *ui.EragoPresenter) *UI {
	mv := ui.NewMultipleView(presenter)
	bg_cmd := widget.NewUniform(theme.Background,
		widget.NewPadder(widget.AxisHorizontal, unit.Ems(0.5),
			ui.NewCommandLine(presenter),
		),
	)
	fixed := ui.NewFixedSplit(ui.EdgeBottom, customT.Lhs(1),
		widget.NewSheet(bg_cmd),
		widget.NewSheet(mv),
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
