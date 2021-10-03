package publisher

import "github.com/mzki/erago/view/exp/text/pubdata"

//go:generate mockgen -destination=./mock/mock_callback.go . Callback

// Callback defines callback interface.
// If callee somehow no longer handles the callback, return error to notify its status to the caller.
type Callback interface {
	// OnPublish is called when Paragraph is fixed by hard return (\n).
	OnPublish(*pubdata.Paragraph) error
	// OnPublishTemporary is called when Paragraph is NOT fixed yet by hard return(\n),
	// but required to show on UI.
	OnPublishTemporary(*pubdata.Paragraph) error
	// OnRemove is called when game thread requests to remove (N-1)-paragraphs which have been fixed
	// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary, thus to remove N-paragraphs.
	OnRemove(nParagraph int) error
	// OnRemoveAll is called when game thread requests to remove all paragraphs which have been fixed
	// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary.
	OnRemoveAll() error
}

// CallbackDefault implements Callback interface.
// User can only set override interface functions to its fields.
// Otherwise the functions not set fields are called to do nothing.
type CallbackDefault struct {
	OnPublishFunc          func(*pubdata.Paragraph) error
	OnPublishTemporaryFunc func(*pubdata.Paragraph) error
	OnRemoveFunc           func(nParagraph int) error
	OnRemoveAllFunc        func() error
}

// OnPublish is called when Paragraph is fixed by hard return (\n).
func (cb *CallbackDefault) OnPublish(p *pubdata.Paragraph) error {
	if f := cb.OnPublishFunc; f != nil {
		return f(p)
	}
	return nil
}

// OnPublishTemporary is called when Paragraph is NOT fixed yet by hard return(\n),
// but required to show on UI.
func (cb *CallbackDefault) OnPublishTemporary(p *pubdata.Paragraph) error {
	if f := cb.OnPublishTemporaryFunc; f != nil {
		return f(p)
	}
	return nil
}

// OnRemove is called when game thread requests to remove (N-1)-paragraphs which have been fixed
// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary, thus to remove N-paragraphs.
func (cb *CallbackDefault) OnRemove(nParagraph int) error {
	if f := cb.OnRemoveFunc; f != nil {
		return f(nParagraph)
	}
	return nil
}

// OnRemoveAll is called when game thread requests to remove all paragraphs which have been fixed
// by calling OnPublish and also temporal Paragraph by calling OnPublishTemporary.
func (cb *CallbackDefault) OnRemoveAll() error {
	if f := cb.OnRemoveAllFunc; f != nil {
		return f()
	}
	return nil
}
