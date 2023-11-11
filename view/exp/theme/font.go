package theme

import (
	"os"
	"sync"

	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	internalfont "github.com/mzki/erago/view/exp/theme/internal/font"
)

const (
	DefaultFontSize = 12.0 // in pt
	DefaultDPI      = 72.0
)

// It is same as Options in golang.org/x/image/font/opentype
// but only DPI and Size fields exist to use easily.
type FontFaceOptions struct {
	// font size in point. if set 0 then use 12.0pt instead.
	Size float64

	// Dot per inch. if set 0 then use 72 DPI.
	DPI float64
	// TODO: more option?
}

// return reciever as opentype.Options.
// because nil TTF Options is ok, it may return nil if reciever is nil.
func (opt *FontFaceOptions) TTFOptions() *opentype.FaceOptions {
	var ttfOpt *opentype.FaceOptions
	if opt != nil {
		var setOpt FontFaceOptions = *opt
		if setOpt.Size == 0 {
			setOpt.Size = DefaultFontSize
		}
		if setOpt.DPI == 0 {
			setOpt.DPI = DefaultDPI
		}
		ttfOpt = &opentype.FaceOptions{
			Size: setOpt.Size,
			DPI:  setOpt.DPI,
			// TODO: more option?
		}
	} else {
		ttfOpt = &opentype.FaceOptions{
			Size: DefaultFontSize,
			DPI:  DefaultDPI,
			// TODO: more option?
		}
	}
	return ttfOpt
}

// DefaultFace returns this varints.
const DefaultFontName = "GenShinGothic-Monospace-Normal.ttf"

// return default monospace font face.
func NewDefaultFace(opt *FontFaceOptions) font.Face {
	f := defaulFont()
	face, err := opentype.NewFace(f, opt.TTFOptions())
	if err != nil {
		panic(err) // default must be succeeded
	}
	return face
}

func defaulFont() *opentype.Font {
	f, err := opentype.Parse(internalfont.MustAsset(DefaultFontName))
	if err != nil {
		panic("font: DefaultFont can not be parsed: " + err.Error())
	}
	return f
}

// parse truetype font file and return its font.Face.
func ParseTruetypeFileFace(file string, opt *FontFaceOptions) (font.Face, error) {
	f, err := parseTruetypeFile(file)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, opt.TTFOptions())
}

func parseTruetypeFile(file string) (*opentype.Font, error) {
	ttf, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return opentype.Parse(ttf)
}

// OneFontCatalog serves only one font object and its face.
type OneFontFaceCatalog struct {
	font *opentype.Font

	// Fields below are under mutex.
	mu   *sync.Mutex
	opt  *FontFaceOptions
	face font.Face
}

// return OneFontFaceCatalog which has only one font face specified by fontfile.
// return error if fontfile is not found
func NewOneFontFaceCatalog(fontfile string, o *FontFaceOptions) (*OneFontFaceCatalog, error) {
	var (
		f   *opentype.Font
		err error
	)
	if fontfile == DefaultFontName {
		f = defaulFont()
	} else {
		f, err = parseTruetypeFile(fontfile)
		if err != nil {
			return nil, err
		}
	}
	ttfOpt := o.TTFOptions()
	face, err := opentype.NewFace(f, ttfOpt)
	return &OneFontFaceCatalog{
		font: f,
		face: face,
		opt:  o,
		mu:   new(sync.Mutex),
	}, err
}

// update its font face using options.
// if opt has empty value. use previous value for that field.
func (cat *OneFontFaceCatalog) UpdateFontFaceOptions(opt *FontFaceOptions) {
	if opt.Size == 0 {
		opt.Size = cat.opt.Size
	}
	if opt.DPI == 0 {
		opt.DPI = cat.opt.DPI
	}

	cat.mu.Lock()
	defer cat.mu.Unlock()
	var err error
	cat.opt = opt
	cat.face, err = opentype.NewFace(cat.font, opt.TTFOptions())
	if err != nil {
		panic(err) // since it reuses previus font. should not fail
	}
}

// Implements theme.FontFaceCatalog interface.
// Return its font face with lock resource to prevent from using other goroutine.
// The argument theme.FontFaceOptions does not affect to returned font.Face.
// Be sure of calling ReleaseFontFace after use font.Face.
func (cat *OneFontFaceCatalog) AcquireFontFace(theme.FontFaceOptions) font.Face {
	cat.mu.Lock()
	return cat.face
}

// Implements theme.FontFaceCatalog interface
// Make the font.Face returned from AcquireFontFace free to use.
// With any arguments, the use freed font.Face will not be changed, got by AcquireFontFace.
// Be sure of calling This after call AcquireFontFace.
func (cat *OneFontFaceCatalog) ReleaseFontFace(theme.FontFaceOptions, font.Face) {
	cat.mu.Unlock()
}
