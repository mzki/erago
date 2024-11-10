package model

import (
	"io"

	"github.com/mzki/erago/filesystem"
)

// FileSystem is same as github.com/mzki/erago/filesystem.FileSystem interface.
// It need to define same interface explicitly to expose mobile interface.
type FileSystem interface {
	Load(filepath string) (ReadCloser, error)
	Store(filepath string) (WriteCloser, error)
	Exist(filepath string) bool
}

type FileSystemGlob interface {
	FileSystem
	Glob(pattern string) (*StringList, error) // mobile not support slice, use StringList instead of that
}

type ReadCloser interface {
	Read([]byte) (int, error)
	Close() error
}

type WriteCloser interface {
	Write([]byte) (int, error)
	Close() error
}

type StringList struct {
	data []string
}

func NewStringList() *StringList           { return &StringList{} }
func (sl *StringList) Append(s string)     { sl.data = append(sl.data, s) }
func (sl *StringList) Len() int            { return len(sl.data) }
func (sl *StringList) Get(i int) string    { return sl.data[i] }
func (sl *StringList) Set(i int, s string) { sl.data[i] = s }

// io.ReadCloser and ReadCloser is compatible, but Load() (io.ReadCloser, ...) and
// Load() (ReadCloser, ...) is treated as diffrent thing. we need to interop between them.

type fsysInterop struct {
	FileSystem
}

// FromMobileFS converts mobile/model/v2.FileSystem to filesystem.FileSystem.
func FromMobileFS(fsys FileSystem) filesystem.FileSystem {
	return &fsysInterop{fsys}
}

func (fsys *fsysInterop) Load(s string) (io.ReadCloser, error) {
	return fsys.FileSystem.Load(s)
}

func (fsys *fsysInterop) Store(s string) (io.WriteCloser, error) {
	return fsys.FileSystem.Store(s)
}

type mobileFSInterop struct {
	filesystem.FileSystem
}

// FromGoFS converts filesystem.FileSystem to mobile/model/v2.FileSystem to .
func FromGoFS(fsys filesystem.FileSystem) FileSystem {
	return &mobileFSInterop{fsys}
}

func (fsys *mobileFSInterop) Load(s string) (ReadCloser, error) {
	return fsys.FileSystem.Load(s)
}

func (fsys *mobileFSInterop) Store(s string) (WriteCloser, error) {
	return fsys.FileSystem.Store(s)
}

type fsysGlobInterop struct {
	fsysInterop
	gfsys FileSystemGlob
}

func FromMobileFSGlob(fsys FileSystemGlob) filesystem.FileSystemGlob {
	return &fsysGlobInterop{fsysInterop{fsys}, fsys}
}

func (fsys *fsysGlobInterop) Glob(pattern string) ([]string, error) {
	slist, err := fsys.gfsys.Glob(pattern)
	return slist.data, err
}

type mobileFSGlobInterop struct {
	mobileFSInterop
	gfsys filesystem.FileSystemGlob
}

func FromGoFSGlob(fsys filesystem.FileSystemGlob) FileSystemGlob {
	return &mobileFSGlobInterop{mobileFSInterop{fsys}, fsys}
}

func (fsys *mobileFSGlobInterop) Glob(pattern string) (*StringList, error) {
	ss, err := fsys.gfsys.Glob(pattern)
	return &StringList{data: ss}, err
}
