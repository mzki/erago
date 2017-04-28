package ui

import (
	"image"
	"image/color"
	icondraw "image/draw"
	"path/filepath"

	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/draw"

	"local/erago/util/log"
	limage "local/erago/view/exp/image"
)

// ImageView is a leaf widget that paints image.Image.
// the painted image is fitted to its widget size so
// that entire view of image is shown.
type ImageView struct {
	node.LeafEmbed

	scaledSrc draw.Image
	scale     float64

	src      image.Image
	filename string

	// it is used to paint error icon.
	// it is nil if src exists, other wise it exists.
	z *iconvg.Rasterizer
}

var globalImagePool = limage.NewPool(limage.DefaultPoolSize)

func NewImageView(filename string) *ImageView {
	var z *iconvg.Rasterizer
	m, err := globalImagePool.Get(filename)
	if err != nil {
		// can not get image resource, fallback to paint error icon.
		log.Infof("Loading Image(%s) FAIL: %v", filename, err)
		m = nil
		z = new(iconvg.Rasterizer)
	}
	v := &ImageView{
		scaledSrc: nil,
		scale:     1.0,
		src:       m,
		filename:  filename,
		z:         z,
	}
	v.Wrapper = v
	return v
}

func (v *ImageView) SrcName() string {
	return filepath.Base(v.filename)
}

func (v *ImageView) iconX(t *theme.Theme) int {
	return t.Pixels(unit.DIPs(200)).Round()
}

func (v *ImageView) Measure(t *theme.Theme, widthHint, heightHint int) {
	if widthHint < 0 {
		widthHint = 0
	}
	if heightHint < 0 {
		heightHint = 0
	}
	// its size depends on parent's measuring.
	v.MeasuredSize = image.Point{widthHint, heightHint}
}

func (v *ImageView) Layout(t *theme.Theme) {
	src := v.src
	if src == nil {
		v.LeafEmbed.Layout(t)
		return
	}

	vSize := v.Rect.Size()
	sSize := src.Bounds().Size()
	vAspect := float64(vSize.X) / float64(vSize.X)
	sAspect := float64(sSize.X) / float64(sSize.X)
	if vAspect >= sAspect {
		// view's rect is wider than src's one.
		// it is not over the view's rect that fitting height of src's rect to rect's height.
		v.scale = float64(vSize.Y) / float64(sSize.Y)
	} else {
		// view's rect is taller than src's one.
		v.scale = float64(vSize.X) / float64(sSize.X)
	}
	v.Mark(node.MarkNeedsPaintBase)
}

func (v *ImageView) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	// TODO:
	// if !v.Marks.NeedsPaintBase() {
	// 	return nil
	// }
	v.Marks.UnmarkNeedsPaintBase()

	if v.Rect.Empty() {
		return nil
	}
	if v.src == nil {
		return v.paintErrorIcon(ctx, origin)
	}

	// when change view size, needs refresh scaled image src.
	sSize := v.src.Bounds().Size()
	newScSize := image.Point{
		X: int(float64(sSize.X) * v.scale),
		Y: int(float64(sSize.Y) * v.scale),
	}
	if v.scaledSrc != nil && !v.scaledSrc.Bounds().Size().Eq(newScSize) {
		v.scaledSrc = nil
	}

	if v.scaledSrc == nil {
		sr := image.Rectangle{Max: newScSize}
		v.scaledSrc = image.NewRGBA(sr)
		draw.NearestNeighbor.Scale(v.scaledSrc, sr, v.src, v.src.Bounds(), draw.Src, nil)
	}

	// fill background color
	vRect := v.Rect.Add(origin)
	draw.Draw(ctx.Dst, vRect, theme.Background.Uniform(ctx.Theme), image.Point{}, draw.Src)

	// move to vRect's coordinate space so that scRect.Min is aligned to
	// vRect's one.
	scRect := image.Rectangle{Max: newScSize}.Add(vRect.Min)
	// move to center
	sz := vRect.Size()
	if d := (sz.X - sz.Y) / 2; d > 0 {
		scRect.Min.X += d
		scRect.Max.X += d
	} else if d < 0 {
		scRect.Min.Y += d
		scRect.Max.Y += d
	}
	// draw a scaled image to destinaition.
	draw.Draw(ctx.Dst, vRect.Intersect(scRect), v.scaledSrc, image.Point{}, draw.Src)
	return nil
}

var (
	errorIcon = icons.AlertErrorOutline

	// since iconvg.DefaultPalette is not a pointer, modification of this is not shared.
	iconvgPalette = iconvg.DefaultPalette
)

func (v *ImageView) paintErrorIcon(ctx *node.PaintBaseContext, origin image.Point) error {
	wr := v.Rect.Add(origin)
	draw.Draw(ctx.Dst, wr, theme.Background.Uniform(ctx.Theme), image.Point{}, draw.Src)

	iconX := ctx.Theme.Pixels(unit.DIPs(200)).Round()
	sz := wr.Size()
	// inset so that wr fits iconX with place at center.
	var inset int
	if sz.X-sz.Y > 0 {
		inset = (sz.Y - iconX) / 2
	} else {
		inset = (sz.X - iconX) / 2
	}
	if inset < 0 {
		return nil
	}
	wr = wr.Inset(inset)

	// fill rest space arisen by not squared rect.
	// sz is updated since wr is updated.
	sz = wr.Size()
	if d := sz.X - sz.Y; d > 0 {
		wr.Min.X += d / 2
		wr.Max.X = wr.Min.X + sz.Y
	} else if d < 0 {
		wr.Min.Y -= d / 2
		wr.Max.Y = wr.Min.Y + sz.X
	}
	fgColor := theme.Foreground.Color(ctx.Theme).(color.RGBA)
	iconvgPalette[0] = fgColor

	v.z.SetDstImage(ctx.Dst, wr, icondraw.Over)
	return iconvg.Decode(v.z, errorIcon, &iconvg.DecodeOptions{
		Palette: &iconvgPalette,
	})
}
