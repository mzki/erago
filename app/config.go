package app

import (
	"errors"
	"time"

	"github.com/mzki/erago"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/serialize/toml"
)

const (
	// default configuration file.
	ConfigFile = "erago.conf"

	DefaultBaseDir = erago.DefaultBaseDir

	LogFileStdOut  = "stdout"    // specify log outputs to stdout
	LogFileStdErr  = "stderr"    // specify log outputs to stderr
	DefaultLogFile = "erago.log" // default output log file.

	LogLevelInfo            = "info"  // logging only information level.
	LogLevelDebug           = "debug" // logging all levels, debug and info.
	DefaultLogLevel         = LogLevelInfo
	DefaultLogLimitMegaByte = 10 // 10 * 1000 * 1000 Bytes

	DefaultFont     = ""   // font file. empty means use builtin font.
	DefaultFontSize = 12.0 // font size in pt

	DefaultWidth  = 800 // initial window width
	DefaultHeight = 600 // initial window height

	DefaultTestingTimeoutSecond = int(erago.DefaultTestingTimeout / time.Second)
)

// Configure for the Applicaltion.
// To build this, use NewConfig instead of struct constructor, AppConfig{}.
type Config struct {
	LogFile          string `toml:"logfile"`
	LogLevel         string `toml:"loglevel"`
	LogLimitMegaByte int64  `toml:"loglimit_megabytes"`

	Font     string  `toml:"font"`     // path for fontfile. empty means that use builtin font.
	FontSize float64 `toml:"fontsize"` // font size in pt.

	Width  int `toml:"width"`  // initial window width.
	Height int `toml:"height"` // initial window height.
	// TODO: Title string  // title on window top.

	// number of lines or bytes per line stored in history.
	// these can take 0 or negative in which case use default value instead.
	HistoryLineCount int `toml:"history_line_count"`
	// TODO: This is not implemented yet, but whole byte limits are builtin implemented
	//HistoryBytesPerLine int `toml:"history_bytes_per_line"`

	// timeout value for testing mode only, in second.
	TestingTimeoutSecond int `toml:"testing_timeout_sec"`

	Game erago.Config `toml:"game"`
}

// return default App config. if baseDir is empty
// use default insteadly.
func NewConfig(baseDir string) *Config {
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}
	return &Config{
		LogFile:          DefaultLogFile,
		LogLevel:         DefaultLogLevel,
		LogLimitMegaByte: DefaultLogLimitMegaByte,
		Font:             DefaultFont,
		FontSize:         DefaultFontSize,
		Width:            DefaultWidth,
		Height:           DefaultHeight,
		HistoryLineCount: int(DefaultAppTextViewOptions.MaxParagraphs),
		//HistoryBytesPerLine: int(DefaultAppTextViewOptions.MaxParagraphBytes),
		TestingTimeoutSecond: DefaultTestingTimeoutSecond,

		Game: erago.NewConfig(baseDir),
	}
}

// ErrDefaultConfigGenerated implies that the specified config file is not found,
// and intead of that default config is generated and used.
var ErrDefaultConfigGenerated error = errors.New("default config generated")

// if config file exists load it and return.
// if not exists return default config and write it.
func LoadConfigOrDefault(file string) (*Config, error) {
	if !filesystem.Exist(file) {
		appConf := NewConfig(DefaultBaseDir)
		// write default config
		if err := toml.EncodeFile(file, appConf); err != nil {
			return nil, err
		}
		return appConf, ErrDefaultConfigGenerated
	}

	appConf := &Config{}
	if err := toml.DecodeFile(file, appConf); err != nil {
		return nil, err
	}
	return appConf, nil
}
