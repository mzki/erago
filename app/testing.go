package app

import (
	"github.com/mzki/erago"
	"github.com/mzki/erago/util/log"
)

// Testing test given script files on appConf context.
// the errors in testing are logged to appConf.LogFile.
func Testing(appConf *Config, scriptFiles []string) {
	if appConf == nil {
		appConf = NewConfig(DefaultBaseDir)
	}

	// returned value must be called once.
	// returned value must be called once.
	reset, err := SetLogConfig(appConf)
	if err != nil {
		log.Infoln("Error: Can't create log file:", err)
		return
	}
	defer reset()

	if scriptFiles == nil || len(scriptFiles) == 0 {
		log.Info("app.Testing: nothing script files to test")
		return
	}

	if err := erago.Testing(appConf.Game, scriptFiles); err != nil {
		log.Infoln("Error: app.Testing:", err)
	} else {
		log.Infoln("PASS: script files,", scriptFiles, ", is OK!")
	}
}
