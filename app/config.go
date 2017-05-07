package app

import (
	"local/erago"
	"local/erago/serialize/toml"
	"local/erago/util"
	"local/erago/util/log"
	"local/erago/view/exp/theme"
)

const (
	// default configuration file.
	ConfigFile = "erago.conf"

	DefaultBaseDir = erago.DefaultBaseDir

	LogFileStdOut  = "stdout"    // specify log outputs to stdout
	LogFileStdErr  = "stderr"    // specify log outputs to stderr
	DefaultLogFile = "erago.log" // default output log file.

	LogLevelInfo  = "info"  // logging only information level.
	LogLevelDebug = "debug" // logging all levels, debug and info.

	DefaultFont     = theme.DefaultFontName // font file. empty means use builtin font.
	DefaultFontSize = 12.0                  // font size in pt

	DefaultWidth  = 800 // initial window width
	DefaultHeight = 600 // initial window height
)

// Configure for the Applicaltion.
// To build this, use NewConfig instead of struct constructor, AppConfig{}.
type Config struct {
	LogFile  string `toml:"logfile"`
	LogLevel string `toml:"loglevel"`

	Font     string  `toml:"font"`     // path for fontfile. empty means that use builtin font.
	FontSize float64 `toml:"fontsize"` // font size in pt.

	Width  int `toml:"width"`  // initial window width.
	Height int `toml:"height"` // initial window height.
	// TODO: Title string  // title on window top.

	Game erago.Config `toml:"game"`
}

// return default App config. if baseDir is empty
// use default insteadly.
func NewConfig(baseDir string) *Config {
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}
	return &Config{
		LogFile:  DefaultLogFile,
		LogLevel: LogLevelInfo,
		Font:     DefaultFont,
		FontSize: DefaultFontSize,
		Width:    DefaultWidth,
		Height:   DefaultHeight,

		Game: erago.NewConfig(baseDir),
	}
}

// update unexported fields. It should be called when
// any exported fields are updated.
func (c *Config) Update() {
	c.Game.SetBaseDir(c.Game.BaseDir)
}

// if config file exists load it and return.
// if not exists return default config and write it.
func LoadConfigOrDefault(file string) (*Config, error) {
	appConf := NewConfig(DefaultBaseDir)
	if !util.FileExists(file) {
		log.Infof("Config file (%v) does not exist. Use default config and write it to file.", file)
		return appConf, toml.EncodeFile(file, appConf) // write default config
	}

	if err := toml.DecodeFile(file, appConf); err != nil {
		return nil, err
	}
	log.Infof("Config file (%v) is loaded successfully , use it.", file)
	appConf.Update()
	return appConf, nil
}
