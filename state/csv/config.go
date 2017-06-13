package csv

import (
	"path/filepath"
)

type Config struct {
	Dir          string // extracting CSV directory
	CharaPattern string
}

func (c Config) filepath(file string) string {
	return filepath.Join(c.Dir, file)
}

func (c Config) charaPattern() string {
	return c.filepath(c.CharaPattern)
}
