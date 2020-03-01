package filesystem

import "io"

// abstraction for the filesystem.
type FileSystem interface {
	Loader

	// create data store entry.
	Store(filepath string) (io.WriteCloser, error)
}

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
	return Default.Load(filepath)
}

func Exist(filepath string) bool {
	return Default.Exist(filepath)
}

func Store(filepath string) (io.WriteCloser, error) {
	return Default.Store(filepath)
}
