package script

import (
	"strings"

	"local/erago/flow"
)

const (
	// detecting error types is by string comparision.

	// non-error but requires quiting the application.
	ScriptQuitMessage = "# EXIT NORMALY #"

	// non-error but interrupts script execution.
	ScriptLongReturnMessage = "# LONG RETURN #"

	// non-error but requires quiting current scene flow and starts next scene.
	ScriptGoToNextSceneMessage = "# GOTO NEXT SCENE #"
)

// check wheather error is special case,
// and return corresponding error, if not matched return given err through.
//
// NOTE: current implementation of gopher-lua does not return error context
// directly, wrap gopher-lua's context.
// Therefore we use string comparision to detect error context instead of error type assertion.
func checkSpecialError(err error) error {
	if err == nil {
		return nil
	}

	mes := err.Error()
	switch {
	case strings.HasPrefix(mes, ScriptQuitMessage):
		return flow.ErrorQuit
	case strings.HasPrefix(mes, ScriptGoToNextSceneMessage):
		return flow.ErrorSceneNext
	case strings.HasPrefix(mes, ScriptLongReturnMessage):
		return nil
	}
	return err
}
