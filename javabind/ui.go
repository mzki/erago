package eragoj

// erago.UI for java bind.
type UI interface {
	Printer
	ViewPrinter
	Layouter
}

type Printer interface {
	Print(s string)
	PrintLabel(s string, width int32)
	PrintButton(caption, command string)
	PrintLine(sym string)
	PrintBar(now, max int64, width int32)
	SetColor(c int32)
	GetColor() int32
	ResetColor()
	SetAlignment(alignment int8)
	GetAlignment() (alignment int8)
	NewPage()
	ClearLine(nline int32)
	ClearScreen()
}

type ViewPrinter interface {
	VPrint(vname, s string) error
	VPrintLabel(vname, s string, width int32) error
	VPrintButton(vname, caption, command string) error
	VPrintLine(vname, s string) error
	VPrintBar(vname string, now, max int64, width int32) error

	VSetColor(vname string, c int32) error
	VGetColor(vname string) (int32, error)
	VResetColor(vname string) error
	VSetAlignment(vname string, alignment int8) error
	VGetAlignment(vname string) (alignment int8, err error)
	VNewPage(vname string) error
	VClearLine(vname string, nline int32) error
	VClearScreen(vname string) error
}

type Layouter interface {
	SetSingleLayout(name string) error
	SetHorizontalLayout(vname1, vname2 string, rate float64) error
	SetVerticalLayout(vname1, vname2 string, rate float64) error
	SetCurrentView(vname string) error
	GetCurrentViewName() string
	GetViewNames() string
}

type singleUI struct {
	Printer
}

func (ui singleUI) VPrint(vname, s string) error {
	ui.Printer.Print(s)
	return nil
}
func (ui singleUI) VPrintLabel(vname, s string, width int32) error {
	ui.Printer.PrintLabel(s, width)
	return nil
}
func (ui singleUI) VPrintButton(vname, caption, command string) error {
	ui.Printer.PrintButton(caption, command)
	return nil
}
func (ui singleUI) VPrintLine(vname, s string) error {
	ui.Printer.PrintLine(s)
	return nil
}
func (ui singleUI) VPrintBar(vname string, now, max int64, width int32) error {
	ui.Printer.PrintBar(now, max, width)
	return nil
}
func (ui singleUI) VSetColor(vname string, c int32) error {
	ui.Printer.SetColor(c)
	return nil
}
func (ui singleUI) VGetColor(vname string) (int32, error) {
	return ui.Printer.GetColor(), nil
}
func (ui singleUI) VResetColor(vname string) error {
	ui.Printer.ResetColor()
	return nil
}
func (ui singleUI) VSetAlignment(vname string, align int8) error {
	ui.Printer.SetAlignment(align)
	return nil
}
func (ui singleUI) VGetAlignment(vname string) (int8, error) {
	return ui.Printer.GetAlignment(), nil
}
func (ui singleUI) VNewPage(vname string) error {
	ui.Printer.NewPage()
	return nil
}
func (ui singleUI) VClearLine(vname string, nline int32) error {
	ui.Printer.ClearLine(nline)
	return nil
}
func (ui singleUI) VClearScreen(vname string) error {
	ui.Printer.ClearScreen()
	return nil
}

func (ui singleUI) SetSingleLayout(name string) error                             { return nil }
func (ui singleUI) SetVerticalLayout(vname1, vname2 string, rate float64) error   { return nil }
func (ui singleUI) SetHorizontalLayout(vname1, vname2 string, rate float64) error { return nil }
func (ui singleUI) SetCurrentView(vname string) error                             { return nil }
func (ui singleUI) GetCurrentViewName() string                                    { return "single" }
func (ui singleUI) GetViewNames() string                                          { return "single" }
