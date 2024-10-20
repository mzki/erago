package filesystem

import (
	"io"

	"github.com/mzki/erago/util/log"
)

// abstraction for the filesystem.
type FileSystem interface {
	Loader

	// create data store entry.
	Store(filepath string) (io.WriteCloser, error)
}

// FileSystemPR is a composit interface with FileSystem and PathResolver interfaces.
type FileSystemPR interface {
	FileSystem
	PathResolver
}

// RFileSystemPR is a Readonly FileSystemPR interface.
type RFileSystemPR interface {
	Loader
	PathResolver
}

// path resolver resolves file path on the filesystem.
type PathResolver interface {
	ResolvePath(path string) (string, error)
}

// FileSytemGlob has ability for Glob files.
type FileSystemGlob interface {
	FileSystem
	Glob(pattern string) ([]string, error)
}

// FileSystemGlobPR is FileSystemGlob with implement PathResolver interface.
type FileSystemGlobPR interface {
	FileSystemGlob
	PathResolver
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
	Default FileSystemGlobPR = Desktop
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

// Glob is wrap function for filepath.Glob with use filesystem.FileSystemPR
func GlobFS(fs FileSystemGlob, pattern string) ([]string, error) {
	return fs.Glob(pattern)
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
func ResolvePathFS(fs FileSystemPR, path string) (string, error) {
	return fs.ResolvePath(path)
}

// OpenWatcher creates Watcher interface from Default FileSystem.
// Note that returned watcher must call Close() after use.
func OpenWatcher() (Watcher, error) {
	return OpenWatcherPR(Default)
}

// OpenWatcherPR creates Watcher interface from given PathResolver.
func OpenWatcherPR(pr PathResolver) (Watcher, error) {
	return newWatcher(pr)
}
