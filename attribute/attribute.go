// package attribute contains property for view interface.
package attribute

import ()

// attribute set
type Attributes struct {
	Alignment
	Colors
}

// text alignment
type Alignment int8

const (
	AlignmentLeft Alignment = iota
	AlignmentCenter
	AlignmentRight
)

// Colors has foreground and background 16bit color as like 0xRRGGBB.
type Colors struct {
	Fg, Bg uint32
}

const (
	DefaultColorFg = 0xffffff // white
	DefaultColorBg = 0x000000 // black
)

// default colors.
var DefaultColors = Colors{DefaultColorFg, DefaultColorBg}

// check both fore and back color is default?
func (c Colors) IsDefault() bool {
	return c.Fg == DefaultColorFg && c.Bg == DefaultColorBg
}

// get hex color of given name. if not found return black
func ColorOfName(name string) (uint32, bool) {
	col, has := HTMLColorTable[name]
	if has {
		return col, true
	}
	return HTMLColorTable["black"], false
}
