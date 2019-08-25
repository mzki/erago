package main

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/mzki/erago/app"
)

var testFlagSet *flag.FlagSet

func init() {
	clearAllFlag()
}

func clearAllFlag() {
	testFlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func setFlag(name, value string) error {
	return testFlagSet.Set(name, value)
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

		parseFlags(testFlagSet, []string{})
		// set flag should be after flag.Parse()
		if err := setFlag(test.FlagName, test.FlagValue); err != nil {
			t.Fatal(err)
		}

		conf := loadConfigOrDefault()
		overwriteConfigByFlag(conf, testFlagSet)

		if ok := test.Equals(conf, test.FlagValue); !ok {
			t.Errorf("flag %v: not overwritten by flag value, expect: %v, got: %v", test.FlagName, test.FlagValue, conf)
		}
	}
}
