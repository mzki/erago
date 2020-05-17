package stub

import (
	"fmt"

	attr "github.com/mzki/erago/attribute"
)

type gameUIStub struct{}

func NewGameUIStub() *gameUIStub {
	return &gameUIStub{}
}

func (ui gameUIStub) Print(s string) error {
	_, err := fmt.Print(s)
	return err
}
func (ui gameUIStub) PrintLabel(s string) error {
	_, err := fmt.Print(s)
	return err
}
func (ui gameUIStub) PrintButton(caption, cmd string) error {
	_, err := fmt.Printf("[%s] %s", cmd, caption)
	return err
}

func (ui gameUIStub) PrintLine(sym string) error {
	_, err := fmt.Println(sym + sym + sym)
	return err
}
func (ui gameUIStub) SetColor(color uint32) error           { return nil }
func (ui gameUIStub) GetColor() (uint32, error)             { return 0x000000, nil }
func (ui gameUIStub) ResetColor() error                     { return nil }
func (ui gameUIStub) SetAlignment(a attr.Alignment) error   { return nil }
func (ui gameUIStub) GetAlignment() (attr.Alignment, error) { return attr.AlignmentLeft, nil }
func (ui gameUIStub) NewPage() error                        { return nil }
func (ui gameUIStub) ClearLine(nline int) error             { return nil }
func (ui gameUIStub) ClearLineAll() error                   { return nil }
func (ui gameUIStub) WindowLineCount() (int, error)         { return 0, nil }
func (ui gameUIStub) WindowRuneWidth() (int, error)         { return 0, nil }
func (ui gameUIStub) LineCount() (int, error)               { return 0, nil }
func (ui gameUIStub) CurrentRuneWidth() (int, error)        { return 0, nil }
func (ui gameUIStub) Sync() error                           { return nil }

// implement Layouter interface
func (gameUIStub) SetLayout(layout *attr.LayoutData) error { return nil }
func (gameUIStub) SetCurrentView(vname string) error       { return nil }
func (gameUIStub) GetCurrentViewName() string              { return "single" }
func (gameUIStub) GetViewNames() []string                  { return []string{"single"} }
