package ui

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/mobile/event/paint"

	"github.com/mzki/erago"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
)

//
// CmdSender is interface for sending information to the model.
type CmdSender interface {
	SendCommand(cmd string)
	SendRawCommand(r rune)
	SendControlSkippingWait(enable bool)
}

// EragoPresenter is mediator between erago.Game model and GUI widgets.
// It sends PresenterTask and ModelError for eventQ, user should handle these.
type EragoPresenter struct {
	game        *erago.Game
	gameRunning bool
	gameRestart chan<- struct{}
	gameDone    <-chan struct{}

	eventQ screen.EventDeque

	ui               uiadapter.UI
	requestObservers []uiadapter.RequestObserver

	syncTimer    *time.Timer
	syncInterval time.Duration

	mu           *sync.Mutex
	inputRequest uiadapter.InputRequestType // under mutex
}

const defaultSyncInterval = 1 * time.Second / 120 // 120 op/sec

var (
	// ErrorGameQuitByRestartRequest indicates the game thread ends by restart request.
	// It is notified via ModelErorr and can be obtained from ModelError.Cause().
	ErrorGameQuitByRestartRequest = fmt.Errorf("Quit model thread by restart request")

	// ErrorGameAlreadyRunning indicates the game thread already running but
	// invalid operations such as starting the game are arrived.
	ErrorGameAlreadyRunning = fmt.Errorf("game already running")
)

//
func NewEragoPresenter(eventQ screen.EventDeque) *EragoPresenter {
	if eventQ == nil {
		panic("nil argument is not allowed")
	}
	return &EragoPresenter{
		game:         erago.NewGame(),
		eventQ:       eventQ,
		syncTimer:    time.NewTimer(defaultSyncInterval),
		syncInterval: defaultSyncInterval,
		mu:           new(sync.Mutex),
	}
}

func (p *EragoPresenter) sync(n node.Node) {
	select {
	case <-p.syncTimer.C: // wait for synchronizing signal
		p.syncTimer.Reset(p.syncInterval)
	}
}

// add uiadapter.RequestObserver to notify inputRequest is changed.
// it is valid before RunGameThread() and not used concurrently.
func (p *EragoPresenter) AddRequestObserver(obs uiadapter.RequestObserver) {
	if p.gameRunning {
		return
	}
	p.requestObservers = append(p.requestObservers, obs)
}

// implements uiadapter.RequestObserver
func (p *EragoPresenter) OnRequestChanged(typ uiadapter.InputRequestType) {
	for _, obs := range p.requestObservers {
		obs.OnRequestChanged(typ)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inputRequest = typ
}

// return any input event is waitng?
func (p *EragoPresenter) InputWaiting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch p.inputRequest {
	case uiadapter.InputRequestInput, uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
		return true
	default:
		return false
	}
}

// return command event is waitng?
func (p *EragoPresenter) CommandWaiting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch p.inputRequest {
	case uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
		return true
	default:
		return false
	}
}

// run game main thread on other gorutine.
// return true if staring at first time, false if multiple calling this.
func (p *EragoPresenter) RunGameThread(ui uiadapter.UI, Conf erago.Config) error {
	if p.gameRunning {
		return ErrorGameAlreadyRunning
	}
	p.gameRunning = true

	if ui == nil {
		panic("EragoPresenter.RunGameThread(): nil UI is not allowed")
	}

	p.ui = ui

	gameDone := make(chan struct{})
	gameRestart := make(chan struct{}, 1) // +1 for pending request
	p.gameDone = gameDone
	p.gameRestart = gameRestart
	go func() {
		defer close(gameDone)
		// initializing game model
		if err := p.init(Conf); err != nil {
			p.onErrorInModel(err)
			p.notifyQuitApp(err)
			return
		}
		defer p.game.Quit()

		p.game.RegisterAllRequestObserver(p)
		defer p.game.UnregisterAllRequestObserver()

		// run game model's main.
		if err := p.game.Main(); err != nil {
			p.onErrorInModel(err)
			p.notifyQuitApp(err)
		} else {
			select {
			case <-gameRestart:
				// send error with restart request
				p.notifyQuitApp(ErrorGameQuitByRestartRequest)
			default:
				p.notifyQuitApp(nil) // send quiting signal without error.
			}
		}
	}()
	return nil
}

