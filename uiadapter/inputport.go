package uiadapter

import (
	"strconv"
	"sync"

	"local/erago/uiadapter/event/input"
	"local/erago/util/deque"
)

// inputport is interface of ui input.
// it filters input.Event and send command to command buffer.
type inputPort struct {
	state      inputState
	macroState inputState

	ebuf deque.EventDeque
	cbuf *commandBuffer

	requestObservers []RequestObserver

	// this mutex controls the fields below.
	mu     *sync.Mutex
	closed bool // the port is closed?
}

func newInputPort() *inputPort {
	return &inputPort{
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
func (port *inputPort) AddRequestObserver(o RequestObserver) {
	port.requestObservers = append(port.requestObservers, o)
}

//  It can not use concurrently.
func (port *inputPort) RemoveRequestObserver(o RequestObserver) {
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
func (port inputPort) requestChanged(typ InputRequestType) {
	for _, obs := range port.requestObservers {
		obs.OnRequestChanged(typ)
	}
}

// starting filtering input Event.
// it blocks until calling Quit() or Send(EventQuit).
// you can use go statement to run other thread.
func (p *inputPort) RunFilter() {
	defer p.close()

	for {
		ev := p.ebuf.NextEvent()
		if ev, ok := ev.(input.Event); ok && ev.Type == input.EventQuit {
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

// send quit event signal
func (port *inputPort) Quit() {
	if port.isClosed() {
		return
	}
	port.ebuf.Send(input.NewEventQuit())
}

// wait for any input.
func (port *inputPort) Wait() error {
	if port.isClosed() {
		return ErrorPipelineClosed
	}

	port.ebuf.Send(internalEventStartInput.New())
	defer port.ebuf.SendFirst(internalEventStopInput.New())

	return port.cbuf.Wait()
}

// func (port *inputPort) TWait(d time.Duration) error {
// 	if port.isClosed {
// 		return ErrorPipelineClosed
// 	}
//
// 	port.ebuf.Send(internalEventStartInput.New())
// 	defer port.ebuf.SendFirst(internalEventStopInput.New())
//
// 	timer := time.NewTimer(d)
// 	defer timer.Stop()
//
// 	select {
// 	case _, ok := <- port.cbuf.ReceiveCh():
// 		if !ok {
// 			return ErrorPipelineClosed
// 		}
// 		return nil
// 	case <- timer.C:
// 		return ErrorTimeouted
// 	}
// }

// wait for string command.
func (port *inputPort) Command() (string, error) {
	if port.isClosed() {
		return "", ErrorPipelineClosed
	}

	port.ebuf.Send(internalEventStartCommand.New())
	defer port.ebuf.SendFirst(internalEventStopCommand.New())

	return port.cbuf.Receive()
}

// wait for integer command.
func (port *inputPort) CommandNumber() (int, error) {
	for {
		cmd, err := port.Command()
		if err != nil {
			return 0, err
		}
		if cmd_no, err := strconv.Atoi(cmd); err == nil {
			return cmd_no, nil
		}
	}
}

// wait for number command that mathes given nums.
func (port *inputPort) CommandNumberSelect(nums ...int) (int, error) {
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
func (port *inputPort) CommandNumberRange(min, max int) (int, error) {
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

// wait for raw input such as user press key.
func (port *inputPort) RawInput() (string, error) {
	if port.isClosed() {
		return "", ErrorPipelineClosed
	}

	port.ebuf.Send(internalEventStartRawInput)
	defer port.ebuf.SendFirst(internalEventStopRawInput)

	return port.cbuf.Receive()
}
