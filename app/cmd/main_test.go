package main

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/mzki/erago/app/config"
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
		Equals    func(*config.Config, interface{}) bool
	}{
		{flagNameLogFile, "<nolog>", func(conf *config.Config, v interface{}) bool {
			return conf.LogFile == v.(string)
		}},
		{flagNameLogLevel, "<nolevel>", func(conf *config.Config, v interface{}) bool {
			return conf.LogLevel == v.(string)
		}},
		{flagNameLogLimit, "1024", func(conf *config.Config, v interface{}) bool {
			i, err := strconv.ParseInt(v.(string), 10, 64)
			if err != nil {
				return false
			}
			return conf.LogLimitMegaByte == i
		}},
		{flagNameFont, "<nofont>", func(conf *config.Config, v interface{}) bool {
			return conf.Font == v.(string)
		}},
		{flagNameFontSize, "-1.0", func(conf *config.Config, v interface{}) bool {
			f, err := strconv.ParseFloat(v.(string), 64)
			if err != nil {
				return false
			}
			return conf.FontSize == f
		}},
		{flagNameTestTimeout, "12", func(conf *config.Config, v interface{}) bool {
			i, err := strconv.ParseInt(v.(string), 10, 32)
			if err != nil {
				return false
			}
			return conf.TestingTimeoutSecond == int(i)
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

func TestMainApp(t *testing.T) {
	t.Skip("It is for debug purpose")
	main()
}
