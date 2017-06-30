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
	Print(s string)
	PrintLabel(s string)
	PrintButton(caption, command string)
	PrintLine(sym string)
	SetColor(c int32)
	GetColor() int32
	ResetColor()
	SetAlignmentLeft()
	SetAlignmentCenter()
	SetAlignmentRight()
	NewPage()
	ClearLine(nline int32)
	ClearLineAll()
	MaxLineWidth() int32
	CurrentLineWidth() int32
	LineCount() int32
}
