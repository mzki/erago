package publisher

import (
	"image/color"
)

// UIntColorToColorRGBA converts from uint32 type RGB color #0xRRGGBB to color.RGBA
// NOTE: A is not considered
func UIntRGBToColorRGBA(c uint32) color.RGBA {
	Color := color.RGBA{}
	Color.R = uint8((c & 0x00ff0000) >> 16)
	Color.G = uint8((c & 0x0000ff00) >> 8)
	Color.B = uint8((c & 0x000000ff) >> 0)
	Color.A = 0xff // fixed
	return Color
}

// Color RGBA converts from color.RGBA to uint32 type RGB color #0xRRGGBB.
// NOTE: A is not considered
func ColorRGBAToUIntRGB(c color.RGBA) uint32 {
	var Color uint32 = 0
	Color = Color | (uint32(c.R) << 16)
	Color = Color | (uint32(c.G) << 8)
	Color = Color | (uint32(c.B) << 0)
	return Color
}
