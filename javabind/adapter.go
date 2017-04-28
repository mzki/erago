package eragoj

import (
	"strings"

	"local/erago/attribute"
)

// adapter for erago.UI interface.
// it converts gomobile bind interface to erago.UI.
type uiAdapter struct {
	UI
}

func (a uiAdapter) PrintLabel(s string, width int) {
	a.UI.PrintLabel(s, int32(width))
}

func (a uiAdapter) VPrintLabel(vname, s string, width int) error {
	return a.UI.VPrintLabel(vname, s, int32(width))
}

func (a uiAdapter) PrintBar(now, max int64, width int) {
	a.UI.PrintBar(now, max, int32(width))
}

func (a uiAdapter) VPrintBar(vname string, now, max int64, width int) error {
	return a.UI.VPrintBar(vname, now, max, int32(width))
}

func (a uiAdapter) ClearLine(n int) {
	a.UI.ClearLine(int32(n))
}

func (a uiAdapter) VClearLine(vname string, n int) error {
	return a.UI.VClearLine(vname, int32(n))
}

func (a uiAdapter) SetColor(color uint32) {
	a.UI.SetColor(int32(color))
}

func (a uiAdapter) VSetColor(vname string, color uint32) error {
	return a.UI.VSetColor(vname, int32(color))
}

func (a uiAdapter) GetColor() uint32 {
	return uint32(a.UI.GetColor())
}

func (a uiAdapter) VGetColor(vname string) (uint32, error) {
	c, err := a.UI.VGetColor(vname)
	return uint32(c), err
}

func (a uiAdapter) SetAlignment(align attribute.Alignment) {
	a.UI.SetAlignment(toAlignmentInt8(align))
}

func (a uiAdapter) VSetAlignment(vname string, align attribute.Alignment) error {
	return a.UI.VSetAlignment(vname, toAlignmentInt8(align))
}

func (a uiAdapter) GetAlignment() attribute.Alignment {
	return toAlignment(a.UI.GetAlignment())
}

func (a uiAdapter) VGetAlignment(vname string) (attribute.Alignment, error) {
	align, err := a.UI.VGetAlignment(vname)
	return toAlignment(align), err
}

const ViewNamesSeparator = ";"

func (a uiAdapter) GetViewNames() []string {
	return strings.Split(a.UI.GetViewNames(), ViewNamesSeparator)
}
