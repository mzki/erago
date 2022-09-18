package model

import (
	"context"
	"encoding/json"

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

var (
	// Export ImageFetchType as int for mobile client.
	ImageFetchNone       int = pubdata.ImageFetchNone
	ImageFetchRawRGBA    int = pubdata.ImageFetchRawRGBA
	ImageFetchEncodedPNG int = pubdata.ImageFetchEncodedPNG
)

func newUIAdapter(ctx context.Context, ui UI, imageFetchType pubdata.ImageFetchType) (*uiAdapter, error) {
	editor := publisher.NewEditor(ctx, publisher.EditorOptions{
		ImageFetchType: imageFetchType,
	})
	// set default
	if err := editor.SetViewSize(20, 40); err != nil {
		return nil, err
	}
	if err := editor.SetCallback(&publisher.CallbackDefault{
		OnPublishFunc: func(p *pubdata.Paragraph) error {
			bs, err := json.Marshal(p)
			if err != nil {
				return err
			}
			return ui.OnPublishJson(string(bs))
		},
		OnPublishTemporaryFunc: func(p *pubdata.Paragraph) error {
			bs, err := json.Marshal(p)
			if err != nil {
				return err
			}
			return ui.OnPublishJsonTemporary(string(bs))
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
