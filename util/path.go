package util

import (
	"path/filepath"
)

// PathManager serves paths under its base directory.
type PathManager struct {
	baseDir string
}

// construct new path manager.
func NewPathManager(baseDir string) PathManager {
	return PathManager{baseDir}
}

// return its base directory.
func (p PathManager) Dir() string {
	return p.baseDir
}

// return path of baseDir/file.
func (p PathManager) Path(file string) string {
	return filepath.Join(p.baseDir, file)
}

// return path of baseDir/elm1/elm2/....
func (p PathManager) Join(elm ...string) string {
	return p.Path(filepath.Join(elm...))
}
