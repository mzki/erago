package script

import (
	"context"
	"errors"
	"strings"

	"github.com/mzki/erago/scene"
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

var (
	// ErrWatchDogtimerExpired indicates script execution takes too long time, it may be infinite loop.
	ErrWatchDogTimerExpired = errors.New("script execution takes too long time, may be infinite loop")

	// errScriptLongReturn indicates scrit execution is interrupted and returned to runtime caller.
	errScriptLongReturn = errors.New("script execution is interrupted and returned to caller")
)

// check whether error is special case,
// and return corresponding error, if not matched return error through.
//
// NOTE: current implementation of gopher-lua does not return error context
// directly, the error wrapped by gopher-lua's context.
// Therefore we use string comparision to detect error context instead of error type assertion.
func (ip Interpreter) checkSpecialError(err error) error {
	if err == nil {
		return nil
	}

	if intErr := ip.extractScriptInterruptError(err.Error()); intErr != nil {
		if intErr == errScriptLongReturn {
			// longReturn is consumed at this runtime layer.
			return nil
		} else {
			// to propagate upper runtime layer
			return intErr
		}
	}
	return err
}

func (ip Interpreter) extractScriptInterruptError(mes string) error {
	switch {
	case strings.Contains(mes, ScriptQuitMessage):
		return scene.ErrorQuit
	case strings.Contains(mes, ScriptGoToNextSceneMessage):
		return scene.ErrorSceneNext
	case strings.Contains(mes, ScriptLongReturnMessage):
		return errScriptLongReturn
	case strings.Contains(mes, scriptCanceledMessage):
		if ip.watchDogTimer.IsExpired() {
			return ErrWatchDogTimerExpired
		}
		return context.Canceled
	default:
		return nil
	}
}
