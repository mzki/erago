package state

import (
	"local/erago/util"
)

const defaultSaveFileDir = "sav"

// Scene Config with csv Config
type Config struct {
	path        util.PathManager
	SaveFileDir string `toml:"savefile_dir"`
}

// return new default config
func NewConfig(basedir string) Config {
	return Config{
		path:        util.NewPathManager(basedir),
		SaveFileDir: defaultSaveFileDir,
	}
}

// return save file path
func (c Config) savePath(file string) string {
	return c.path.Join(c.SaveFileDir, file)
}

// set base direcotry which is prefixed SaveFileDir.
func (c *Config) SetBaseDir(baseDir string) {
	c.path = util.NewPathManager(baseDir)
}
