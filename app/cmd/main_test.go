package main

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/mzki/erago/app"
)

func clearAllFlag() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func setFlag(name, value string) error {
	return flag.CommandLine.Set(name, value)
}

func TestOverwriteFlag(t *testing.T) {
	testcases := []struct {
		FlagName  string
		FlagValue string
		Equals    func(*app.Config, interface{}) bool
	}{
		{flagNameLogFile, "<nolog>", func(conf *app.Config, v interface{}) bool {
			return conf.LogFile == v.(string)
		}},
		{flagNameLogLevel, "<nolevel>", func(conf *app.Config, v interface{}) bool {
			return conf.LogLevel == v.(string)
		}},
		{flagNameFont, "<nofont>", func(conf *app.Config, v interface{}) bool {
			return conf.Font == v.(string)
		}},
		{flagNameFontSize, "-1.0", func(conf *app.Config, v interface{}) bool {
			f, err := strconv.ParseFloat(v.(string), 64)
			if err != nil {
				return false
			}
			return conf.FontSize == f
		}},
	}
	for _, test := range testcases {
		clearAllFlag()

		parseFlags()
		// set flag should be after flag.Parse()
		if err := setFlag(test.FlagName, test.FlagValue); err != nil {
			t.Fatal(err)
		}

		conf := loadConfigOrDefault()
		overwriteConfigByFlag(conf)

		if ok := test.Equals(conf, test.FlagValue); !ok {
			t.Errorf("flag %v: not overwritten by flag value, expect: %v, got: %v", test.FlagName, test.FlagValue, conf)
		}
	}
}
