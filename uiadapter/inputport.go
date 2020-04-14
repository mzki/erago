package uiadapter

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/mzki/erago/uiadapter/event/input"
	"github.com/mzki/erago/util/deque"
)

// Waiting functions without timeout feature, such as Wait(), Command() and so on,
// use this timeout threshold which is enough for most cases.
const DefaultMaxWaitDuration = 365 * (24 * time.Hour) // 1 year

// inputport is interface of ui input.
// it filters input.Event and send command to command buffer.
// handling incoming event (user input) is responsible for ebuf,
// and outgoing event (command) is responsible for cbuf.
type inputPort struct {
	syncer *lineSyncer

	state      inputState
	macroState inputState

	ebuf deque.EventDeque
	cbuf *commandBuffer

	requestObservers []RequestObserver

	// this mutex controls the fields below.
	mu     *sync.Mutex
	closed bool // the port is closed?
}

func newInputPort(ls *lineSyncer) *inputPort {
	return &inputPort{
		syncer:     ls,
		state:      inputIdling{},
		macroState: macroNothing{},
		ebuf:       deque.NewEventDeque(),
		cbuf:       newCommandBuffer(),
		mu:         new(sync.Mutex),
		closed:     false,
	}
}

// RequestObserver observes changing input state.
type RequestObserver interface {
	OnRequestChanged(InputRequestType)
}

type RequestObserverFunc func(InputRequestType)

// implements RequestObserver interface.
func (fn RequestObserverFunc) OnRequestChanged(typ InputRequestType) {
	fn(typ)
}

// indicates requesting of current input.
type InputRequestType int8

const (
	// request is none.
	InputRequestNone = iota

	// request command which is confirmed by user.
	InputRequestCommand

	// request just input which is empty command by user confirming.
	InputRequestInput

	// request raw inputting such as pressed key by user.
	InputRequestRawInput
)

//  It can not use concurrently.
func (port *inputPort) RegisterRequestObserver(o RequestObserver) {
	port.requestObservers = append(port.requestObservers, o)
}

//  It can not use concurrently.
func (port *inputPort) UnregisterRequestObserver(o RequestObserver) {
	observers := port.requestObservers
	for i, obs := range observers {
		if obs == o {
			copy(observers[i:], observers[i+1:])
			port.requestObservers = observers[:len(observers)-1]
			return
		}
	}
}

//  It can not use concurrently.
func (port *inputPort) requestChanged(typ InputRequestType) {
	for _, obs := range port.requestObservers {
		obs.OnRequestChanged(typ)
	}
}

// starting filtering input Event.
// It blocks until context is canceled,
// you can use go statement to run other thread.
// After canceling, inputPort is closed and can not be used.
//
// It returns error which indicates what context is canceled by.
func (p *inputPort) RunFilter(ctx context.Context) error {
	doneCh := make(chan struct{}, 1)
	go func() {
		defer close(doneCh)
		for {
			ev := p.ebuf.NextEvent()
			if ev, ok := ev.(input.Event); ok && ev.Type == input.EventQuit {
				doneCh <- struct{}{}
				return
			}

			// update macro state
			macroNext := p.macroState.NextState(p, ev)
			p.updateState(p.macroState, macroNext)
			p.macroState = macroNext

			// update input state
			next := p.state.NextState(p, ev)
			p.updateState(p.state, next)
			p.state = next
		}
	}()

	// Because p.close closes p.ebuf,
	// for-loop in above goroutine is quited at the moment.
	defer p.close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return nil
	}
}

func (p *inputPort) updateState(current, next inputState) {
	if next.Type() != current.Type() {
		current.Exit(p)
		next.Enter(p)
	}
}

func (port *inputPort) isClosed() bool {
	port.mu.Lock()
	defer port.mu.Unlock()
	return port.closed
}

// Close inputport so that any sending event is ignored.
func (port *inputPort) close() {
	port.mu.Lock()
	port.closed = true
	port.mu.Unlock()

	port.cbuf.Close()
	port.ebuf.Send(input.NewEventQuit())
}

// send input event to input port.
func (port *inputPort) Send(ev input.Event) {
	if port.isClosed() {
		return
	}
	port.ebuf.Send(ev)
}

// send quit event signal to terminate RunFilter.
func (port *inputPort) Quit() {
	if port.isClosed() {
		return
	}
	port.ebuf.Send(input.NewEventQuit())
}

//
// funcions for waiting user input.
//

// wait for any input. it will never return until getting any input.
func (port *inputPort) Wait() error {
	return port.WaitWithTimeout(context.Background(), DefaultMaxWaitDuration)
}

