package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultMaxFileSize = 3 * 1024 * 1024 // 3MByte
)

var (
	// Desktop is a FileSystem for the desktop environment
	Desktop = &OSFileSystem{MaxFileSize: DefaultMaxFileSize}
	// String is a adaptation of strings.Buffer with Loader interface.
	String Loader = LoaderFunc(StringReadCloser)
)

// LoaderFunc implements Loader interface.
type LoaderFunc func(string) (io.ReadCloser, error)

func (fn LoaderFunc) Load(filepath string) (reader io.ReadCloser, err error) {
	return fn(filepath)
}

func (fn LoaderFunc) Exist(filepath string) bool {
	reader, err := fn(filepath)
	if err != nil {
		return false
	}
	reader.Close()
	return true
}

// OSFileSystem is a adaptation of the os.Open() with Loader interface.
type OSFileSystem struct {
	MaxFileSize int64 // in bytes
}

func (osfs *OSFileSystem) ResolvePath(fpath string) (string, error) {
	return filepath.Clean(fpath), nil
}

func (osfs *OSFileSystem) Load(filepath string) (reader io.ReadCloser, err error) {
	finfo, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("can not fetch file info: %v", err)
	}

	if maxSize := osfs.MaxFileSize; maxSize > 0 && finfo.Size() > maxSize {
		return nil, fmt.Errorf("file(%s) is too large size(>%v) to load", filepath, maxSize)
	}

	return os.Open(filepath)
	// Close() is responsible for the caller.
}

func (osfs *OSFileSystem) Exist(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func (osfs *OSFileSystem) Store(fpath string) (writer io.WriteCloser, err error) {
	// make directory of given path. if exist do nothing.
	err = os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return nil, fmt.Errorf("can not create store directory: %v", err)
	}
	fp, err := os.Create(fpath)
	if err != nil {
		return nil, fmt.Errorf("can not create store file: %v", err)
	}
	return fp, nil
}

// StringReadCloser is helper function which creates io.ReadCloser from a entire content
// to adapt Loader interface.
func StringReadCloser(content string) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(content)), nil
}
