// package uiadapter converts UI interface to model interface
package uiadapter

import (
	"github.com/mzki/erago/uiadapter/event/input"
)

// UIAdapter converts input and output data in cannonical manner.
type UIAdapter struct {
	*inputPort
	*outputPort
}

func New(ui UI) *UIAdapter {
	syncer := &lineSyncer{ui}
	return &UIAdapter{
		inputPort:  newInputPort(syncer),
		outputPort: newOutputPort(ui, syncer),
	}
}

// Input Port interface.
type Sender interface {
	// send input event to app.
	Send(ev input.Event)
	// register listener for changing input request type.
	RegisterRequestObserver(InputRequestType, RequestObserver)
	// unregister listener for changing input request type.
	UnregisterRequestObserver(InputRequestType)
	// short hand for Send(QuitEvent).
	Quit()
}

// get input interface
func (ad UIAdapter) GetInputPort() Sender {
	return ad.inputPort
}

//
// Functions for crossing input/output ports.
//

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

//
// Synchronize feature.
//

type lineSyncer struct {
	s Syncer
	// TODO: holds buffering state? For Unbuffered, LineBuffered or FullBuffered.
	// Currently implements only LineBuffered.
}

// SyncText should be called when any output text exist.
func (ls *lineSyncer) SyncText() error {
	// TODO if ls.Unbuffered {
	//        return s.Sync()
	//      }

	return nil // no operation
}

// SyncLine should be called when output text encounter newline "\n".
func (ls *lineSyncer) SyncLine() error {
	// TODO if !ls.LineBuffered { return nil } // no operation

	return ls.s.Sync()
}

// SyncWait should be called when program waiting for any user input.
func (ls *lineSyncer) SyncWait() error {
	return ls.s.Sync()
}

// Sync is equivalent to call of Syncer.Sync(), which is called any buffered state.
// It exist to utilize interface for SyncXXX functions.
func (ls *lineSyncer) Sync() error {
	return ls.s.Sync()
}
