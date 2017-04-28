package erago

import (
	"fmt"
	"runtime"

	"local/erago/flow/scene"
	"local/erago/flow/script"
	"local/erago/state"
	"local/erago/state/csv"
	"local/erago/uiadapter"
	"local/erago/uiadapter/event/input"
)

//
// Game is entry point of the application.
// It implements Sender interface to send user event from external.
// But Sender is valid after Game.Init(), accsessing it causes panic before initializing.
type Game struct {
	ipr   *script.Interpreter
	scene *scene.SceneManager
	state *state.GameState

	uiAdapter *uiadapter.UIAdapter
	sender    uiadapter.Sender

	config Config
}

// Constructs game object with config. If nil config is given
// use default config insteadly.
func NewGame() *Game {
	return &Game{}
}

// Initialize game by UserInterface and game config.
// It returns error of initializing game.
// The empty game config is ok in which use default game Config.
//
// After this, Game.Sender is available.
func (g *Game) Init(ui uiadapter.UI, config Config) error {
	g.uiAdapter = uiadapter.New(ui)
	g.sender = g.uiAdapter.GetInputPort()

	emptyConf := Config{}
	if emptyConf == config {
		config = NewConfig(DefaultBaseDir)
	}
	g.config = config

	csv_manager := csv.NewCsvManager()
	if err := csv_manager.Initialize(config.CSVConfig); err != nil {
		return err
	}
	gamestate := state.NewGameState(csv_manager, config.StateConfig)

	ui_controller := &struct { // must be pointer because its fields are changed later.
		*uiadapter.UIAdapter
		*scene.SceneManager
	}{
		UIAdapter:    g.uiAdapter,
		SceneManager: nil, // will be set later
	}

	// NOTE: scene and Interpreter has cross reference.
	g.ipr = script.NewInterpreter(gamestate, ui_controller, config.ScriptConfig)
	g.scene = scene.NewSceneManager(g.uiAdapter, g.ipr, gamestate, config.SceneConfig)
	ui_controller.SceneManager = g.scene

	// extract user scripts and register it.
	if err := g.ipr.LoadSystem(); err != nil {
		return err
	}

	return nil
}

// It return input port which is used to send user event.
// But game implements Sender interface, so using this may be special case.
func (g Game) InputPort() uiadapter.Sender {
	if a := g.uiAdapter; a != nil {
		return g.uiAdapter.GetInputPort()
	}
	panic("Game: game is not initialized")
}

// Run game main flow.
// It blocks until causing something of error in the flow.
// So you should use it in the other thread.
//
// Example:
//	go func() {
//		game.Main()
//	}()
//
// It returns nil if game quits correctly, otherwise return erorr containing
// any panic in the flow.
func (g *Game) Main() error {
	return withRecoverRun(g.main)
}

// capture panic as error in this thread
func withRecoverRun(run func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			buf := make([]byte, 4096)
			buf_end := runtime.Stack(buf, false)
			err = fmt.Errorf("panic in game.Main: %v\n%s", rec, string(buf[:buf_end]))
		}
	}()
	return run()
}

func (g *Game) main() error {
	// run filtering user input on other thread.
	go g.uiAdapter.RunFilter()
	defer g.uiAdapter.Quit()

	// run game flow.
	err := g.scene.Run()
	if err == uiadapter.ErrorPipelineClosed {
		return nil
	}
	return err
}

// implements uiadapter.Sender interface.
// send input event to game running. it can be used asynchrobously.
// For more detail for input event, see input package.
func (g *Game) Send(ev input.Event) {
	if s := g.sender; s != nil {
		s.Send(ev)
	}
}

// implements uiadapter.Sender interface.
// add input request chaged ovserver which can be used asynchrobously.
func (g *Game) AddRequestObserver(obs uiadapter.RequestObserver) {
	if s := g.sender; s != nil {
		s.AddRequestObserver(obs)
		return
	}
	panic("Game: game is not initialized")
}

// remove input request chaged ovserver which can be used asynchrobously.
func (g *Game) RemoveRequestObserver(obs uiadapter.RequestObserver) {
	if s := g.sender; s != nil {
		s.RemoveRequestObserver(obs)
		return
	}
	panic("Game: game is not initialized")
}

// implements uiadapter.Sender interface.
// quit game by external.
func (g *Game) Quit() {
	if g.uiAdapter == nil {
		panic("Game: game is not initialized")
	}
	g.uiAdapter.GetInputPort().Quit()
	g.scene.Free()
	g.ipr.Quit()
}
