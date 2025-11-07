package publisher

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/golang/groupcache/lru"
	"github.com/mzki/erago/util/log"
	"github.com/mzki/erago/view/exp/text"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"golang.org/x/image/math/fixed"
)

// ImageBytesLoader loads image byte contents with different image fetch methods.
// It caches returned image bytes.
type ImageBytesLoader struct {
	cache   *lru.Cache
	loader  *text.TextScaleImageLoader
	encoder *png.Encoder

	fetchType pubdata.ImageFetchType
}

// DefaultCachedImageSize is used as a number of cached images.
var DefaultCachedImageSize = text.DefaultCachedImageSize

// ImageLoader itself cached result, so it is OK that
// internal cache has always 1 entry.
const internalImageLoaderCacheSize = 1

// NewImageBytesLoader create new instance.
// If cache size is less than or equal to 0, it will be treated as DefaultCachedImageSize
func NewImageBytesLoader(
	cacheSize int,
	fontSingleWidthPx, fontHeightPx fixed.Int26_6,
	fetchType pubdata.ImageFetchType,
) *ImageBytesLoader {
	if cacheSize <= 0 {
		cacheSize = DefaultCachedImageSize
	}
	return &ImageBytesLoader{
		cache: lru.New(cacheSize),
		loader: text.NewTextScaleImageLoader(
			internalImageLoaderCacheSize,
			fontHeightPx,
			fontSingleWidthPx,
		),
		// encoder instanciate later since it may not used.
		fetchType: fetchType,
	}
}

// ImageFetchResult is a result of ImageBytesLoader::LoadBytes()
// It contains image bytes, text-scale image size, px-scale image size
// and fetched type of bytes.
type ImageFetchResult struct {
	Bytes     []byte
	TsSize    text.TextScaleSize
	PxSize    image.Point
	FetchType pubdata.ImageFetchType
}

func (loader *ImageBytesLoader) LoadBytes(src string, widthInRW, heightInLC int) ImageFetchResult {
	cacheKey := loader.imageKey(src, widthInRW, heightInLC)
	if entry, ok := loader.cache.Get(cacheKey); ok {
		return entry.(ImageFetchResult)
	}
	// cache miss hit. create image fetch result.
	img, tsSize := loader.loadInternal(src, widthInRW, heightInLC)
	imgBytes, fetchType := loader.createImageBytes(img)
	pxSize := loader.loader.CalcImageSize(tsSize.Width, tsSize.Height)
	ret := ImageFetchResult{
		Bytes:     imgBytes,
		TsSize:    tsSize,
		PxSize:    pxSize,
		FetchType: fetchType,
	}
	loader.cache.Add(cacheKey, ret)
	return ret
}

func (loader *ImageBytesLoader) imageKey(src string, widthInRW, heightInLC int) lru.Key {
	return lru.Key(fmt.Sprintf("%s-%dx%d-%d", src, widthInRW, heightInLC, loader.fetchType))
}

func (loader *ImageBytesLoader) loadInternal(src string, widthInRW, heightInLC int) (image.Image, text.TextScaleSize) {
	imgData, tsSize, err := loader.loader.GetResized(src, widthInRW, heightInLC)
	if err != nil {
		log.Debugf("Failed to image load: %v", err)
		log.Debug("Replace to fallback image")

		// TODO: Replace fallback image

		// NOTE: Assumes width must be > 0, otherwise use constant width
		// NOTE: Use 1:1 aspect ratio black image now.
		if tsSize.Width == 0 {
			tsSize.Width = 10
		}
		tsSize.Height = tsSize.Width
		imgSize := loader.loader.CalcImageSize(tsSize.Width, tsSize.Height)
		imgData = image.NewRGBA(image.Rect(0, 0, imgSize.X, imgSize.Y))
	}

	return imgData, tsSize
}

func (loader *ImageBytesLoader) createImageBytes(img image.Image) ([]byte, pubdata.ImageFetchType) {
	switch loader.fetchType {
	case ImageFetchRawRGBA:
		if rgba, ok := img.(*image.RGBA); ok {
			return rgba.Pix, loader.fetchType
		} else {
			log.Debug("expect RGBA image internally but somohow not")
			return nil, ImageFetchNone
		}
	case ImageFetchEncodedPNG:
		if loader.encoder == nil {
			// lazy creation since other ImageFetchType is not needed the encoder.
			loader.encoder = &png.Encoder{CompressionLevel: png.BestSpeed}
		}
		buf := &bytes.Buffer{}
		err := loader.encoder.Encode(buf, img)
		if err != nil {
			log.Debugf("image encode failed: %v", err)
			return nil, ImageFetchNone
		}
		return buf.Bytes(), loader.fetchType
	case ImageFetchNone:
		fallthrough
	default:
		return nil, ImageFetchNone
	}
}