// wait for any input with context. it can cancel by cancelation for context.
// it returns error which is uiadapter.ErrorPipelineClosed,
// context.DeadLineExceeded or context.Canceled.
func (port *inputPort) WaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	if port.isClosed() {
		return ErrorPipelineClosed
	}

	// Synchronize to UI state before stating user input
	if err := port.syncer.SyncWait(); err != nil {
		return err
	}

	port.ebuf.Send(internalEventStartInput.New())
	defer port.ebuf.SendFirst(internalEventStopInput.New())

	port.requestChanged(InputRequestInput)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return port.waitWithContext(timeCtx)
}

func (port *inputPort) waitWithContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- port.cbuf.Wait()
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		port.cbuf.Cancel()
		<-errCh // wait for ending goroutine.
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// wait for string command.
func (port *inputPort) Command() (string, error) {
	return port.CommandWithTimeout(context.Background(), DefaultMaxWaitDuration)
}

// wait for string command with context. it can cancel by cancelation for context.
// it returns command string and error which is uiadapter.ErrorPipelineClosed,
// context.DeadLineExceeded or context.Canceled.
func (port *inputPort) CommandWithTimeout(ctx context.Context, timeout time.Duration) (string, error) {
	if port.isClosed() {
		return "", ErrorPipelineClosed
	}

	// Synchronize to UI state before stating user input
	if err := port.syncer.SyncWait(); err != nil {
		return "", err
	}

	port.ebuf.Send(internalEventStartCommand.New())
	defer port.ebuf.SendFirst(internalEventStopCommand.New())

	port.requestChanged(InputRequestCommand)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return port.commandWithContext(timeCtx)
}

func (port *inputPort) commandWithContext(ctx context.Context) (string, error) {
	cmdCh := make(chan struct {
		Cmd string
		Err error
	}, 1)
	go func() {
		cmd, err := port.cbuf.Receive()
		cmdCh <- struct {
			Cmd string
			Err error
		}{Cmd: cmd, Err: err}
		close(cmdCh)
	}()

	select {
	case <-ctx.Done():
		port.cbuf.Cancel()
		<-cmdCh // wait for ending goroutine.
		return "", ctx.Err()
	case cmd := <-cmdCh:
		return cmd.Cmd, cmd.Err
	}
}

// wait for integer command.
func (port *inputPort) CommandNumber() (int, error) {
	return port.CommandNumberWithTimeout(context.Background(), DefaultMaxWaitDuration)
}

// wait for integer command.
func (port *inputPort) CommandNumberWithTimeout(ctx context.Context, timeout time.Duration) (int, error) {
	if port.isClosed() {
		return 0, ErrorPipelineClosed
	}

	// Synchronize to UI state before stating user input
	if err := port.syncer.SyncWait(); err != nil {
		return 0, err
	}

	port.ebuf.Send(internalEventStartCommand.New())
	defer port.ebuf.SendFirst(internalEventStopCommand.New())

	port.requestChanged(InputRequestCommand)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		cmd, err := port.commandWithContext(timeCtx)
		if err != nil {
			return 0, err
		}
		if cmd_no, err := strconv.Atoi(cmd); err == nil {
			return cmd_no, nil
		}
	}
}

// wait for raw input without user confirming, hit enter key.
func (port *inputPort) RawInput() (string, error) {
	return port.RawInputWithTimeout(context.Background(), DefaultMaxWaitDuration)
}

// wait for raw input with timeout.
// raw input does not need user confirming, hit enter key.
// It returns command string and error which is uiadapter.ErrorPipelineClosed or
// context.DeadLineExceeded.
func (port *inputPort) RawInputWithTimeout(ctx context.Context, timeout time.Duration) (string, error) {
	if port.isClosed() {
		return "", ErrorPipelineClosed
	}

	// Synchronize to UI state before stating user input
	if err := port.syncer.SyncWait(); err != nil {
		return "", err
	}

	port.ebuf.Send(internalEventStartRawInput)
	defer port.ebuf.SendFirst(internalEventStopRawInput)

	port.requestChanged(InputRequestRawInput)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return port.commandWithContext(timeCtx)
}

// TODO: below functions should be implemented in the user layer?

// wait for number command that matches given nums.
func (port *inputPort) CommandNumberSelect(ctx context.Context, nums ...int) (int, error) {
	if len(nums) == 0 {
		return port.CommandNumber()
	}
	for {
		got, err := port.CommandNumber()
		if err != nil {
			return 0, err
		}
		for _, n := range nums {
			if n == got {
				return got, nil
			}
		}
	}
}

// wait for number command that mathes in range [min : max]
func (port *inputPort) CommandNumberRange(ctx context.Context, min, max int) (int, error) {
	if min > max {
		return port.CommandNumber()
	}
	for {
		got, err := port.CommandNumber()
		if err != nil {
			return 0, err
		}
		if min <= got && got <= max {
			return got, nil
		}
	}
}
