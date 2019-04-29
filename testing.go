package erago

import (
	"fmt"

	"github.com/mzki/erago/stub"
)

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
	if err := game.Init(stub.NewFlowGameController(), conf); err != nil {
		return fmt.Errorf("Game.Init() Fail: %v", err)
	}
	defer game.Quit()

	err := withRecoverRun(func() error {
		for _, s := range script_files {
			if err := game.ipr.DoFile(s); err != nil {
				// TODO: integrates multiple errors for scripts as one error?
				return fmt.Errorf("script(%s) Fail: %v", s, err)
			}
		}
		return nil
	})
	return err
}
