package model

import (
	"context"

	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
)

// adapter for erago.UI interface.
// it converts model.UI interface to uiadapter.Printer one.
type uiAdapter struct {
	uiadapter.Printer

	editor   *publisher.Editor
	mobileUI UI
}

func newUIAdapter(ctx context.Context, ui UI) (*uiAdapter, error) {
	editor := publisher.NewEditor(ctx)
	// set default
	if err := editor.SetViewSize(20, 40); err != nil {
		return nil, err
	}
	if err := editor.SetCallback(&publisher.CallbackDefault{
		OnPublishFunc: func(p *pubdata.Paragraph) error {
			return ui.OnPublish((*Paragraph)(p))
		},
		OnPublishTemporaryFunc: func(p *pubdata.Paragraph) error {
			return ui.OnPublishTemporary((*Paragraph)(p))
		},
		OnRemoveFunc:    ui.OnRemove,
		OnRemoveAllFunc: ui.OnRemoveAll,
	}); err != nil {
		return nil, err
	}
	return &uiAdapter{
		Printer:  editor,
		editor:   editor,
		mobileUI: ui,
	}, nil
}

// implement uiadapter.RequestChangedListener
func (a *uiAdapter) OnRequestChanged(req uiadapter.InputRequestType) {
	switch req {
	case uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
		a.mobileUI.OnCommandRequested()
	case uiadapter.InputRequestInput:
		a.mobileUI.OnInputRequested()
	case uiadapter.InputRequestNone:
		a.mobileUI.OnInputRequestClosed()
	}
}
