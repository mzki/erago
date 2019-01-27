package uiadapter

import (
	"local/erago/attribute"
)

// Interfaces for the presentaion layer.
type UI interface {
	Printer
	Layouter
}

// output interface.
// note that these functions may be called asynchronously.
type Printer interface {
	Syncer // implements Syncer interface

	// Print text to screen.
	// It should implement moving to next line by "\n".
	Print(s string) error

	// Print label text to screen.
	// It should not be separated in wrapping text.
	PrintLabel(s string) error

	// Print Clickable button text. it shows caption on screen and emits command
	// when it is selected. It is no separatable in wrapping text.
	PrintButton(caption, command string) error

	// Print Line using sym.
	// given sym #, output line is: ############...
	PrintLine(sym string) error

	// Set and Get Color using 0xRRGGBB for 24bit color
	SetColor(color uint32) error
	GetColor() (color uint32, err error)
	ResetColor() error

	// Set and Get Alignment
	SetAlignment(attribute.Alignment) error
	GetAlignment() (attribute.Alignment, error)

	// skip current lines to display none.
	// TODO: is it needed?
	NewPage() error

	// Clear lines specified number.
	ClearLine(nline int) error

	// Clear all lines containing historys.
	ClearLineAll() error

	// max rune width to fill the view width.
	MaxRuneWidth() (int, error)

	// current rune width in the editting line.
	CurrentRuneWidth() (int, error)

	// line count to fill the view heght.
	LineCount() (int, error)
}

// Syncer is a interface for synchronizing output and display state.
type Syncer interface {
	// Sync flushes any pending output result, PrintXXX or ClearLine,
	// at UI implementor. It can also use rate limitting for PrintXXX functions.
	Sync() error
}

// Layouting interface. it should be implemented to
// build multiple window user interface.
// These functions are called asynchronously.
type Layouter interface {
	// set new layout acording to attribute.LayoutData.
	// it may return error if LayoutData is invalid.
	//
	// More details for LayoutData structure is in erago/attribute package.
	SetLayout(layout *attribute.LayoutData) error

	// set default output view by view name.
	// Printer's functions will output to a default view.
	// it may return error if vname is not found.
	SetCurrentView(vname string) error

	// return default output view name.
	GetCurrentViewName() string

	// return existing views name in multiple layout.
	GetViewNames() []string
}

// SingleUI implements partial UI interface, Layouter.
// Printer interface is injected by user to build complete
// UI interface.
//
// Thus, you can implement only Printer interface
// for complete UI interface:
//
//   UI = SingleUI{implements_only_printer}
//
type SingleUI struct {
	Printer
}

func (ui SingleUI) SetLayout(*attribute.LayoutData) error { return nil }
func (ui SingleUI) SetCurrentView(vname string) error     { return nil }
func (ui SingleUI) GetCurrentViewName() string            { return "single" }
func (ui SingleUI) GetViewNames() []string                { return []string{"single"} }
