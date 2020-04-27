package model

import (
	"fmt"

	"github.com/mzki/erago"
	"github.com/mzki/erago/app"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
)

// because reference cycle is not allowed and
// UI is provieded by mobile devices,
// a instance of the model object is not exposed.
//
// Insteadly mobile devices just calls static function with
// UI reference.

var (
	game         *erago.Game
	mobileUI     *uiAdapter
	logCloseFunc func()
	initialized  = false
)

func Init(ui UI, baseDir string) error {
	if initialized {
		panic("game already initialized")
	}

	// setup mobile filesystem to properly access resources.
	mobileFS := filesystem.Mobile
	mobileFS.CurrentDir = baseDir
	filesystem.Default = mobileFS // replace file system used by erago

	// load config file
	configPath, err := mobileFS.ResolvePath(app.ConfigFile)
	if err != nil {
		return fmt.Errorf("Can not use base directory %v. err: %v", baseDir, err)
	}
	appConfig, err := app.LoadConfigOrDefault(configPath)
	switch err {
	case nil, app.ErrDefaultConfigGenerated:
	default:
		return fmt.Errorf("Config load error: %v", err)
	}

	if appConfig != nil {
		// set log level, destinations
		closeFunc, err := app.SetLogConfig(appConfig)
		if err != nil {
			return fmt.Errorf("Log configure error: %v", err)
		}
		// just store it, called on Quit()
		logCloseFunc = closeFunc
	}

	// create game instance
	game = erago.NewGame()
	mobileUI = newUIAdapter(ui)
	if err := game.Init(uiadapter.SingleUI{mobileUI}, appConfig.Game); err != nil {
		return err
	}
	game.RegisterAllRequestObserver(mobileUI)
	initialized = true
	return nil
}

// AppContext manages application context.
type AppContext interface {
	// it is called when mobile.app is quited.
	// the native framework must be quited by this signal.
	// the argument erorr indicates why app is quited.
	// nil error means quit correctly.
	NotifyQuit(error)
}

// run game thread.
// the game thread runs on background so it returns imediately.
// quiting the game thread is notifyed through AppContext.NotifyQuit().
func Main(appContext AppContext) {
	if !initialized {
		panic("Main(): Init() must be called firstly")
	}
	if game == nil {
		panic("Main(): nil game state")
	}
	go func() {
		// start game engine
		err := game.Main()
		appContext.NotifyQuit(err)
	}()
}

func Quit() {
	initialized = false

	if mobileUI != nil {
		game.UnregisterAllRequestObserver()
		mobileUI = nil
	}
	if game != nil {
		game.Quit()
		game = nil
	}
	if logCloseFunc != nil {
		logCloseFunc()
		logCloseFunc = nil
	}
}

func SendCommand(cmd string) {
	if !initialized {
		panic("SendCommand(): Init() must be called firstly")
	}
	game.Send(input.NewEventCommand(cmd))
}

func SendSkippingWait() {
	if !initialized {
		panic("SendSkippingWait(): Init() must be called firstly")
	}
	game.Send(input.NewEventControl(input.ControlStartSkippingWait))
}

func SendStopSkippingWait() {
	if !initialized {
		panic("SendStopSkippingWait(): Init() must be called firstly")
	}
	game.Send(input.NewEventControl(input.ControlStopSkippingWait))
}
