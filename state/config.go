package state

import (
	"path/filepath"
)

// configuration for game state.
type Config struct {
	SaveFileDir string
}

// return save file path
func (c Config) savePath(file string) string {
	return filepath.Join(c.SaveFileDir, file)
}
