package model

// erago.UI for java bind.
type UI interface {
	Printer

	// it is called when mobile.app requires inputting
	// user's command.
	OnCommandRequested()

	// it is called when mobile.app requires just input any command.
	OnInputRequested()

	// it is called when mobile.app no longer requires any input,
	// such as just-input and command.
	OnInputRequestClosed()
}

// Printer is interface for the printing content to
// the view. Any functions of this are called asynchronously.
type Printer interface {
	Print(s string) error
	PrintLabel(s string) error
	PrintButton(caption, command string) error
	PrintLine(sym string) error
	SetColor(c int32) error
	GetColor() (int32, error)
	ResetColor() error
	SetAlignmentLeft() error
	SetAlignmentCenter() error
	SetAlignmentRight() error
	NewPage() error
	ClearLine(nline int32) error
	ClearLineAll() error
	MaxLineWidth() (int32, error)
	CurrentLineWidth() (int32, error)
	LineCount() (int32, error)
}
