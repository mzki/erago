package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

var (
	// Mobile is a FileSystem for the mobile environment that
	// requires absolute path to access file since the notion of
	// current directory is different from desktop.
	Mobile FileSystemPR = &AbsPathFileSystem{
		CurrentDir: "",
		Backend:    Desktop,
	}
)

// AbsPathFileSystem completes absolute path for every file access.
// The absolute path is made by using filepath.Abs when CurrentDir is set to empty,
// or made by file.Join(CurrentDir, relativePath) when CurrentDir is set.
// The Backend is used to access File API. and The OSFileSystem is used as Backend when
// it is nil.
//
// AbsPathFileSystem implements FileSystem, and fs.FS interface.
type AbsPathFileSystem struct {
	CurrentDir string
	Backend    FileSystem
}

// AbsDirFileSystem create new FileSystem in which all files are under the dirPath.
// absDirPath must be aboslute path. Otherwise method call like Load will fail.
// Backend FileSystem is used for Default. User can modify Backend for specific purpose.
func AbsDirFileSystem(absDirPath string) *AbsPathFileSystem {
	return &AbsPathFileSystem{
		CurrentDir: absDirPath,
		Backend:    Default,
	}
}

// ResolvePath complete parent directory path to fpath when fpath is a relative path.
// It returns fpath itself when fpath is already absolute path.
func (absfs *AbsPathFileSystem) ResolvePath(fpath string) (string, error) {
	if filepath.IsAbs(fpath) {
		return fpath, nil
	}

	// fpath seems to be relative file path, complete parent directory path.
	if absfs.CurrentDir == "" {
		return filepath.Abs(fpath)
	} else if filepath.IsAbs(absfs.CurrentDir) {
		return filepath.Clean(filepath.Join(absfs.CurrentDir, fpath)), nil
	} else {
		return "", fmt.Errorf("AbsPathFileSystem: CurrentDir is not absolute path: %s", absfs.CurrentDir)
	}
}

func (absfs *AbsPathFileSystem) mustBackend() FileSystem {
	if absfs.Backend == nil {
		absfs.Backend = &OSFileSystem{MaxFileSize: DefaultMaxFileSize}
	}
	return absfs.Backend
}

func (absfs *AbsPathFileSystem) Load(fpath string) (reader io.ReadCloser, err error) {
	fpath, err = absfs.ResolvePath(fpath)
	if err != nil {
		return nil, fmt.Errorf("AbsPathFileSystem.Load() error: %v", err) // TODO use %w on go1.13 above
	}
	backend := absfs.mustBackend()
	return backend.Load(fpath)
}

func (absfs *AbsPathFileSystem) Exist(fpath string) bool {
	fpath, err := absfs.ResolvePath(fpath)
	if err != nil {
		return false
	}
	backend := absfs.mustBackend()
	return backend.Exist(fpath)
}

func (absfs *AbsPathFileSystem) Store(fpath string) (writer io.WriteCloser, err error) {
	fpath, err = absfs.ResolvePath(fpath)
	if err != nil {
		return nil, fmt.Errorf("AbsPathFileSystem.Store() error: %v", err) // TODO use %w on go1.13 above
	}
	backend := absfs.mustBackend()
	return backend.Store(fpath)
}

// implements fs.FS interface.
func (absfs *AbsPathFileSystem) Open(fpath string) (file fs.File, err error) {
	fpath, err = absfs.ResolvePath(fpath)
	if err != nil {
		return nil, err
	}

	backend := absfs.mustBackend()
	if fsys, ok := backend.(fs.FS); ok {
		return fsys.Open(fpath)
	}

	r, err := backend.Load(fpath)
	if err != nil {
		return nil, err
	}
	if file, ok := r.(fs.File); ok {
		return file, nil
	} else {
		r.Close()
		return nil, &fs.PathError{Op: "open", Path: fpath, Err: fmt.Errorf("not supported")}
	}
}

// Implements FileSystemGlob interface.
func (absfs *AbsPathFileSystem) Glob(pattern string) ([]string, error) {
	pattern, err := absfs.ResolvePath(pattern)
	if err != nil {
		return nil, fmt.Errorf("AbsPathFileSystem.Glob() error: %w", err)
	}
	backend := absfs.mustBackend()
	if globFS, ok := backend.(FileSystemGlob); ok {
		return GlobFS(globFS, pattern)
	} else {
		return nil, fmt.Errorf("Glob is not supported at backend filesystem")
	}
}
