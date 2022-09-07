package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/groupcache/lru"
	"golang.org/x/image/draw"
)

// Pool holds image caches.
// concurrent use is OK.
type Pool struct {
	mu    *sync.Mutex
	cache *lru.Cache // under mutex because cache is not safe for concurrently.
}

const DefaultPoolSize = 10

// Pool has cachedSize cache entries. the oldest image
// is removed from cache using LRU.
// use DefaultPoolSize if cachedSize <= 0.
func NewPool(cachedSize int) *Pool {
	if cachedSize <= 0 {
		cachedSize = DefaultPoolSize
	}
	return &Pool{
		mu:    new(sync.Mutex),
		cache: lru.New(cachedSize),
	}
}

// get cached image by using image's file name.
// if not found, load image data from file
// and return loaded image with loading error.
// error nil means loaded image found.
func (p *Pool) Get(file string) (image.Image, error) {
	p.mu.Lock()
	v, ok := p.cache.Get(file)
	p.mu.Unlock()
	if ok {
		return v.(image.Image), nil
	}

	// new arrival key, construct new image.
	m, err := AutoLoadFile(file)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.cache.Add(file, m)
	p.mu.Unlock()
	return m, nil
}

// auto detect file extension, png, jpeg, and jpg, and
// return loaded image data with error.
// error contains file not found, unsupported extension, ... etc.
func AutoLoadFile(file string) (image.Image, error) {
	ext := filepath.Ext(file)
	if len(ext) == 0 {
		return nil, fmt.Errorf("file must have the extension like .png, .jpeg, or .jpg.")
	}
	ext = ext[1:] // remove first characater "."

	fp, err := os.Open(file)
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
	case "png":
		return png.Decode(r)
	case "jpeg", "jpg":
		return jpeg.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported file type(%s)", ext)
	}
}

// LoadOptions is options for image loading.
type LoadOptions struct {
	ResizedSize image.Point // size of resized image. Empty this means no resized.
}

// It is almost same as Get(), except that it accepts options for loaded image property.
func (p *Pool) GetWithOptions(file string, opt LoadOptions) (image.Image, error) {
	key := createImageKey(file, opt)

	p.mu.Lock()
	defer p.mu.Unlock()

	// attempt cache
	v, ok := p.cache.Get(key)
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
		m = resizeImage(m, opt)
	}
	p.cache.Add(key, m)
	return m, nil
}

func createImageKey(file string, opt LoadOptions) string {
	key := file + "_" + fmt.Sprintf("%dx%d", opt.ResizedSize.X, opt.ResizedSize.Y)
	return key
}

func resizeImage(src image.Image, opt LoadOptions) image.Image {
	dst := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: opt.ResizedSize})
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
