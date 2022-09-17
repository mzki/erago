package text

import (
	"image"

	egimg "github.com/mzki/erago/view/exp/image"
	"golang.org/x/image/math/fixed"
)

// TextScaleSize is size represented by text-scale rune width and line count.
type TextScaleSize struct {
	Width  int
	Height int
}

// TextScaleImageLoader is image loader which accepts
// text scale image size, such as rune width and line count for loading options.
// The loaded image is cached internally.
type TextScaleImageLoader struct {
	loader *egimg.Loader

	fontHeight      fixed.Int26_6
	fontSingleWidth fixed.Int26_6
}

// DefaultCachedImageSize is a default value for a cached image size.
var DefaultCachedImageSize = egimg.DefaultCacheSize

// NewTextScaleImageLoader create new instance.
func NewTextScaleImageLoader(cachedImageSize int, fontHeight, fontSingleWidth fixed.Int26_6) *TextScaleImageLoader {
	return &TextScaleImageLoader{
		loader:          egimg.NewLoader(cachedImageSize),
		fontHeight:      fontHeight,
		fontSingleWidth: fontSingleWidth,
	}
}

// CalcImageSize converts text scale size into pixel scale size by using pixel scale size of the font.
func (l *TextScaleImageLoader) CalcImageSize(resizedWidthInRW, resizedHeightInLC int) image.Point {
	return image.Point{
		X: int26_6_Mul(l.fontSingleWidth, fixed.I(resizedWidthInRW)).Ceil(),
		Y: int26_6_Mul(l.fontHeight, fixed.I(resizedHeightInLC)).Ceil(),
	}
}

func (l *TextScaleImageLoader) FontHeight() fixed.Int26_6      { return l.fontHeight }
func (l *TextScaleImageLoader) FontSingleWidth() fixed.Int26_6 { return l.fontSingleWidth }

// GetResized get resized image and text-scaled image size from image loader. The result is cached
// for combination of the arguments. If both of resizedXXX is zero, the result
// is not resized. If either of resizedXXX is zero, the resized size of that is
// auto-filled by the other with keep aspect ratio of original image size.
func (l *TextScaleImageLoader) GetResized(file string, resizedWidthInRW, resizedHeightInLC int) (image.Image, TextScaleSize, error) {
	opt := egimg.LoadOptions{ /* Empty means no resized */ }
	if resizedWidthInRW != 0 || resizedHeightInLC != 0 {
		opt.ResizedSize = l.CalcImageSize(resizedWidthInRW, resizedHeightInLC)
	}
	img, err := l.loader.GetWithOptions(file, opt)
	if err != nil {
		return nil, TextScaleSize{}, err
	}
	twSize := l.calcTextScaleSize(img.Bounds().Size())
	return img, twSize, nil
}

func (l *TextScaleImageLoader) calcTextScaleSize(imgSize image.Point) TextScaleSize {
	rwF := int26_6_Div(fixed.I(imgSize.X), l.fontSingleWidth)
	lcF := int26_6_Div(fixed.I(imgSize.Y), l.fontHeight)
	return TextScaleSize{
		Width:  rwF.Ceil(),
		Height: lcF.Ceil(),
	}
}
