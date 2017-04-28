package input

import (
	"errors"
)

// Event has information of inputed command and error etc..
type Event struct {
	Type    EventType
	Control ControlType
	Command string
}

type EventType int8

const (
	EventNone     EventType = iota // dummy event
	EventCommand                   //	emits command by user confirming.
	EventRawInput                  // raw input such as user key press.
	EventControl                   // controll signal for input state..
	EventQuit                      // terminate signal
)

type ControlType int8

const (
	// these are used with EventControl, which controlls
	// scene flow. Typically, EventControl is sent when pushed special
	// key such as Ctrl-Any, Shift-Any, F1, F2, ...,  ESC and etc.
	ControlNone ControlType = iota // dummy

	// starting skipping wait. skip means game flow is running
	// without any waits.
	ControlStartSkippingWait

	// stopping skipping wait. skip means game flow is running
	// without any waits.
	ControlStopSkippingWait

	// stop current running macro.
	ControlInterruptMacro
)

// make new InputEvent type EventCommand. cmd = "" means emitting command nothing.
func NewEventCommand(cmd string) Event {
	return Event{
		Type:    EventCommand,
		Command: cmd,
	}
}

// make new InputEvent type EventControl.
func NewEventControl(ctrl ControlType) Event {
	return Event{
		Type:    EventControl,
		Control: ctrl,
	}
}

// make new input event type EventRawInput.
// it is intended for user key press, so accepts only
// one character or rune.
func NewEventRawInput(r rune) Event {
	return Event{
		Type:    EventRawInput,
		Command: string(r),
	}
}

// make quit event.
// it is intended to send quit siganl.
func NewEventQuit() Event {
	return Event{Type: EventQuit}
}

// error represents quit signal.
var ErrorQuit error = errors.New("input.Event: normal termination")
