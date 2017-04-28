package uiadapter

import (
	"local/erago/uiadapter/event/input"
	"local/erago/uiadapter/macro"
)

type macroNothing struct {
	baseInputState
}

func (s macroNothing) Type() inputStateType { return typeMacroNothing }

func (s macroNothing) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		if ev.Type == input.EventCommand {
			if m, err := macro.Parse(ev.Command); err == nil {
				return macroRunning{macro: m}
			}
		}
	}
	return s
}

// running macro, ignoring any user command.
type macroRunning struct {
	baseInputState
	macro *macro.Macro
}

func (s macroRunning) Enter(p *inputPort) {
	p.cbuf.StartMacro(s.macro)
}

func (s macroRunning) Exit(p *inputPort) {
	p.cbuf.StopMacro()
}

func (s macroRunning) Type() inputStateType { return typeMacroRunning }

func (s macroRunning) NextState(p *inputPort, ev interface{}) inputState {
	switch ev := ev.(type) {
	case input.Event:
		if ev.Type == input.EventControl &&
			ev.Control == input.ControlInterruptMacro {
			return macroNothing{}
		}
	case internalEvent:
		if ev.Type == internalEventMacroDone {
			return macroNothing{}
		}
	}
	return s
}
