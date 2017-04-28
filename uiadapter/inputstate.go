package uiadapter

import (
	"local/erago/uiadapter/event/input"
)

// represents state of input handling.
type inputState interface {
	// enter the state.
	Enter(p *inputPort)

	// transition to next state.
	NextState(p *inputPort, ev interface{}) inputState

	// exit the state.
	Exit(p *inputPort)

	// self type
	Type() inputStateType
}

type inputStateType uint8

const (
	typeInputIdling inputStateType = iota
	typeCommandWaiting
	typeInputWaiting
	typeWaitSkipping
	typeRawInputWaiting

	typeMacroNothing
	typeMacroRunning
)

func (t inputStateType) String() string {
	switch t {
	case typeInputIdling:
		return "idling"
	case typeCommandWaiting:
		return "command waiting"
	case typeInputWaiting:
		return "input waiting"
	case typeWaitSkipping:
		return "wait skipping"
	case typeRawInputWaiting:
		return "raw input waiting"
	case typeMacroNothing:
		return "macro nothing"
	case typeMacroRunning:
		return "macro running"
	default:
		return "unknown state"
	}
}

// used for uiadapter internally.
type (
	internalEvent struct {
		Type internalEventType
	}

	internalEventType uint8
)

func (t internalEventType) New() internalEvent {
	return internalEvent{t}
}

const (
	internalEventNone internalEventType = iota

	// start/stop inputting of any user command.
	internalEventStartCommand
	internalEventStopCommand

	// start/stop any just user input.
	// it treats only whether user confirming?
	internalEventStartInput
	internalEventStopInput

	// start/stop any user raw input.
	internalEventStartRawInput
	internalEventStopRawInput

	// notify macro is done.
	internalEventMacroDone
)

// it is inherited by any child of inputState.
// it do nothing for any user input, entering the state, and  exiting the state.
type baseInputState struct{}

func (s baseInputState) Enter(p *inputPort) {}
func (s baseInputState) Exit(p *inputPort)  {}

// it waits signal of starting input only, in which any user input is ignored.
type inputIdling struct {
	baseInputState
}

func (s inputIdling) Type() inputStateType { return typeInputIdling }

func (s inputIdling) Enter(p *inputPort) {
	p.requestChanged(InputRequestNone)
}

func (s inputIdling) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		// ignore any event.
	case internalEvent:
		switch ev.Type {
		case internalEventStartCommand:
			return commandWaiting{}
		case internalEventStartInput:
			return inputWaiting{}
		case internalEventStartRawInput:
			return rawInputWaiting{}
		}
	}
	return s
}

// waiting for any user command by user confirming.
type commandWaiting struct {
	baseInputState
}

func (s commandWaiting) Type() inputStateType { return typeCommandWaiting }

func (s commandWaiting) Enter(p *inputPort) {
	p.requestChanged(InputRequestCommand)
	p.cbuf.Clear()
}

func (s commandWaiting) Exit(p *inputPort) {
	p.cbuf.Clear()
	if p.cbuf.MacroSize() == 0 {
		p.ebuf.SendFirst(internalEventMacroDone)
	}
}

func (s commandWaiting) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		if ev.Type == input.EventCommand {
			p.cbuf.Send(ev.Command)
		}
	case internalEvent:
		if ev.Type == internalEventStopCommand {
			return inputIdling{}
		}
	}
	return s
}

// wating for just inputting which is typically emittion of empty command.
type inputWaiting struct {
	baseInputState
}

func (s inputWaiting) Type() inputStateType { return typeInputWaiting }

func (s inputWaiting) Enter(p *inputPort) {
	p.requestChanged(InputRequestInput)
	p.cbuf.Clear()
}

func (s inputWaiting) Exit(p *inputPort) {
	p.cbuf.Clear()
}

func (s inputWaiting) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		switch ev.Type {
		case input.EventControl:
			if ev.Control == input.ControlStartSkippingWait {
				return waitSkipping{}
			}
		case input.EventCommand:
			p.cbuf.Send(ev.Command)
		}
	case internalEvent:
		if ev.Type == internalEventStopInput {
			return inputIdling{}
		}
	}
	return s
}

// skipping wait().
type waitSkipping struct {
	baseInputState
}

func (s waitSkipping) Enter(p *inputPort) {
	p.requestChanged(InputRequestNone)
	p.cbuf.Send("") // TODO: awake pending of command buffer by signal only?
}

func (s waitSkipping) Exit(p *inputPort) {
}

func (s waitSkipping) Type() inputStateType { return typeWaitSkipping }

func (s waitSkipping) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		switch ev.Type {
		case input.EventControl:
			if ev.Control == input.ControlStopSkippingWait {
				return inputIdling{}
			}
		case input.EventCommand:
			// do nothing
		}
	case internalEvent:
		switch ev.Type {
		case internalEventStartInput:
			p.cbuf.Send("") // awake waiting for any input.
		case internalEventStartCommand:
			return commandWaiting{}
		case internalEventStartRawInput:
			return rawInputWaiting{}
		}
	}
	return s
}

// waiting for raw user input (i.e. pressed key on keyboard).
type rawInputWaiting struct {
	baseInputState
}

func (s rawInputWaiting) Type() inputStateType { return typeRawInputWaiting }

func (s rawInputWaiting) Enter(p *inputPort) {
	p.requestChanged(InputRequestRawInput)
	p.cbuf.Clear()
}

func (s rawInputWaiting) Exit(p *inputPort) {
	p.cbuf.Clear()
}

func (s rawInputWaiting) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		switch ev.Type {
		case input.EventRawInput:
			p.cbuf.Send(ev.Command)
		case input.EventCommand:
			p.cbuf.Send(ev.Command)
		}

	case internalEvent:
		if ev.Type == internalEventStopRawInput {
			return inputIdling{}
		}
	}
	return s
}
