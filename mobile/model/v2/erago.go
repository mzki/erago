package model

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mzki/erago"
	"github.com/mzki/erago/app/config"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
	"github.com/mzki/erago/util/log"
	"golang.org/x/image/math/fixed"
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

	mainDone chan struct{} = nil
)

type InitOptions struct {
	// ImageFetchType indicates how game engine notifies image data to UI.
	// It should either one of ImageFetch* values. e.g. ImageFetchRawRGBA.
	ImageFetchType int

	// MessageByteEncoding indicates which byte encoding is used to notify struct
	// data to UI side on calling APIs of UI interface, typically OnPublishXXX.
	MessageByteEncoding int

	// FileSystem is used for reading and writing files for erago package files.
	// It can be nil, in that case OS default filesystem is used.
	FileSystem FileSystemGlob
}

func Init(ui UI, baseDir string, options *InitOptions) error {
	if initialized {
		panic("game already initialized")
	}

	// setup mobile filesystem to properly access resources.
	var absBaseDir string
	if filepath.IsAbs(baseDir) {
		absBaseDir = baseDir
	} else {
		var err error
		absBaseDir, err = filepath.Abs(baseDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for: %v, %w", baseDir, err)
		}
	}

	// setup filesystem
	mobileFS := createMobileFS(absBaseDir, options.FileSystem)
	filesystem.Default = mobileFS // replace file system used by erago

	// load config file
	configPath, err := mobileFS.ResolvePath(config.ConfigFile)
	if err != nil {
		theErr := fmt.Errorf("can not use base directory %v. err: %w", baseDir, err)
		return theErr
	}
	appConfig, err := config.LoadConfigOrDefault(configPath)
	switch err {
	case nil, config.ErrDefaultConfigGenerated:
	default:
		theErr := fmt.Errorf("config load error: %w", err)
		return theErr
	}
	confChanged, confChangedMsg := disableDesktopFeatures(appConfig)
	// confChanged and its message will be handled after log setup done.

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
		closeFunc, err := config.SetupLogConfig(appConfig)
		if err != nil {
			theErr := fmt.Errorf("log configure error: %w", err)
			return theErr
		}
		// just store it, called on Quit()
		closeFuncs = append(closeFuncs, closeFunc)

		// now, log destination is activated, and can be used.
	}
	if confChanged {
		log.Infof("Config changed for mobile specific: %v", confChangedMsg)
	}

	// create game instance
	ctx, cancel := context.WithCancel(context.Background())
	theGameContext = ctx
	closeFuncs = append(closeFuncs, cancel)

	theGame = erago.NewGame()
	mobileUI, err = newUIAdapter(ctx, ui, uiAdapterOptions{
		ImageFetchType:      pbImageFetchType(options.ImageFetchType),
		MessageByteEncoding: options.MessageByteEncoding,
	})
	if err != nil {
		theErr := fmt.Errorf("UIAdapter construction failed: %w", err)
		log.Infof("%v", theErr)
		return theErr
	}
	if err := theGame.Init(uiadapter.SingleUI{Printer: mobileUI}, appConfig.Game); err != nil {
		theErr := fmt.Errorf("game initialization failed: %w", err)
		log.Infof("%v", theErr)
		return theErr
	}
	theGame.RegisterAllRequestObserver(mobileUI)
	mainDone = nil // indicates it is not running yet.
	initialized = true
	return nil
}

func disableDesktopFeatures(appConf *config.Config) (changed bool, message string) {
	msgList := []string{}
	// ReloadFileChange must be disabled since User interact with single application at the moment,
	// User does not do play game and edit script in parall. This feature should be for Desktop and developer only.
	if appConf.Game.ScriptConfig.ReloadFileChange {
		appConf.Game.ScriptConfig.ReloadFileChange = false
		changed = true
		msgList = append(msgList, "Game.Script.ReloadFileChange = false")
	}
	// LogFile must be file rather than stdout or stderr since User can not see
	// stdout nor stderr output in normal way.
	if appConf.LogFile != config.DefaultLogFile {
		appConf.LogFile = config.DefaultLogFile
		changed = true
		msgList = append(msgList, "LogFile = "+appConf.LogFile)
	}
	// LogLimitMegaByte should be less than or equal to Default limit size (10MB)
	// since larger size would fill up storage which is relatively small than desktop.
	if appConf.LogLimitMegaByte > config.DefaultLogLimitMegaByte {
		appConf.LogLimitMegaByte = config.DefaultLogLimitMegaByte
		changed = true
		msgList = append(msgList, fmt.Sprintf("LogLimitMegaByte = %v", appConf.LogLimitMegaByte))
	}
	message = strings.Join(msgList, ",")
	return
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
	if mainDone != nil {
		panic("Main(): already running")
	}
	mainDone = make(chan struct{})
	go func(game *erago.Game) {
		// start game engine
		err := game.Main()
		if err != nil {
			theErr := fmt.Errorf("Game.Main() failed: %w", err)
			log.Infof("%v", theErr)
		}
		close(mainDone)
		appContext.NotifyQuit(err)
	}(theGame) // evaluate current value to avoid race condition inside goroutne.
}

func Quit() {
	initialized = false

	if theGame != nil {
		theGame.Quit()

		// make sure game.Main goroutine is done when Main() was called.
		if mainDone != nil {
			const waitTime = 3 * time.Second
			select {
			case <-time.After(waitTime):
				log.Infof("game.Main is not stopped for %v, attempt to force quit", waitTime)
			case <-mainDone:
				// OK
			}
			mainDone = nil
		}
		// unregister should call after game.Main is done since it is not goroutine safe.
		theGame.UnregisterAllRequestObserver()
		theGame = nil
	}
	if mobileUI != nil {
		mobileUI = nil
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

func SetTextUnitPx(textUnitWidthPx, textUnitHeightPx float64) error {
	if !initialized {
		panic("SetViewSize(): Init() must be called firstly")
	}
	return mobileUI.editor.SetTextUnitPx(fixed.Point26_6{
		X: floatToFixedInt(textUnitWidthPx),
		Y: floatToFixedInt(textUnitHeightPx),
	})
}
