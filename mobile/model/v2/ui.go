package model_v2

import (
	"github.com/mzki/erago/view/exp/text/pubdata"
)

// erago.UI for java bind.
type UI interface {
	Callback

	// it is called when mobile.app requires inputting
	// user's command.
	OnCommandRequested()

	// it is called when mobile.app requires just input any command.
	OnInputRequested()

	// it is called when mobile.app no longer requires any input,
	// such as just-input and command.
	OnInputRequestClosed()
}

// Re-define struct to export to gomobile.
type Paragraph pubdata.Paragraph

// Re-define publisher.Callback interface to export gomobile
type Callback interface {
	OnPublish(*Paragraph) error
	OnPublishTemporary(*Paragraph) error
	OnRemove(nParagraph int) error
	OnRemoveAll() error
}
