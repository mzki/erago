package stub

import (
	"context"
	"fmt"
	"time"
)

// implements scene.IOController
type sceneIOController struct {
	*gameUIStub
}

// return implements scene.IOController
func NewFlowGameController() *sceneIOController {
	return &sceneIOController{NewGameUIStub()}
}

func (ui sceneIOController) RawInput() (string, error) {
	return "", nil
}
func (ui sceneIOController) RawInputWithTimeout(context.Context, time.Duration) (string, error) {
	return "", nil
}
func (ui sceneIOController) Command() (string, error) {
	return "", nil
}
func (ui sceneIOController) CommandWithTimeout(context.Context, time.Duration) (string, error) {
	return "", nil
}
func (ui sceneIOController) CommandNumber() (int, error) {
	return 0, nil
}
func (ui sceneIOController) CommandNumberWithTimeout(context.Context, time.Duration) (int, error) {
	return 0, nil
}
func (ui sceneIOController) CommandNumberRange(ctx context.Context, min, max int) (int, error) {
	return 0, nil
}
func (ui sceneIOController) CommandNumberSelect(ctx context.Context, ns ...int) (int, error) {
	return 0, nil
}

func (ui sceneIOController) Wait() error {
	return nil
}
func (ui sceneIOController) WaitWithTimeout(context.Context, time.Duration) error {
	return nil
}

func (ui sceneIOController) PrintPlain(s string) error {
	ui.Print(s)
	return nil
}

func (ui sceneIOController) PrintL(s string) error {
	_, err := fmt.Println(s)
	return err
}

func (ui sceneIOController) PrintC(s string, width int) error {
	ui.Print(s)
	return nil
}

func (ui sceneIOController) PrintW(s string) error {
	ui.Print(s)
	return nil
}
func (ui sceneIOController) PrintBar(int64, int64, int, string, string) error {
	ui.Print("[#..]")
	return nil
}

func (ui sceneIOController) TextBar(int64, int64, int, string, string) (string, error) {
	return "[#..]", nil
}
func (ui sceneIOController) VPrint(vname, s string) error {
	ui.Print(s)
	return nil
}
func (ui sceneIOController) VPrintL(vname, s string) error {
	ui.PrintL(s)
	return nil
}
func (ui sceneIOController) VPrintC(vname, s string, width int) error {
	ui.PrintC(s, width)
	return nil
}
func (ui sceneIOController) VPrintPlain(vname, s string) error {
	ui.PrintPlain(s)
	return nil
}
func (ui sceneIOController) VPrintW(vname, s string) error {
	ui.PrintW(s)
	return nil
}
func (ui sceneIOController) VPrintButton(vname, caption, cmd string) error {
	ui.PrintButton(caption, cmd)
	return nil
}
func (ui sceneIOController) VPrintBar(vname string, now, max int64, width int, fg, bg string) error {
	ui.PrintBar(now, max, width, fg, bg)
	return nil
}
func (ui sceneIOController) VPrintLine(vname, sym string) error {
	ui.PrintLine(sym)
	return nil
}
func (ui sceneIOController) VClearLine(vname string, nline int) error {
	ui.ClearLine(nline)
	return nil
}
func (ui sceneIOController) VClearLineAll(vname string) error {
	ui.ClearLineAll()
	return nil
}
func (ui sceneIOController) VNewPage(vname string) error {
	ui.NewPage()
	return nil
}
func (ui sceneIOController) SetSingleLayout(string) error                      { return nil }
func (ui sceneIOController) SetHorizontalLayout(string, string, float64) error { return nil }
func (ui sceneIOController) SetVerticalLayout(string, string, float64) error   { return nil }
