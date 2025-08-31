package app

import (
	"fmt"
	"os"
	"time"

	"github.com/mzki/erago"
	"github.com/mzki/erago/app/config"
	"github.com/mzki/erago/util/log"
)

// Testing test given script files on appConf context.
// the errors in testing are logged to appConf.LogFile. It returns testing succeded or not.
func Testing(appConf *config.Config, scriptFiles []string) bool {
	if appConf == nil {
		appConf = config.NewConfig(config.DefaultBaseDir)
	}

	// returned value must be called once.
	reset, err := config.SetupLogConfig(appConf)
	if err != nil {
		// TODO: what is better way to handle fatal error in this case?
		fmt.Fprintf(os.Stderr, "log configuration failed: %v\n", err)
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
