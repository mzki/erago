package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/mzki/erago/filesystem"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

// Loader holds image caches.
// concurrent use is OK.
type Loader struct {
	mu    *sync.Mutex
	cache *lru.Cache // under mutex because cache is not safe for concurrently.
}

const DefaultCacheSize = 10

// Loader has cachedSize cache entries. the oldest image
// is removed from cache using LRU.
// use DefaultCacheSize if cachedSize <= 0.
func NewLoader(cachedSize int) *Loader {
	if cachedSize <= 0 {
		cachedSize = DefaultCacheSize
	}
	return &Loader{
		mu:    new(sync.Mutex),
		cache: lru.New(cachedSize),
	}
}

// get cached image by using image's file name.
// if not found, load image data from file
// and return loaded image with loading error.
// error nil means loaded image found.
func (l *Loader) Get(file string) (image.Image, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.cache.Get(file)
	if ok {
		return v.(image.Image), nil
	}

	// new arrival key, construct new image.
	m, err := AutoLoadFile(file)
	if err != nil {
		return nil, err
	}
	l.cache.Add(file, m)
	return m, nil
}

// auto detect file extension, png, jpeg, and jpg, and
// return loaded image data with error.
// error contains file not found, unsupported extension, ... etc.
func AutoLoadFile(file string) (image.Image, error) {
	ext := filepath.Ext(file)
	if len(ext) == 0 {
		return nil, fmt.Errorf("file must have the extension like .png, .jpeg, or .jpg")
	}
	ext = ext[1:] // remove first characater "."

	fp, err := filesystem.Load(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return AutoLoad(fp, ext)
}

// load image data from io.Reader r with its image type, png, jpeg, and jpg, and
// return loaded image with error.
// error contains unsupported extension.
func AutoLoad(r io.Reader, ext string) (image.Image, error) {
	switch ext {
	case "png", "PNG":
		return png.Decode(r)
	case "jpeg", "jpg", "JPEG", "JPG":
		return jpeg.Decode(r)
	case "webp", "WEBP":
		return webp.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported file type(%s)", ext)
	}
}

// LoadOptions is options for image loading.
type LoadOptions struct {
	// Size of resized image. Empty this means no resized.
	// If either X or Y of size is zero, auto filled it by calculating
	// from the other with keep aspect ratio.
	// For example, Let source image size (W, H) = (1920, 1080) and resized size
	// (RW, RH) = (960, 0) then auto-filled resized size (RW', RH') = (960, 540)
	// since 1920:1080 = 960:540.
	ResizedSize image.Point
}

// It is almost same as Get(), except that it accepts options for loaded image property.
func (l *Loader) GetWithOptions(file string, opt LoadOptions) (image.Image, error) {
	key := createImageKey(file, opt)

	l.mu.Lock()
	defer l.mu.Unlock()

	// attempt cache
	v, ok := l.cache.Get(key)
	if ok {
		return v.(image.Image), nil
	}
	// new arrival key, construct new image.
	m, err := AutoLoadFile(file)
	if err != nil {
		return nil, err
	}
	// resize original image
	if opt.ResizedSize != (image.Point{}) {
		// auto-fill resized size if either X or Y is zero.
		if opt.ResizedSize.X == 0 || opt.ResizedSize.Y == 0 {
			opt.ResizedSize = fixedAspectRateSize(m.Bounds().Size(), opt.ResizedSize)
		}
		m = resizeImage(m, opt)
	}
	l.cache.Add(key, m)
	return m, nil
}

func createImageKey(file string, opt LoadOptions) string {
	key := file + "_" + fmt.Sprintf("%dx%d", opt.ResizedSize.X, opt.ResizedSize.Y)
	return key
}

func fixedAspectRateSize(srcSize image.Point, resizedSize image.Point) image.Point {
	if srcSize.X == 0 || srcSize.Y == 0 {
		return resizedSize // can not calculate aspect ratio
	}

	aspectRateXY := float64(srcSize.X) / float64(srcSize.Y)
	if resizedSize.X == 0 {
		return image.Point{
			X: int(float64(resizedSize.Y) * aspectRateXY),
			Y: resizedSize.Y,
		}
	} else if resizedSize.Y == 0 {
		return image.Point{
			X: resizedSize.X,
			Y: int(float64(resizedSize.X) / aspectRateXY),
		}
	} else {
		return resizedSize // Have user requested value. Not needed to fixed aspect
	}
}

func resizeImage(src image.Image, opt LoadOptions) image.Image {
	dst := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: opt.ResizedSize})
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
