package model

import (
	"github.com/mzki/erago"
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
	game        *erago.Game
	mobileUI    *uiAdapter
	initialized = false
)

func Init(ui UI, baseDir string) error {
	if initialized {
		panic("game already initialized")
	}

	game = erago.NewGame()
	mobileUI = newUIAdapter(ui)
	if err := game.Init(uiadapter.SingleUI{mobileUI}, erago.NewConfig(baseDir)); err != nil {
		return err
	}
	game.AddRequestObserver(mobileUI)
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
		err := game.Main()
		appContext.NotifyQuit(err)
	}()
}

func Quit() {
	if !initialized {
		panic("Quit(): Init() must be called firstly")
	}
	initialized = false

	if mobileUI != nil {
		game.RemoveRequestObserver(mobileUI)
		mobileUI = nil
	}
	if game != nil {
		game.Quit()
		game = nil
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
