package model

// erago.UI for java bind.
type UI interface {
	CallbackJson

	// it is called when mobile.app requires inputting
	// user's command.
	OnCommandRequested()

	// it is called when mobile.app requires just input any command.
	OnInputRequested()

	// it is called when mobile.app no longer requires any input,
	// such as just-input and command.
	OnInputRequestClosed()
}

// Callbacks with json message if use complex structure.
type CallbackJson interface {
	OnPublishJson(string) error
	OnPublishJsonTemporary(string) error
	OnRemove(nParagraph int) error
	OnRemoveAll() error
}
