package main

import (
	"flag"
	"fmt"
	"os"

	// "net/http"
	// _ "net/http/pprof"
	// "runtime"

	"github.com/mzki/erago/app"
	"github.com/mzki/erago/infra/buildinfo"
)

var (
	buildInfo = buildinfo.Get()
)

const Title = "erago"

var flagSet = flag.CommandLine

func main() {
	// runtime.SetBlockProfileRate(1)
	// go func() {
	// 	log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	// }()

	mode, args := parseFlags(flagSet, os.Args[1:])

	appConf := loadConfigOrDefault()
	overwriteConfigByFlag(appConf, flagSet)

	switch mode {
	case runMain:
		fullTitle := Title + " " + buildInfo.Version + "-" + buildInfo.CommitHash
		app.Main(fullTitle, appConf)
	case runTest:
		ok := app.Testing(appConf, args)
		if ok {
			fmt.Fprintln(os.Stderr, "PASSED") // OK case also stderr to separate status messege and standard output in test execution.
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, "FAILED")
			os.Exit(1)
		}
	case runPackaging:
		ok := app.Packaging("./", appConf, appConfigPath, args)
		if ok {
			fmt.Fprintln(os.Stderr, "DONE")
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, "FAILED")
			os.Exit(1)
		}
	}
	os.Exit(0)
}

type runningMode int

const (
	runMain runningMode = iota
	runTest
	runPackaging
)

const appConfigPath = app.ConfigFile

var (
	LogFile              string  = app.DefaultLogFile
	LogLevel             string  = app.LogLevelInfo
	LogLimitMegaByte     int64   = app.DefaultLogLimitMegaByte
	Font                 string  = app.DefaultFont
	FontSize             float64 = app.DefaultFontSize
	TestingTimeoutSecond int     = app.DefaultTestingTimeoutSecond
)

const (
	flagNameLogFile  = "logfile"
	flagNameLogLevel = "loglevel"
	flagNameLogLimit = "loglimitmb"
	flagNameFont     = "font"
	flagNameFontSize = "fontsize"

	flagNameTest        = "test"
	flagNameTestTimeout = "test.timeoutsec"
	flagNamePackaging   = "packaging"
	flagNameVersion     = "version"
)

func parseFlags(flags *flag.FlagSet, argv []string) (runningMode, []string) {
	flags.Usage = printHelp

	flags.StringVar(&LogFile, flagNameLogFile, LogFile, "`output-file` to write log. { stdout | stderr } is OK.")
	flags.StringVar(&LogLevel, flagNameLogLevel, LogLevel, "`level` = { info | debug }.\n\t"+
		"info outputs information level log only, and debug also outputs debug level log.")
	flags.Int64Var(&LogLimitMegaByte, flagNameLogLimit, LogLimitMegaByte, "`log-limit-megabyte` to limits output data size for logging.")
	flags.StringVar(&Font, flagNameFont, Font, "`font-path` to print text on the screen. use builtin default if empty")
	flags.Float64Var(&FontSize, flagNameFontSize, FontSize, "`font-size` to print text on the screen, in point(Pt.).")

	flags.IntVar(&TestingTimeoutSecond, flagNameTestTimeout, TestingTimeoutSecond, "`test.timeout-sec` for timeout of test execution in second")

	testing := false
	flags.BoolVar(&testing, flagNameTest, testing, "run tests and quit. after given this flag,"+
		" script files to test are required in the command-line arguments")

	packaging := false
	flags.BoolVar(&packaging, flagNamePackaging, packaging, "run package creation and quit. after given this flag,"+
		" extra arguments are treated as the additional files to be included in the package.")

	showVersion := false
	flags.BoolVar(&showVersion, flagNameVersion, showVersion, "show version info and quit.")

	flags.Parse(argv)

	// show version and exit immediately
	if showVersion {
		fmt.Println(buildInfo.Version)
		os.Exit(0) // normal termination
	}

	// return running mode
	if testing {
		return runTest, flags.Args()
	}
	if packaging {
		return runPackaging, flag.Args()
	}
	return runMain, nil
}

func printHelp() {
	progName := os.Args[0]
	fmt.Fprintf(os.Stderr, `Usage: %s [options] [testing-scripts...]

  %s is a platform to create and play the adventure game 
  on a console-like screen.

  any flag values same as '%s' file overwrites the values 
  loaded from the file.

`, progName, progName, app.ConfigFile)
	flagSet.PrintDefaults()
}

func overwriteConfigByFlag(config *app.Config, flags *flag.FlagSet) {
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case flagNameLogFile:
			config.LogFile = LogFile
		case flagNameLogLevel:
			config.LogLevel = LogLevel
		case flagNameLogLimit:
			config.LogLimitMegaByte = LogLimitMegaByte
		case flagNameFont:
			config.Font = Font
		case flagNameFontSize:
			config.FontSize = FontSize
		case flagNameTestTimeout:
			config.TestingTimeoutSecond = TestingTimeoutSecond
		}
	})
}

func loadConfigOrDefault() *app.Config {
	appConf, err := app.LoadConfigOrDefault(appConfigPath)
	switch err {
	case app.ErrDefaultConfigGenerated:
		// TODO this message is shown at app.Main which starts logger?
		fmt.Fprintf(os.Stderr, "Config file (%v) does not exist. Use default config and write it to file.", app.ConfigFile)
		fallthrough
	case nil:
		// no errors. do nothing.
	default:
		// fatal error for loading config. quits
		panic(err)
	}

	return appConf
}
