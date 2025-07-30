package model

// erago.UI for java bind.
type UI interface {
	CallbackBytes

	// it is called when mobile.app requires inputting
	// user's command.
	OnCommandRequested()

	// it is called when mobile.app requires just input any command.
	OnInputRequested()

	// it is called when mobile.app no longer requires any input,
	// such as just-input and command.
	OnInputRequestClosed()
}

// Callbacks with binary message if use complex structure.
// Binary format used in OnPublishXXX() is depending on MessageByteEncoding in InitOptions.
type CallbackBytes interface {
	OnPublishBytes([]byte) error
	OnPublishBytesTemporary([]byte) error
	OnRemove(nParagraph int) error
	OnRemoveAll() error
}
