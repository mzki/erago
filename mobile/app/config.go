package mobile

import (
	"os"
	"path/filepath"

	"local/erago/app"
	"local/erago/util/log"
)

const (
	minFontSize = 15 // in dip.
)

var configInstance = &Config{
	fontSize: minFontSize,
}

type Config struct {
	fontSize float64 // in dip
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) SetFontSize(f float64) *Config {
	c.fontSize = minFontSize
	return c
}

// set up app.Config for mobile with root directory
// on mobile and mobile specifc config.
func mobileConfig(mobileDir string, conf *Config) (*app.Config, error) {
	// set log configuration. in mobile, log level is limitted to info and
	// destination is stderr.
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stderr)

	confPath := filepath.Join(mobileDir, app.ConfigFile)
	appConf, err := app.LoadConfigOrDefault(confPath)
	if err != nil {
		log.Infof("Error: LoadConfigFile(%s) FAIL: %v", confPath, err)
		return nil, err
	}

	// overwrite by mobile config.
	appConf.FontSize = conf.fontSize
	if fSize := appConf.FontSize; fSize < minFontSize {
		log.Infof("font size is too small (%.1f dip), use %.1f dip insteadly.", fSize, minFontSize)
		appConf.FontSize = minFontSize
	}

	return appConf, nil
}
