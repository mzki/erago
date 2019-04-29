package main

import (
	"flag"
	"fmt"
	// "net/http"
	// _ "net/http/pprof"
	"os"
	// "runtime"

	"local/erago/app"
	"local/erago/util/log"
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

	fullTitle := Title + " " + version + "-" + commit
	log.Info("-- " + fullTitle + " --")

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		log.Info(err)
		return
	}

	mode, args := parseFlags(appConf)
	switch mode {
	case runMain:
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

func parseFlags(c *app.Config) (runningMode, []string) {
	flag.Usage = printHelp

	flag.StringVar(&c.LogFile, "logfile", c.LogFile, "`output-file` to write log. { stdout | stderr } is OK.")
	flag.StringVar(&c.LogLevel, "loglevel", c.LogLevel, "`level` = { info | debug }.\n\t"+
		"info outputs information level log only, and debug also outputs debug level log.")
	flag.StringVar(&c.Font, "font", c.Font, "`font-path` to print text on the screen. use builtin default if empty")
	flag.Float64Var(&c.FontSize, "fontsize", c.FontSize, "`font-size` to print text on the screen, in point(Pt.).")

	testing := false
	flag.BoolVar(&testing, "test", testing, "run tests and quit. after given this flag,"+
		" script files to test are required in the command-line arguments")

	flag.Parse()

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
