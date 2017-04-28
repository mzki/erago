package stub

import (
	"fmt"

	"local/erago/attribute"
	"local/erago/uiadapter"
)

type gameUIStub struct{}

func NewGameUIStub() uiadapter.UI {
	return uiadapter.SingleUI{&gameUIStub{}}
}

func (ui gameUIStub) Print(s string) {
	fmt.Print(s)
}
func (ui gameUIStub) PrintLabel(s string) {
	fmt.Print(s)
}
func (ui gameUIStub) PrintButton(caption, cmd string) {
	fmt.Printf("[%s] %s", cmd, caption)
}
func (ui gameUIStub) PrintLine(sym string) {
	fmt.Println(sym + sym + sym)
}
func (ui gameUIStub) SetColor(color uint32)              {}
func (ui gameUIStub) GetColor() uint32                   { return 0x000000 }
func (ui gameUIStub) ResetColor()                        {}
func (ui gameUIStub) SetAlignment(a attribute.Alignment) {}
func (ui gameUIStub) GetAlignment() attribute.Alignment  { return attribute.AlignmentLeft }
func (ui gameUIStub) NewPage()                           {}
func (ui gameUIStub) ClearLine(nline int)                {}
func (ui gameUIStub) ClearLineAll()                      {}
func (ui gameUIStub) LineCount() int                     { return 0 }
func (ui gameUIStub) CurrentRuneWidth() int              { return 0 }
func (ui gameUIStub) MaxRuneWidth() int                  { return 0 }
