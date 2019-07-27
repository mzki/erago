package erago

import (
	"context"
	"fmt"
	"time"

	"github.com/mzki/erago/stub"
	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/uiadapter/event/input"
)

const testingMaxTime = 10 * time.Second

// Run testing flow.
//
// It runs script files with Config and returns
// these execution error. nil means all test scripts
// are passed.
// In script execution, input function returns
// no sense values, so the test scripts requiring
// user input can not be tested.
func Testing(conf Config, script_files []string) error {
	game := NewGame()
	if err := game.Init(stub.NewGameUIStub(), conf); err != nil {
		return fmt.Errorf("Game.Init() Fail: %v", err)
	}
	defer game.Quit()

	// create testing context with timeout.
	testCtx, cancel := context.WithTimeout(context.Background(), testingMaxTime)
	defer cancel()

	// run filtering user input on other thread.
	go game.uiAdapter.RunFilter(testCtx)
	defer game.uiAdapter.Quit()

	// add request observer to emulate user input.
	game.AddRequestObserver(&requestObserver{game})

	// do testing
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		err := withRecoverRun(func() error {
			for _, s := range script_files {
				if err := game.ipr.DoFile(s); err != nil {
					// TODO: integrates multiple errors for scripts as one error?
					return fmt.Errorf("script(%s) Fail: %v", s, err)
				}
			}
			return nil
		})
		errCh <- err
	}()

	// wait completion for testing
	select {
	case err := <-errCh:
		return err
	case <-testCtx.Done():
		return testCtx.Err()
	}
}

type requestObserver struct {
	inputPort uiadapter.Sender
}

func (r *requestObserver) OnRequestChanged(typ uiadapter.InputRequestType) {
	switch typ {
	case uiadapter.InputRequestCommand, uiadapter.InputRequestInput, uiadapter.InputRequestRawInput:
		r.inputPort.Send(input.NewEventCommand(""))
	}
}
