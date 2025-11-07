package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mzki/erago"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/serialize/toml"
	"github.com/mzki/erago/util/log"
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

	DefaultHistoryLineCount = 1024

	DefaultImageCacheSize = 32 // filelimit 3MB * 32 = 96MB
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

	// number of image cache entries for printImage. it will reduce image loading time for same image and options,
	// but will increase memory usage in user device. 0 or negative value can be set and treated as default value.
	ImageCacheSize int `toml:"image_cache_size"`

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
		HistoryLineCount: DefaultHistoryLineCount,
		//HistoryBytesPerLine: int(DefaultAppTextViewOptions.MaxParagraphBytes),
		ImageCacheSize:       DefaultImageCacheSize,
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

	appConf := NewConfig(DefaultBaseDir) // default value will be remain when missing at decoded config.
	if err := toml.DecodeFile(file, appConf); err != nil {
		return nil, err
	}
	return appConf, nil
}

// set up log configuration and return finalize function with internal error.
// when returned error, the finalize function is nil and need not be called.
func SetupLogConfig(appConf *Config) (func(), error) {
	// set log level.
	switch level := appConf.LogLevel; level {
	case LogLevelInfo:
		log.SetLevel(log.InfoLevel)
	case LogLevelDebug:
		log.SetLevel(log.DebugLevel)
	default:
		log.Infof("unknown log level(%s). use 'info' level insteadly.", level)
		log.SetLevel(log.InfoLevel)
	}

	// set log distination
	var (
		dstString string
		writer    io.WriteCloser
		closeFunc func()
	)
	switch logfile := appConf.LogFile; logfile {
	case LogFileStdOut, "":
		dstString = "Stdout"
		writer = os.Stdout
		closeFunc = func() {}
	case LogFileStdErr:
		dstString = "Stdout"
		writer = os.Stderr
		closeFunc = func() {}
	default:
		dstString = logfile
		fp, err := filesystem.Store(logfile)
		if err != nil {
			return nil, err
		}
		writer = fp
		closeFunc = func() { fp.Close() }
	}
	logLimit := appConf.LogLimitMegaByte * 1000 * 1000
	if logLimit < 0 {
		logLimit = 0
	}
	log.SetOutput(log.LimitWriter(writer, logLimit))
	if err := testingLogOutput("log output sanity check..."); err != nil {
		closeFunc()
		return nil, err
	}
	log.Infof("Output log to %s", dstString)

	return closeFunc, nil
}

func testingLogOutput(msg string) error {
	log.Debug(msg)
	err := log.Err()
	switch {
	case errors.Is(err, log.ErrOutputDiscardedByLevel):
	case errors.Is(err, io.EOF):
	case err == nil:
	default:
		return fmt.Errorf("log output error: %w", err)
	}
	return nil // normal operation
}
