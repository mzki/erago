package model

import (
	"context"
	"encoding/json"
	"time"

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
	ImageFetchNone       int = int(publisher.ImageFetchNone)
	ImageFetchRawRGBA    int = int(publisher.ImageFetchRawRGBA)
	ImageFetchEncodedPNG int = int(publisher.ImageFetchEncodedPNG)
)

var imageFetchTypeIntToEnum = map[int]publisher.ImageFetchType{
	ImageFetchNone:       publisher.ImageFetchNone,
	ImageFetchRawRGBA:    publisher.ImageFetchRawRGBA,
	ImageFetchEncodedPNG: publisher.ImageFetchEncodedPNG,
}

func pbImageFetchType(v int) publisher.ImageFetchType {
	if ret, ok := imageFetchTypeIntToEnum[v]; ok {
		return ret
	} else {
		return publisher.ImageFetchNone
	}
}

const (
	MessageByteEncodingJson int = iota
	MessageByteEncodingProtobuf
)

type uiAdapterOptions struct {
	ImageFetchType      publisher.ImageFetchType
	ImageCacheSize      int
	MessageByteEncoding int

	EnableDebugTimestamp bool
}

func newUIAdapter(ctx context.Context, ui UI, opt uiAdapterOptions) (*uiAdapter, error) {
	editor := publisher.NewEditor(ctx, publisher.EditorOptions{
		ImageFetchType: opt.ImageFetchType,
		ImageCacheSize: opt.ImageCacheSize,
	})
	// set default
	if err := editor.SetViewSize(20, 40); err != nil {
		return nil, err
	}
	binEncode := newParagraphBinaryEncodeFunc(opt.MessageByteEncoding)
	if err := editor.SetCallback(&publisher.CallbackDefault{
		OnPublishFunc: func(p *pubdata.Paragraph) error {
			bs, err := binEncode(p)
			if err != nil {
				return err
			}
			// debug timestamp should be first than Publish so that listener can know
			// the timestamp to link to the next Publish.
			if opt.EnableDebugTimestamp {
				now := time.Now()
				err = ui.OnDebugTimestamp(p.Id, now.Format(time.RFC3339Nano), now.UnixNano())
				if err != nil {
					return err
				}
			}
			return ui.OnPublishBytes(bs)
		},
		OnPublishTemporaryFunc: func(p *pubdata.Paragraph) error {
			bs, err := binEncode(p)
			if err != nil {
				return err
			}
			return ui.OnPublishBytesTemporary(bs)
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

type paragraphBinaryEncoderFunc func(*pubdata.Paragraph) ([]byte, error)

func newParagraphBinaryEncodeFunc(encoding int) paragraphBinaryEncoderFunc {
	var encoder paragraphBinaryEncoderFunc
	switch encoding {
	case MessageByteEncodingProtobuf:
		encoder = func(p *pubdata.Paragraph) ([]byte, error) {
			return p.MarshalVT()
		}
	case MessageByteEncodingJson:
		fallthrough
	default:
		encoder = func(p *pubdata.Paragraph) ([]byte, error) {
			return json.Marshal(p)
		}
	}
	return encoder
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
