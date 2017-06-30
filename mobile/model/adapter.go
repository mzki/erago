package model

import (
	"local/erago/attribute"
	"local/erago/uiadapter"
)

// adapter for erago.UI interface.
// it converts model.UI interface to erago.UI one.
type uiAdapter struct {
	UI

	// uiAdapter context synchronized to model.UI.
	// these fields are managed by this scope
	// because these values are not exported to mobile.
	alignment attribute.Alignment
}

func newUIAdapter(ui UI) *uiAdapter {
	return &uiAdapter{
		UI:        ui,
		alignment: attribute.AlignmentLeft,
	}
}

// implement uiadapter.RequestChangedListener
func (a *uiAdapter) OnRequestChanged(req uiadapter.InputRequestType) {
	switch req {
	case uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
		a.UI.OnCommandRequested()
	case uiadapter.InputRequestInput:
		a.UI.OnInputRequested()
	case uiadapter.InputRequestNone:
		a.UI.OnInputRequestClosed()
	}
}

func (a *uiAdapter) SetAlignment(align attribute.Alignment) {
	// store value to view context
	a.alignment = align
	switch align {
	case attribute.AlignmentLeft:
		a.UI.SetAlignmentLeft()
	case attribute.AlignmentCenter:
		a.UI.SetAlignmentCenter()
	case attribute.AlignmentRight:
		a.UI.SetAlignmentRight()
	}
}

func (a *uiAdapter) GetAlignment() attribute.Alignment {
	// return uiAdapter context value which may not be synchronized
	// to model.UI context.
	return a.alignment
}

func (a *uiAdapter) SetColor(color uint32) {
	a.UI.SetColor(int32(color))
}

func (a *uiAdapter) GetColor() uint32 {
	return uint32(a.UI.GetColor())
}

func (a *uiAdapter) ClearLine(n int) {
	a.UI.ClearLine(int32(n))
}

func (a *uiAdapter) CurrentRuneWidth() int {
	return int(a.UI.CurrentLineWidth())
}

func (a *uiAdapter) MaxRuneWidth() int {
	return int(a.UI.MaxLineWidth())
}

func (a *uiAdapter) LineCount() int {
	return int(a.UI.LineCount())
}
