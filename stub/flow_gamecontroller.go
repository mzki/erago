package stub

import (
	"fmt"

	"local/erago/uiadapter"
)

// implements flow.GameController
type flowGameController struct {
	uiadapter.UI
}

// return implements flow.GameController
func NewFlowGameController() *flowGameController {
	return &flowGameController{NewGameUIStub()}
}

func (ui flowGameController) RawInput() (string, error) {
	return "", nil
}
func (ui flowGameController) Command() (string, error) {
	return "", nil
}
func (ui flowGameController) CommandNumber() (int, error) {
	return 0, nil
}
func (ui flowGameController) CommandNumberRange(min, max int) (int, error) {
	return 0, nil
}
func (ui flowGameController) CommandNumberSelect(ns ...int) (int, error) {
	return 0, nil
}

func (ui flowGameController) Wait() error {
	return nil
}

func (ui flowGameController) PrintPlain(s string) {
	ui.Print(s)
}

func (ui flowGameController) PrintL(s string) {
	fmt.Println(s)
}

func (ui flowGameController) PrintC(s string, width int) {
	ui.Print(s)
}

func (ui flowGameController) PrintW(s string) error {
	ui.Print(s)
	return nil
}
func (ui flowGameController) PrintBar(int64, int64, int, string, string) {
	ui.Print("[#..]")
	return
}

func (ui flowGameController) TextBar(int64, int64, int, string, string) string {
	return "[#..]"
}
func (ui flowGameController) VPrint(vname, s string) error {
	ui.Print(s)
	return nil
}
func (ui flowGameController) VPrintL(vname, s string) error {
	ui.PrintL(s)
	return nil
}
func (ui flowGameController) VPrintC(vname, s string, width int) error {
	ui.PrintC(s, width)
	return nil
}
func (ui flowGameController) VPrintPlain(vname, s string) error {
	ui.PrintPlain(s)
	return nil
}
func (ui flowGameController) VPrintW(vname, s string) error {
	ui.PrintW(s)
	return nil
}
func (ui flowGameController) VPrintButton(vname, caption, cmd string) error {
	ui.PrintButton(caption, cmd)
	return nil
}
func (ui flowGameController) VPrintBar(vname string, now, max int64, width int, fg, bg string) error {
	ui.PrintBar(now, max, width, fg, bg)
	return nil
}
func (ui flowGameController) VPrintLine(vname, sym string) error {
	ui.PrintLine(sym)
	return nil
}
func (ui flowGameController) VClearLine(vname string, nline int) error {
	ui.ClearLine(nline)
	return nil
}
func (ui flowGameController) VClearLineAll(vname string) error {
	ui.ClearLineAll()
	return nil
}
func (ui flowGameController) VNewPage(vname string) error {
	ui.NewPage()
	return nil
}
func (ui flowGameController) SetSingleLayout(string) error                      { return nil }
func (ui flowGameController) SetHorizontalLayout(string, string, float64) error { return nil }
func (ui flowGameController) SetVerticalLayout(string, string, float64) error   { return nil }
