package scene

import (
	"context"
	"time"

	attr "github.com/mzki/erago/attribute"
)

// for convenience, attribute.Alignment are redeclared here,
// so that user need not to import attribute package explicitly.
const (
	AlignmentLeft   = attr.AlignmentLeft
	AlignmentCenter = attr.AlignmentCenter
	AlignmentRight  = attr.AlignmentRight
)

// io controller has interfaces for external io layer,
// inputting, outputting and layouting etc.
type IOController interface {
	InputPort
	OutputPort
	Layouter
}

// input interface.
type InputPort interface {
	// RawInput returns user input (i.e. key press) directly.
	// It waits until user input any key.
	RawInput() (string, error)

	// RawInputWithTimeout returns user input (i.e. key press) directly.
	// it can be canceled by canceling context.
	// This function timeouts with given timeout duration, and returns error context.DeadlineExceeded.
	RawInputWithTimeout(context.Context, time.Duration) (string, error)

	// Command returns string command which is emitted with user confirming.
	// It waits until user emit any command.
	Command() (string, error)

	// it returns string command which is emitted with user confirming.
	// it can be canceled by canceling context.
	// This function timeouts with given timeout duration, and returns error context.DeadlineExceeded.
	CommandWithTimeout(context.Context, time.Duration) (string, error)

	// Same as Command but return number command.
	// It waits until user emit any number command.
	CommandNumber() (int, error)

	// return number command limiting in range [min:max]
	// it can be canceled by canceling context.
	// This function timeouts with given timeout duration, and returns error context.DeadlineExceeded.
	CommandNumberWithTimeout(context.Context, time.Duration) (int, error)

	// return number command limiting in range [min:max]
	CommandNumberRange(ctx context.Context, min, max int) (int, error)

	// return number command matching to given candidates.
	CommandNumberSelect(context.Context, ...int) (int, error)

	// wait for any user confirming.
	// shorthand for Wait with context.Background.
	Wait() error

	// wait for any user confirming.
	// it can be canceled by canceling context.
	// This function timeouts with given timeout duration, and returns error context.DeadlineExceeded.
	WaitWithTimeout(ctx context.Context, timeout time.Duration) error
}

// output interface.
type OutputPort interface {
	// print text with parsing button pattern.
	// when text is matched to the pattern print text as button.
	Print(text string) error
	PrintL(text string) error // print the text added "\n" to end.
	PrintC(text string, width int) error
	PrintW(text string) error
	PrintButton(caption, command string) error
	PrintPlain(text string) error
	PrintLine(sym string) error
	PrintBar(now, max int64, width int, fg, bg string) error
	TextBar(now, max int64, width int, fg, bg string) (string, error)
	PrintImage(file string, widthInRW, heightInLC int) error
	MeasureImageSize(file string, widthInRW, heightInLC int) (width, height int, err error)
	PrintSpace(widthInRW int) error

	NewPage() error
	ClearLine(nline int) error
	ClearLineAll() error

	// Prefixed V functions perform same as not V-prefixed functions.
	// But difference is that targeting for the view frame specified by name and
	// return error of such view frame is not found.
	VPrint(vname, text string) error
	VPrintL(vname, text string) error
	VPrintC(vname, text string, width int) error
	VPrintW(vname, text string) error
	VPrintButton(vname, caption, command string) error
	VPrintPlain(vname, text string) error
	VPrintLine(vname, sym string) error
	VPrintBar(vname string, now, max int64, width int, fg, bg string) error

	VNewPage(vname string) error
	VClearLine(vname string, nline int) error
	VClearLineAll(vname string) error

	CurrentRuneWidth() (int, error) // rune width of currently editing line.
	LineCount() (int, error)        // line count as it increases outputting new line
	WindowRuneWidth() (int, error)  // max rune width for view width.
	WindowLineCount() (int, error)  // max line count for view height.

	// output options
	SetColor(color uint32) error
	GetColor() (uint32, error)
	ResetColor() error

	GetAlignment() (attr.Alignment, error)
	SetAlignment(attr.Alignment) error
}

// layouter layouts output screen.
// screen is divided by unit named view.
//
// if it is not implement actually,
// ViewPrinter's functions should be handled correctry.
type Layouter interface {
	// set new layout. since it is low level function, use SetSingleLayout,
	// SetHorizontalLayout and SetVerticalLayout to layout simply.
	//
	// More detail is in erago/attribute package.
	SetLayout(layout *attr.LayoutData) error

	// set single view layout. if name is already exist, text in the view is kept.
	SetSingleLayout(name string) error

	// TODO: { Horizontal | Vertical }Layout splits current view?
	// which give more flexibilty for layouting.

	// split screen into left and right views which have unique name.
	// rate is separator position on entire screen.
	// if name is already exist in views, text in the view is kept.
	SetHorizontalLayout(vname1, vname2 string, rate float64) error

	// split screen into upper and bottom views which have unique name.
	// if name is already exist in views, text in the view is kept.
	SetVerticalLayout(vname1, vname2 string, rate float64) error

	// set default output view by view name.
	SetCurrentView(vname string) error

	// return default output view name.
	GetCurrentViewName() string

	// return existing views name
	GetViewNames() []string
}
