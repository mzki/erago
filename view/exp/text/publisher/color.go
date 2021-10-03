package publisher

import (
	"image/color"
)

// UInt32ColorToColorRGBA converts from uint32 type RGB color #0xRRGGBB to color.RGBA
// NOTE: A is not considered
func UInt32RGBToColorRGBA(c uint32) color.RGBA {
	Color := color.RGBA{}
	Color.R = uint8((c & 0x00ff0000) >> 16)
	Color.G = uint8((c & 0x0000ff00) >> 8)
	Color.B = uint8((c & 0x000000ff) >> 0)
	Color.A = 0xff // fixed
	return Color
}

// Int32ColorToColorRGBA converts from int32 type RGB color #0xRRGGBB to color.RGBA
// NOTE: A is not considered
func Int32RGBToColorRGBA(c int32) color.RGBA {
	return UInt32RGBToColorRGBA(uint32(c))
}

// Color RGBA converts from color.RGBA to uint32 type RGB color #0xRRGGBB.
// NOTE: A is not considered
func ColorRGBAToUInt32RGB(c color.RGBA) uint32 {
	var Color uint32 = 0
	Color = Color | (uint32(c.R) << 16)
	Color = Color | (uint32(c.G) << 8)
	Color = Color | (uint32(c.B) << 0)
	return Color
}

// Color RGBA converts from color.RGBA to int32 type RGB color #0xRRGGBB.
// NOTE: A is not considered
func ColorRGBAToInt32RGB(c color.RGBA) int32 {
	return int32(ColorRGBAToUInt32RGB(c))
}
