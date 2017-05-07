// package theme defines custom theme for UI.
package theme

import (
	"image"
	"image/color"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/math/fixed"
)

var (
	// Default Palette for painting, Material Dark Theme.
	DefaultPalette = theme.Palette{
		// Material Theme Dark 1, 2, 4 at https://material.google.com/style/color.html#color-themes
		theme.Dark:    image.Uniform{C: color.RGBA{0x00, 0x00, 0x00, 0xff}},
		theme.Neutral: image.Uniform{C: color.RGBA{0x21, 0x21, 0x21, 0xff}},
		theme.Light:   image.Uniform{C: color.RGBA{0x42, 0x42, 0x42, 0xff}},

		// Material LightBlue 400 for button text
		theme.Accent: image.Uniform{C: color.RGBA{0x29, 0xf6, 0xf6, 0xff}},

		// Material White to write primaly text
		theme.Foreground: image.Uniform{C: color.RGBA{0xff, 0xff, 0xff, 0xff}},

		// Material Gray 800~900, Theme Dark 3.
		theme.Background: image.Uniform{C: color.RGBA{0x30, 0x30, 0x30, 0xff}},
	}

	// Dark Divider color.
	DarkDivider = theme.StaticColor(color.RGBA{0xff, 0xff, 0xff, 0x1f})

	// Light Divider color.
	LightDivider = theme.StaticColor(color.RGBA{0x00, 0x00, 0x00, 0x1f})

	// Neutral Divider color.
	NeutralDivider = theme.StaticColor(color.RGBA{0xbd, 0xbd, 0xbd, 0xff})

	// Default theme for painting.
	Default = theme.Theme{
		FontFaceCatalog: defaultFontFaceCatalog,
		Palette:         &DefaultPalette,
	}

	defaultFontFaceCatalog theme.FontFaceCatalog
)

func init() {
	var err error
	defaultFontFaceCatalog, err = NewOneFontFaceCatalog(DefaultFontName, nil)
	if err != nil {
		panic("theme.init(): " + err.Error())
	}
	Default.FontFaceCatalog = defaultFontFaceCatalog
}

// UnitLh is 1 character cell height, same as sum of
// Ascent.Ceil and Descent.Ceil.
//
// if use this as unit.Value v, v.String() method and
// theme.Pixels(v) will be occured panic.
// use this package theme's UnitString() and Pixels()
// insteadly.
const UnitLh = unit.Unit(1 << 6)

// returns Lh as unit.Value.
func Lhs(f float64) unit.Value { return unit.Value{F: f, U: UnitLh} }

// Extension of theme.Pixels() which can treat Lh unit.Value.
// Otherwise, it returns same as theme.Pixels()
// Note that it call Theme.AcquireFontFace and ReleaseFontFace internaly.
func Pixels(t *theme.Theme, v unit.Value) fixed.Int26_6 {
	if v.U != UnitLh {
		return t.Pixels(v)
	}

	f := t.AcquireFontFace(theme.FontFaceOptions{})
	m := f.Metrics()
	h := m.Ascent + m.Descent
	t.ReleaseFontFace(theme.FontFaceOptions{}, f)

	return fixed.Int26_6(v.F * float64(h))
}
