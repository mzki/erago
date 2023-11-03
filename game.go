package erago

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/mzki/erago/infra/repo"
	"github.com/mzki/erago/infra/script"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
)

// Game is entry point of the application.
// It implements Sender interface to send user event from external.
// Sender is valid after Game.Init(), so accsessing it causes panic before initializing.
//
// Game object is invalid after retruned from Game.Main() or calling Game.Quit().
// To reuse it, you must call Game.Init() first, then call Main().
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

	if emptyConf := (Config{}); emptyConf == config {
		config = NewConfig(DefaultBaseDir)
	}
	g.config = config

	csv_manager := csv.NewCsvManager()
	if err := csv_manager.Initialize(config.CSVConfig); err != nil {
		return err
	}
	gamestate := state.NewGameState(csv_manager, repo.NewFileRepository(csv_manager, config.RepoConfig))
	g.state = gamestate

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

	// register some special scenes
	g.scene.RegisterSceneFunc(sceneNameBooting, g.sceneBooting)

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
//
//	go func() {
//		game.Main(ctx)
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

// erago starts from boot scene
const startSceneName = sceneNameBooting

func (g *Game) main() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run filtering user input on other thread.
	go g.uiAdapter.RunFilter(ctx)
	defer g.uiAdapter.Quit()

	// finalize game flows to avoid reference cycle.
	defer g.scene.Free()
	defer g.ipr.Quit()

	// run game flow.
	g.ipr.SetContext(ctx)
	err := g.scene.Run(ctx, startSceneName)
	if errors.Is(err, uiadapter.ErrorPipelineClosed) {
		return nil
	}
	return err
}

// implements uiadapter.Sender interface.
// quit game by external.
func (g *Game) Quit() {
	if g.uiAdapter == nil {
		panic("Game: game is not initialized")
	}
	g.uiAdapter.Quit()
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
func (g *Game) RegisterRequestObserver(typ uiadapter.InputRequestType, obs uiadapter.RequestObserver) {
	if s := g.sender; s != nil {
		s.RegisterRequestObserver(typ, obs)
		return
	}
	panic("Game: game is not initialized")
}

// remove input request chaged ovserver which can be used asynchrobously.
func (g *Game) UnregisterRequestObserver(typ uiadapter.InputRequestType) {
	if s := g.sender; s != nil {
		s.UnregisterRequestObserver(typ)
		return
	}
	panic("Game: game is not initialized")
}

// helper function to register handler for all of input request type.
func (g *Game) RegisterAllRequestObserver(obs uiadapter.RequestObserver) {
	if s := g.sender; s != nil {
		uiadapter.RegisterAllRequestObserver(s, obs)
		return
	}
	panic("Game: game is not initialized")
}

// helper function to unregister handler for all of input request type.
func (g *Game) UnregisterAllRequestObserver() {
	if s := g.sender; s != nil {
		uiadapter.UnregisterAllRequestObserver(s)
		return
	}
	panic("Game: game is not initialized")
}
