package csv

import (
	"local/erago/util"
)

type Config struct {
	path util.PathManager

	Dir              string `toml:"load_dir"` // extracting CSV directory
	LoadCharaPattern string `toml:"load_chara_pattern"`
}

const (
	defaultCSVDir           = "CSV"
	defaultLoadCharaPattern = "Chara/Chara*"
)

// default CSV config
func NewConfig(basedir string) Config {
	return Config{
		path:             util.NewPathManager(basedir),
		Dir:              defaultCSVDir,
		LoadCharaPattern: defaultLoadCharaPattern,
	}
}

// set baseDir which is prefixed to Dir and LoadCharaPattern.
func (c *Config) SetBaseDir(baseDir string) {
	c.path = util.NewPathManager(baseDir)
}

func (c Config) loadPathOf(file string) string {
	return c.path.Join(c.Dir, file)
}

func (c Config) charaPattern() string {
	return c.path.Join(c.Dir, c.LoadCharaPattern)
}
