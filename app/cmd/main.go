package main

import (
	"flag"
	"fmt"
	"os"
	// "net/http"
	// _ "net/http/pprof"
	// "runtime"

	"github.com/mzki/erago/app"
)

var (
	version string = "dev"
	commit  string = "none"
)

const Title = "erago"

func main() {
	// runtime.SetBlockProfileRate(1)
	// go func() {
	// 	log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	// }()

	mode, args := parseFlags()

	appConf := loadConfigOrDefault()
	switch mode {
	case runMain:
		fullTitle := Title + " " + version + "-" + commit
		app.Main(fullTitle, appConf)
	case runTest:
		app.Testing(appConf, args)
	}
}

type runningMode int

const (
	runMain runningMode = iota
	runTest
)

var (
	LogFile  string  = app.DefaultLogFile
	LogLevel string  = app.LogLevelInfo
	Font     string  = app.DefaultFont
	FontSize float64 = app.DefaultFontSize
)

func parseFlags() (runningMode, []string) {
	flag.Usage = printHelp

	flag.StringVar(&LogFile, "logfile", LogFile, "`output-file` to write log. { stdout | stderr } is OK.")
	flag.StringVar(&LogLevel, "loglevel", LogLevel, "`level` = { info | debug }.\n\t"+
		"info outputs information level log only, and debug also outputs debug level log.")
	flag.StringVar(&Font, "font", Font, "`font-path` to print text on the screen. use builtin default if empty")
	flag.Float64Var(&FontSize, "fontsize", FontSize, "`font-size` to print text on the screen, in point(Pt.).")

	testing := false
	flag.BoolVar(&testing, "test", testing, "run tests and quit. after given this flag,"+
		" script files to test are required in the command-line arguments")

	showVersion := false
	flag.BoolVar(&showVersion, "version", showVersion, "show version info and quit.")

	flag.Parse()

	// show version and exit immediately
	if showVersion {
		fmt.Println(version)
		os.Exit(0) // normal termination
	}

	// return running mode
	if testing {
		return runTest, flag.Args()
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
	flag.PrintDefaults()
}

func loadConfigOrDefault() *app.Config {
	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
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
