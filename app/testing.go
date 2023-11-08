package app

import (
	"time"

	"github.com/mzki/erago"
	"github.com/mzki/erago/util/log"
)

// Testing test given script files on appConf context.
// the errors in testing are logged to appConf.LogFile. It returns testing succeded or not.
func Testing(appConf *Config, scriptFiles []string) bool {
	if appConf == nil {
		appConf = NewConfig(DefaultBaseDir)
	}

	// returned value must be called once.
	// returned value must be called once.
	reset, err := SetLogConfig(appConf)
	if err != nil {
		log.Infoln("Error: Can't create log file:", err)
		return false
	}
	defer reset()

	if len(scriptFiles) == 0 {
		log.Info("app.Testing: nothing script files to test")
		return false
	}

	if err := erago.Testing(appConf.Game, scriptFiles, time.Duration(appConf.TestingTimeoutSecond)*time.Second); err != nil {
		log.Infoln("Error: app.Testing:", err)
		return false
	} else {
		log.Infoln("PASS: script files,", scriptFiles, ", is OK!")
		return true
	}
}
