package model_v2

import (
	"context"
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
	theGameContext context.Context
	theGame        *erago.Game
	mobileUI       *uiAdapter
	logCloseFunc   func()
	initialized    = false
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

	// finalize handler
	closeFuncs := make([]func(), 0, 4)
	logCloseFunc = func() {
		// call resource release funcions by reversed order
		// since release order should reversed from registeration order.
		for i := len(closeFuncs) - 1; i >= 0; i-- {
			closeFuncs[i]()
		}
	}
	defer func() {
		// initalization failed, release resources
		if !initialized {
			if logCloseFunc != nil {
				logCloseFunc()
				logCloseFunc = nil
			}
		}
	}()

	if appConfig != nil {
		// set log level, destinations
		closeFunc, err := app.SetLogConfig(appConfig)
		if err != nil {
			return fmt.Errorf("Log configure error: %v", err)
		}
		// just store it, called on Quit()
		closeFuncs = append(closeFuncs, closeFunc)
	}

	// create game instance
	ctx, cancel := context.WithCancel(context.Background())
	theGameContext = ctx
	closeFuncs = append(closeFuncs, cancel)

	theGame = erago.NewGame()
	mobileUI, err = newUIAdapter(ctx, ui)
	if err != nil {
		return err
	}
	if err := theGame.Init(uiadapter.SingleUI{Printer: mobileUI}, appConfig.Game); err != nil {
		return err
	}
	theGame.RegisterAllRequestObserver(mobileUI)
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
	if theGame == nil {
		panic("Main(): nil game state")
	}
	go func() {
		// start game engine
		err := theGame.Main()
		appContext.NotifyQuit(err)
	}()
}

func Quit() {
	initialized = false

	if mobileUI != nil {
		theGame.UnregisterAllRequestObserver()
		mobileUI = nil
	}
	if theGame != nil {
		theGame.Quit()
		theGame = nil
	}
	if logCloseFunc != nil {
		logCloseFunc()
		logCloseFunc = nil
	}
	if theGameContext != nil {
		// cancel is called by logCloseFunc
		theGameContext = nil
	}
}

func SendCommand(cmd string) {
	if !initialized {
		panic("SendCommand(): Init() must be called firstly")
	}
	theGame.Send(input.NewEventCommand(cmd))
}

func SendSkippingWait() {
	if !initialized {
		panic("SendSkippingWait(): Init() must be called firstly")
	}
	theGame.Send(input.NewEventControl(input.ControlStartSkippingWait))
}

func SendStopSkippingWait() {
	if !initialized {
		panic("SendStopSkippingWait(): Init() must be called firstly")
	}
	theGame.Send(input.NewEventControl(input.ControlStopSkippingWait))
}

func SetViewSize(lineCount, lineRuneWidth int) error {
	if !initialized {
		panic("SetViewSize(): Init() must be called firstly")
	}
	return mobileUI.editor.SetViewSize(lineCount, lineRuneWidth)
}
