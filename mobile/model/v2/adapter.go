package model

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
)

// adapter for erago.UI interface.
// it converts model.UI interface to uiadapter.Printer one.
type uiAdapter struct {
	uiadapter.Printer

	editor    *publisher.Editor
	throttler *CallbackThrottler
	mobileUI  UI
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
	ThrottleDuration    time.Duration

	EnableDebugTimestamp bool
}

// Normalize normalizes optional parameters into valid range.
func (opt *uiAdapterOptions) Normalize() {
	const maxDuration = 1 * time.Second
	const minDuration = 1000 * time.Millisecond / 120 // duration for 120 fps
	if opt.ThrottleDuration > maxDuration {
		opt.ThrottleDuration = maxDuration
	} else if opt.ThrottleDuration < minDuration {
		opt.ThrottleDuration = minDuration
	}
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

	binEncode := newParagraphListBinaryEncodeFunc(opt.MessageByteEncoding)
	if opt.EnableDebugTimestamp {
		// override binEncode to cotain DebugTimeStamp feature.
		orgBinEncode := binEncode
		binEncode = func(pl *pubdata.ParagraphList) ([]byte, error) {
			bs, err := orgBinEncode(pl)
			if err != nil {
				return nil, err
			}
			// debug timestamp should be first notified to UI than Publish so that listener can know
			// the timestamp to link to the next Publish.
			// Here, use latest paragraph Id as debug timestamp.
			if lastIdx := len(pl.Paragraphs) - 1; lastIdx >= 0 {
				now := time.Now()
				err = ui.OnDebugTimestamp(pl.Paragraphs[lastIdx].Id, now.Format(time.RFC3339Nano), now.UnixNano())
				if err != nil {
					return nil, err
				}
			}

			return bs, nil
		}
	}
	throttler := NewCallbackThrottler(ctx, opt.ThrottleDuration, ui, binEncode)
	throttler.StartThrottle()
	if err := editor.SetCallback(throttler); err != nil {
		return nil, err
	}
	return &uiAdapter{
		Printer:   editor,
		editor:    editor,
		throttler: throttler,
		mobileUI:  ui,
	}, nil
}

func (adp *uiAdapter) Close() error {
	eerr := adp.editor.Close()
	terr := adp.throttler.Close()
	return errors.Join(eerr, terr)
}

func newParagraphListBinaryEncodeFunc(encoding int) paragraphListBinaryEncoderFunc {
	var encoder paragraphListBinaryEncoderFunc
	switch encoding {
	case MessageByteEncodingProtobuf:
		encoder = func(p *pubdata.ParagraphList) ([]byte, error) {
			return p.MarshalVT()
		}
	case MessageByteEncodingJson:
		fallthrough
	default:
		encoder = func(p *pubdata.ParagraphList) ([]byte, error) {
			return json.Marshal(p)
		}
	}
	return encoder
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
	a.throttler.OnRequestChanged(req)
}
