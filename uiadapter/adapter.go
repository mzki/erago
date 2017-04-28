// package uiadapter converts UI interface to model interface
package uiadapter

import (
	"local/erago/uiadapter/event/input"
)

// UIAdapter converts input and output data in cannonical manner.
type UIAdapter struct {
	*inputPort
	*outputPort
}

func New(ui UI) *UIAdapter {
	return &UIAdapter{
		inputPort:  newInputPort(),
		outputPort: newOutputPort(ui),
	}
}

// Input Port interface.
type Sender interface {
	// send input event to app.
	Send(ev input.Event)
	// set listener for changing input request type.
	AddRequestObserver(RequestObserver)
	// remove listener for changing input request type.
	RemoveRequestObserver(RequestObserver)
	// short hand for Send(QuitEvent).
	Quit()
}

// get input interface
func (ad UIAdapter) GetInputPort() Sender {
	return ad.inputPort
}

// print text then wait any input.
func (ad UIAdapter) PrintW(s string) error {
	ad.PrintL(s)
	return ad.Wait()
}

// print text to view spcfied by name then wait any input.
func (ad UIAdapter) VPrintW(vname, s string) error {
	if err := ad.VPrintL(vname, s); err != nil {
		return err
	}
	return ad.Wait()
}
