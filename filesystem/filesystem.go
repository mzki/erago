package filesystem

import (
	"io"
	"path/filepath"

	"github.com/mzki/erago/util/log"
)

// abstraction for the filesystem.
type FileSystem interface {
	Loader

	// create data store entry.
	Store(filepath string) (io.WriteCloser, error)
}

// path resolver resolves file path on the filesystem.
type PathResolver interface {
	ResolvePath(path string) (string, error)
}

// NopPathResolver implements PathResolver interface.
type NopPathResolver struct{}

// ResolvePath returns path as is and no error.
func (NopPathResolver) ResolvePath(path string) (string, error) { return path, nil }

// Loader is a platform depended file loader which searches file path and
// return its content as io.Reader.
type Loader interface {
	// Load loads content specified by the relative path
	// from erago system root directory.
	// It returns io.Reader for the loaded content with no error,
	// or returns nil with file loading error.
	Load(filepath string) (reader io.ReadCloser, err error)

	// Exist checks whether given filepath exist from erago system root directory.
	// It returns true when the filepath exists, otherwise return false.
	Exist(filepath string) bool
}

var (
	// Default is a default FileSystem to be used by exported functions.
	Default FileSystem = Desktop
)

func Load(filepath string) (reader io.ReadCloser, err error) {
	log.Debugf("FileSystem.Load: %s", filepath)
	return Default.Load(filepath)
}

func Exist(filepath string) bool {
	return Default.Exist(filepath)
}

func Store(filepath string) (io.WriteCloser, error) {
	log.Debugf("FileSystem.Store: %s", filepath)
	return Default.Store(filepath)
}

// Glob is wrap function for filepath.Glob with use filesystem.Default
func Glob(pattern string) ([]string, error) {
	log.Debugf("FileSystem.Glob: %s", pattern)
	return GlobFS(Default, pattern)
}

// Glob is wrap function for filepath.Glob with use filesystem.FileSystem
// if FileSystem also implements PathResolver, use it to resolve path.
func GlobFS(fs FileSystem, pattern string) ([]string, error) {
	var err error
	pattern, err = ResolvePathFS(fs, pattern)
	if err != nil {
		return nil, err
	}
	return filepath.Glob(pattern)
}

// ResolvePath resolve file path under filesystem.Default.
// if Default also implements PathResolver, use it to resolve path,
// otherwise returns path itself.
func ResolvePath(path string) (string, error) {
	return ResolvePathFS(Default, path)
}

// ResolvePathFS resolve file path under given FileSystem.
// if FileSystem also implements PathResolver, use it to resolve path,
// otherwise returns path itself.
func ResolvePathFS(fs FileSystem, path string) (string, error) {
	if pr, ok := fs.(PathResolver); ok {
		return pr.ResolvePath(path)
	}
	return path, nil
}

// OpenWatcher creates Watcher interface from Default FileSystem.
// If Default FileSystem not implements PathResolver interface, use NopPathResover for
// create Watcher. Note that returned watcher must call Close() after use.
func OpenWatcher() (Watcher, error) {
	if pr, ok := Default.(PathResolver); ok {
		return OpenWatcherPR(pr)
	} else {
		log.Debug("Default FileSystem not implement PathResolver. Use NopPathResolver instead of that.")
		return newWatcher(NopPathResolver{})
	}
}

// OpenWatcherPR creates Watcher interface from given PathResolver.
func OpenWatcherPR(pr PathResolver) (Watcher, error) {
	return newWatcher(pr)
}
