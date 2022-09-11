package model

import (
	"github.com/mzki/erago/attribute"
	"github.com/mzki/erago/uiadapter"
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

func (a *uiAdapter) PrintImage(string, int, int) error {
	panic("Not implemented")
}

func (a *uiAdapter) MeasureImageSize(string, int, int) (int, int, error) {
	panic("Not implemented")
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

func (a *uiAdapter) SetAlignment(align attribute.Alignment) error {
	// store value to view context
	a.alignment = align
	switch align {
	case attribute.AlignmentLeft:
		return a.UI.SetAlignmentLeft()
	case attribute.AlignmentCenter:
		return a.UI.SetAlignmentCenter()
	case attribute.AlignmentRight:
		return a.UI.SetAlignmentRight()
	}
	return nil
}

func (a *uiAdapter) GetAlignment() (attribute.Alignment, error) {
	// return uiAdapter context value which may not be synchronized
	// to model.UI context.
	return a.alignment, nil
}

func (a *uiAdapter) SetColor(color uint32) error {
	return a.UI.SetColor(int32(color))
}

func (a *uiAdapter) GetColor() (uint32, error) {
	c, err := a.UI.GetColor()
	return uint32(c), err
}

func (a *uiAdapter) ClearLine(n int) error {
	return a.UI.ClearLine(int32(n))
}

func (a *uiAdapter) WindowRuneWidth() (int, error) {
	w, err := a.UI.WindowLineWidth()
	return int(w), err
}

func (a *uiAdapter) WindowLineCount() (int, error) {
	c, err := a.UI.WindowLineCount()
	return int(c), err
}

func (a *uiAdapter) CurrentRuneWidth() (int, error) {
	w, err := a.UI.CurrentLineWidth()
	return int(w), err
}

func (a *uiAdapter) LineCount() (int, error) {
	c, err := a.UI.LineCount()
	return int(c), err
}