// RestartGameThread restarts game thread with another game instance.
// Exsiting game thread is stopped and diposed.
// It returns nil when restart succeeded and also notifies
// ModelError.Cause() == ErrorGameQuitByRestartRequest through the event queue.
// Otherwise returns error.
func (p *EragoPresenter) RestartGameThread(ui uiadapter.UI, conf erago.Config) error {
	if p.gameRunning {
		// stop game thread safely and recreate game intance.

		select {
		case p.gameRestart <- struct{}{}:
			// OK. do nothing
		default:
			return fmt.Errorf("GameRestart request can not be sent. invalid state.")
		}
		const QuitTimeout = 10 * time.Second
		p.Quit()
		select {
		case <-p.gameDone:
			// OK. do nothing
		case <-time.After(QuitTimeout):
			return fmt.Errorf("GameQuit signal timeout for %v", QuitTimeout)
		}
		// create new game instance and dispose old.
		p.game = erago.NewGame()
	}
	return p.RunGameThread(ui, conf)
}

func (p *EragoPresenter) init(Conf erago.Config) error {
	ui := p.ui
	return p.game.Init(ui, Conf)
}

// show error message to ui. quiting signal is not sent.
func (p *EragoPresenter) onErrorInModel(err error) {
	if err == nil {
		return
	}
	if ui := p.ui; ui != nil {
		ui.PrintLine("#")
		ui.Print(err.Error())
		ui.Print("\n終了します\n")
		p.eventQ.Send(paint.Event{})
	}
}

// notify signal to quit application.
// a cause nil means quiting correctly.
func (p *EragoPresenter) notifyQuitApp(cause error) {
	p.eventQ.Send(ModelError{cause})
}

// send quit signal to the internal model execution.
// it must be called after RunGameThread().
func (p *EragoPresenter) Quit() {
	p.game.Send(input.NewEventQuit())
	p.gameRunning = false
}

// Mark any node.Marks, NeedsPaint etc, to node n. It is queued in main event queue
// and execute on UI thread, not execute immdiately.
func (p *EragoPresenter) Mark(n node.Node, mark node.Marks) {
	p.eventQ.Send(PresenterTask(func() {
		if n.Wrappee().Marks&mark == 0 {
			n.Mark(mark)
		}
	}))
}

// implements CmdSender interface.
func (p *EragoPresenter) SendCommand(cmd string) {
	p.game.Send(input.NewEventCommand(cmd))
}

// implements CmdSender interface.
func (p *EragoPresenter) SendRawCommand(r rune) {
	p.game.Send(input.NewEventRawInput(r))
}

// implements CmdSender interface.
func (p *EragoPresenter) SendControlSkippingWait(enable bool) {
	control := input.ControlStartSkippingWait
	if !enable {
		control = input.ControlStopSkippingWait
	}
	p.game.Send(input.NewEventControl(control))
}

// ModelError reperesents error in the Game model.
// The Game also returns error(nil) if quiting it correctly.
// To distinguish quiting signal and an error, check ModelError.HasCause().
// It means error that having a cause, quiting signal otherwise.
type ModelError struct {
	err error
}

// implements error interface.
func (me ModelError) Error() string {
	return "Game Execution Error:\n" + me.err.Error()
}

// whether model's error has cause? model error if cause exist, quiting signal otherwise.
func (me ModelError) HasCause() bool {
	return me.err != nil
}

// Cause returns error which caused this ModelError,
// or may return nil when game thread quiting correctly.
func (me ModelError) Cause() error {
	return me.err
}

// Presenter pushes asynchronized task for screen.EventDeque.
// it should handled on UI thread.
type PresenterTask func()

// execute task.
func (task PresenterTask) Run() {
	task()
}
