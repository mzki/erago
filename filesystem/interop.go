package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

// InteropFileSystem interops between fs.FS interface and FileSystem interface.
// It implements fs.FS and FileSystem interface.
type InteropFileSystem struct {
	Backend fs.FS
}

// FromFS converts fs.FS interface into FileSystem interface.
func FromFS(fsys fs.FS) FileSystem {
	return &InteropFileSystem{
		Backend: fsys,
	}
}

func fsPath(path string) string {
	return filepath.Clean(filepath.ToSlash(path))
}

func (ifs *InteropFileSystem) mustBackend() fs.FS {
	if ifs.Backend == nil {
		ifs.Backend = FileSystem(&OSFileSystem{MaxFileSize: DefaultMaxFileSize}).(fs.FS)
	}
	return ifs.Backend
}

// implements FileSystem interface.
func (ifs *InteropFileSystem) Load(path string) (io.ReadCloser, error) {
	path = fsPath(path)
	return ifs.mustBackend().Open(path)
}

// implements FileSystem interface.
func (ifs *InteropFileSystem) Store(path string) (io.WriteCloser, error) {
	path = fsPath(path)
	backend := ifs.mustBackend()
	if fsystem, ok := backend.(FileSystem); ok {
		return fsystem.Store(path)
	} else {
		return nil, fmt.Errorf("file create operation is not supported")
	}
}

// implements FileSystem interface.
func (ifs *InteropFileSystem) Exist(path string) bool {
	path = fsPath(path)
	file, err := ifs.mustBackend().Open(path)
	if err != nil {
		return false
	} else {
		file.Close()
		return true
	}
}

// Implement fs.FS interface
func (ifs *InteropFileSystem) Open(path string) (fs.File, error) {
	path = fsPath(path)
	return ifs.mustBackend().Open(path)
}
