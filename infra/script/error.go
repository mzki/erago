package script

import (
	"context"
	"strings"

	"local/erago/scene"
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

var (
	// script context is canceled.
	scriptCanceledMessage = context.Canceled.Error()
)

// check whether error is special case,
// and return corresponding error, if not matched return error through.
//
// NOTE: current implementation of gopher-lua does not return error context
// directly, the error wrapped by gopher-lua's context.
// Therefore we use string comparision to detect error context instead of error type assertion.
func checkSpecialError(err error) error {
	if err == nil {
		return nil
	}

	mes := err.Error()
	switch {
	case strings.Contains(mes, ScriptQuitMessage):
		return scene.ErrorQuit
	case strings.Contains(mes, ScriptGoToNextSceneMessage):
		return scene.ErrorSceneNext
	case strings.Contains(mes, ScriptLongReturnMessage):
		return nil
	case strings.Contains(mes, scriptCanceledMessage):
		return context.Canceled
	}
	return err
}
