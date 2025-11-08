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

	// OnDebugTimestamp is separated API. Its timestamp is linked to the publish by
	// publishId and Paragraph.Id in publish bytes.
	// This separation is considering backward compatibility of current API OnPublishByte to
	// avoid to extend new argument.
	// timestamp is RFC3379+nano precision, like "2006-01-02T15:04:05.999999999Z07:00".
	// epochTimeNano is epoch time a.k.a Unit time in nanoseond precision.
	OnDebugTimestamp(publishId int64, timestamp string, epochTimeNano int64) error
}
