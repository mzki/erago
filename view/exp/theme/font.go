package theme

import (
	"io/ioutil"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"

	internalfont "github.com/mzki/erago/view/exp/theme/internal/font"
)

// It is same as Options in github.com/golang/freetype/truetype.go
// but only DPI and Size fields exist to use easily.
type FontFaceOptions struct {
	// font size in point. if set 0 then use 12.0pt instead.
	Size float64

	// Dot per inch. if set 0 then use 72 DPI.
	DPI float64
	// TODO: more option?
}

// return reciever as truetype.Options.
// because nil TTF Options is ok, it may return nil if reciever is nil.
func (opt *FontFaceOptions) TTFOptions() *truetype.Options {
	var ttfOpt *truetype.Options
	if opt != nil {
		ttfOpt = &truetype.Options{
			Size: opt.Size,
			DPI:  opt.DPI,
			// TODO: more option?
		}
	}
	return ttfOpt
}

// DefaultFace returns this varints.
const DefaultFontName = "GenShinGothic-Monospace-Normal.ttf"

// return default monospace font face.
func NewDefaultFace(opt *FontFaceOptions) font.Face {
	f := defaultTruetypeFont()
	return truetype.NewFace(f, opt.TTFOptions())
}

func defaultTruetypeFont() *truetype.Font {
	f, err := truetype.Parse(internalfont.MustAsset(DefaultFontName))
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
	return truetype.NewFace(f, opt.TTFOptions()), nil
}

func parseTruetypeFile(file string) (*truetype.Font, error) {
	ttf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return truetype.Parse(ttf)
}

// OneFontCatalog serves only one font object and its face.
type OneFontFaceCatalog struct {
	font *truetype.Font

	// Fields below are under mutex.
	mu   *sync.Mutex
	opt  *FontFaceOptions
	face font.Face
}

// return OneFontFaceCatalog which has only one font face specified by fontfile.
// return error if fontfile is not found
func NewOneFontFaceCatalog(fontfile string, o *FontFaceOptions) (*OneFontFaceCatalog, error) {
	var (
		f   *truetype.Font
		err error
	)
	if fontfile == DefaultFontName {
		f = defaultTruetypeFont()
	} else {
		f, err = parseTruetypeFile(fontfile)
		if err != nil {
			return nil, err
		}
	}
	ttfOpt := o.TTFOptions()
	face := truetype.NewFace(f, ttfOpt)
	return &OneFontFaceCatalog{
		font: f,
		face: face,
		opt:  o,
		mu:   new(sync.Mutex),
	}, nil
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
	cat.opt = opt
	cat.face = truetype.NewFace(cat.font, opt.TTFOptions())
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
