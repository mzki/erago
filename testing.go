package erago

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mzki/erago/stub"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
	"github.com/mzki/erago/util/deque"
)

// DefaultTestingTimeout is default value for testing timeout.
// test will fail when execution time exceed this time.
const DefaultTestingTimeout = 60 * time.Second

// Run testing flow.
//
// It runs script files with Config and returns
// these execution error. nil means all test scripts
// are passed.
// In script execution, input function returns
// no sense values, so the test scripts requiring
// user input can not be tested.
func Testing(conf Config, script_files []string, timeout time.Duration) error {
	game := NewGame()
	if err := game.Init(stub.NewGameUIStub(), conf); err != nil {
		return fmt.Errorf("Game.Init() Fail: %v", err)
	}
	defer game.Quit()

	// create testing context with timeout.
	testCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// run filtering user input on other thread.
	go game.uiAdapter.RunFilter(testCtx)
	defer game.uiAdapter.Quit()

	observer := newRequestObserver(game)

	// add request observer to emulate user input.
	uiadapter.RegisterAllRequestObserver(game, observer)
	defer uiadapter.UnregisterAllRequestObserver(game)

	// Prepare test libraries in interpreter.
	game.ipr.OpenTestingLibs(observer)
	game.ipr.SetContext(testCtx)
	defer game.ipr.Quit()

	// do testing
	return withRecoverRun(func() error {
		for _, s := range script_files {
			// TODO set timeout for each script file?
			if err := game.ipr.DoFile(s); err != nil {
				if errors.Is(err, uiadapter.ErrorPipelineClosed) {
					// indicates timeout error. Add user friendly information.
					err = fmt.Errorf("script execution too long time (>%v):\n%w", DefaultTestingTimeout, err)
				}
				// TODO: integrates multiple errors for scripts as one error?
				return fmt.Errorf("script(%s) Fail: %w", s, err)
			}
			// TODO: show passed file names in progress?
		}
		return nil
	})
}

type requestObserver struct {
	inputPort uiadapter.Sender

	inputQ *inputEventQ
}

func newRequestObserver(port uiadapter.Sender) *requestObserver {
	return &requestObserver{
		inputPort: port,
		inputQ:    &inputEventQ{deque.NewEventDeque(), 0},
	}
}

func (r *requestObserver) OnRequestChanged(typ uiadapter.InputRequestType) {
	switch typ {
	case uiadapter.InputRequestInput:
		r.inputPort.Send(input.NewEventCommand(""))
	case uiadapter.InputRequestCommand, uiadapter.InputRequestRawInput:
		if r.inputQ.Size() <= 0 {
			// inputQueue is empty. send empty command instead to avoid infinite waiting.
			r.inputPort.Send(input.NewEventCommand(""))
		} else {
			ev := r.inputQ.NextEvent()
			r.inputPort.Send(ev)
		}
	}
}

// InputQueuer implementation

func (r *requestObserver) Append(x string) (n int) {
	r.inputQ.Send(input.NewEventCommand(x))
	return int(r.inputQ.Size())
}

func (r *requestObserver) Prepend(x string) (n int) {
	r.inputQ.SendFirst(input.NewEventCommand(x))
	return int(r.inputQ.Size())
}

func (r *requestObserver) Clear()    { r.inputQ.Clear() }
func (r *requestObserver) Size() int { return int(r.inputQ.Size()) }

// inputEventQ wraps deque.EventDeque and its size to be used for single thread context.
type inputEventQ struct {
	deque.EventDeque
	size int64
}

func (q *inputEventQ) NextEvent() input.Event {
	if q.size <= 0 {
		panic("empty queue causes infinite stuck")
	}
	ev := q.EventDeque.NextEvent().(input.Event)
	q.size--
	if q.size < 0 {
		q.size = 0
	}
	return ev
}

func (q *inputEventQ) Clear() {
	for i := 0; i < int(q.size); i++ {
		_ = q.EventDeque.NextEvent()
	}
	q.size = 0
}

func (q *inputEventQ) Send(ev input.Event)      { q.EventDeque.Send(ev); q.size++ }
func (q *inputEventQ) SendFirst(ev input.Event) { q.EventDeque.SendFirst(ev); q.size++ }
func (q *inputEventQ) Size() int64              { return q.size }
