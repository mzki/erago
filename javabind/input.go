package eragoj

import (
	"local/erago/event/input"
	goadapter "local/erago/uiadapter"
)

type InputPort struct {
	sender goadapter.Sender
}

func (ip InputPort) Quit() {
	ip.sender.Quit()
}

func (ip InputPort) SendCommand(cmd string) {
	ip.sender.Send(input.NewEventCommand(cmd))
}

func (ip InputPort) SendControll(controll int8) {
	go_ctrl := controllTypeTable[controll]
	ip.sender.Send(input.NewEventControll(go_ctrl))
}

var controllTypeTable = map[int8]input.ControllType{
	ControllNone:           input.ControllNone,
	ControllSkipWait:       input.ControllSkipWait,
	ControllStopSkipWait:   input.ControllStopSkipWait,
	ControllInterruptMacro: input.ControllInterruptMacro,
}

const (
	ControllNone = iota
	ControllSkipWait
	ControllStopSkipWait
	ControllInterruptMacro
)

func (ip InputPort) SetOnRequestChanged(listener RequestChangedListener) {
	ip.sender.SetOnRequestChanged(func(typ goadapter.InputRequestType) {
		listener.OnRequestChanged(inputRequestTypeTable[typ])
	})
}

type RequestChangedListener interface {
	OnRequestChanged(int8)
}

var inputRequestTypeTable = map[goadapter.InputRequestType]int8{
	goadapter.InputRequestNone:    InputRequestNone,
	goadapter.InputRequestInput:   InputRequestInput,
	goadapter.InputRequestCommand: InputRequestCommand,
	goadapter.InputRequestRaw:     InputRequestRaw,
}

const (
	InputRequestNone = iota
	InputRequestInput
	InputRequestCommand
	InputRequestRaw
)
