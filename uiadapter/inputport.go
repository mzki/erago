package uiadapter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mzki/erago/uiadapter/event/input"
	"github.com/mzki/erago/util/deque"
	"github.com/mzki/erago/util/log"
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

	requestObservers [inputRequestTypeLen]RequestObserver

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
	// OnRequestChanged is called when user changes input request by calling
	// input APIs such as WaitXXX, CommandXXX and RawInputXXX.
	// This function is called on same context as input APIs.
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
	InputRequestNone InputRequestType = iota

	// request command which is confirmed by user.
	InputRequestCommand

	// request just input which is empty command by user confirming.
	InputRequestInput

	// request raw inputting such as pressed key by user.
	InputRequestRawInput

	// input request size
	inputRequestTypeLen
)

// It can not use concurrently.
func (port *inputPort) RegisterRequestObserver(typ InputRequestType, o RequestObserver) {
	if typ < InputRequestNone || typ >= inputRequestTypeLen {
		panic("invalid input reqeust type")
	}
	port.requestObservers[typ] = o
}

// It can not use concurrently.
func (port *inputPort) UnregisterRequestObserver(typ InputRequestType) {
	if typ < InputRequestNone || typ >= inputRequestTypeLen {
		panic("invalid input reqeust type")
	}
	port.requestObservers[typ] = nil
}

// It can not use concurrently.
func (port *inputPort) requestChanged(typ InputRequestType) {
	if obs := port.requestObservers[typ]; obs != nil {
		obs.OnRequestChanged(typ)
	}
}

// Helper function for register RequestObserver for all of InputRequestType.
// The RequestObservers already registered are overwritten.
func RegisterAllRequestObserver(sender Sender, o RequestObserver) {
	for i := int(InputRequestNone); i < int(inputRequestTypeLen); i++ {
		typ := InputRequestType(i)
		sender.RegisterRequestObserver(typ, o)
	}
}

// Helper function for unregister RequestObserver for all of InputRequestType.
func UnregisterAllRequestObserver(sender Sender) {
	for i := int(InputRequestNone); i < int(inputRequestTypeLen); i++ {
		typ := InputRequestType(i)
		sender.UnregisterRequestObserver(typ)
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
		log.Debugf("inputPort.updateState: current %T, next %T", current, next)
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

// helper function to send internal event with type check
func (port *inputPort) sendInternalEvent(ev internalEvent) {
	port.ebuf.Send(ev)
}

// helper function to send internal event first with type check
func (port *inputPort) sendInternalEventFirst(ev internalEvent) {
	port.ebuf.SendFirst(ev)
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

	port.sendInternalEvent(internalEventStartInput.New())
	defer port.sendInternalEventFirst(internalEventStopInput.New())

	port.requestChanged(InputRequestInput)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return port.waitWithContext(timeCtx)
}

func (port *inputPort) waitWithContext(ctx context.Context) error {
	port.cbuf.PrepareWaitReceive()
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
		if errors.Is(err, ErrorCommandWaitCanceled) {
			return fmt.Errorf("??? waitWithContext is cancelled by external factor: %w", err)
		}
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

	port.sendInternalEvent(internalEventStartCommand.New())
	defer port.sendInternalEventFirst(internalEventStopCommand.New())

	port.requestChanged(InputRequestCommand)
	defer port.requestChanged(InputRequestNone)

	timeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return port.commandWithContext(timeCtx)
}

func (port *inputPort) commandWithContext(ctx context.Context) (string, error) {
	port.cbuf.PrepareWaitReceive()
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
		if errors.Is(cmd.Err, ErrorCommandWaitCanceled) {
			return cmd.Cmd, fmt.Errorf("??? waitWithContext is cancelled by external factor: %w", cmd.Err)
		}
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

	port.sendInternalEvent(internalEventStartCommand.New())
	defer port.sendInternalEventFirst(internalEventStopCommand.New())

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

	port.sendInternalEvent(internalEventStartRawInput.New())
	defer port.sendInternalEventFirst(internalEventStopRawInput.New())

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
