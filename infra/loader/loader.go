package loader

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Loader is a platform depended file loader which searches module name and
// return its content as io.Reader.
type Loader interface {
	// Load loads content specified by the relative path
	// from erago system root directory.
	// It returns io.Reader for the loaded content with no error,
	// or returns nil with file loading error.
	Load(filepath string) (reader io.ReadCloser, err error)
}

const (
	DefaultMaxFileSize = 3 * 1024 * 1024 // 3MByte
)

var (
	// OS is a adaptation of the os.Open() with Loader interface.
	OS = &OSLoader{MaxFileSize: DefaultMaxFileSize}
	// String is a adaptation of strings.Buffer with Loader interface.
	String = LoaderFunc(StringReadCloser)
)

// LoaderFunc implements Loader interface.
type LoaderFunc func(string) (io.ReadCloser, error)

func (fn LoaderFunc) Load(filepath string) (reader io.ReadCloser, err error) {
	return fn(filepath)
}

// OSLoader is a adaptation of the os.Open() with Loader interface.
type OSLoader struct {
	MaxFileSize int64 // in bytes
}

func (osldr *OSLoader) Load(filepath string) (reader io.ReadCloser, err error) {
	finfo, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("can not fetch file info: %v", err)
	}

	if maxSize := osldr.MaxFileSize; maxSize > 0 && finfo.Size() > maxSize {
		return nil, fmt.Errorf("file(%s) is too large size(>%v) to load", filepath, maxSize)
	}

	return os.Open(filepath)
	// Close() is responsible for the caller.
}

// StringReadCloser is helper function which creates io.ReadCloser from a entire content
// to adapt Loader interface.
func StringReadCloser(content string) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(content)), nil
}
