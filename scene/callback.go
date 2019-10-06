package scene

// Scripter handles calling script function in era module.
// These functions can return spcial error defined in the package scene
// In such case, scene flow will be changed by type of special error.
type Scripter interface {
	// oprations for era module.
	EraCall(string) error
	EraCallBoolArgInt(string, int64) (bool, error)
	HasEraValue(string) bool
}

//go:generate go run gen_callback_doc.go
const (
	// separator for the script function name
	ScrSep = "_"

	// script event is optional function. It is OK to not exists.
	// it is called at everywhere
	ScrEventPrefix = "event"

	// script scene is replacement for entire flow in a scene.
	// It requires setting to next scene in the script function.
	ScrScenePrefix = "scene"

	// script replace is replacement function for partial flow in a scene.
	// If it is called, the partial flow of original scene
	// does not through.
	ScrReplacePrefix = "replace"

	// script user is requirement function for implementing flow in a scene.
	// The function of this type must exist and be called from a specific scene.
	// Therefore, the game is aborted if definition is not found.
	ScrUserPrefix = "user"
)

type callBacker struct {
	Scripter
	io IOController
}

// Call script function if exists and return error of
// calling result. If the function is not found
// do nothing and return nil.
func (cb callBacker) maybeCall(fn_name string) error {
	if cb.Scripter.HasEraValue(fn_name) {
		return cb.Scripter.EraCall(fn_name)
	}
	return nil
}

// Call script function with arg int64 if exists and
// return bool and error of calling result.
// return false, nil if function is not found.
func (cb callBacker) maybeCallBoolArgInt(fn_name string, arg int64) (bool, error) {
	if cb.Scripter.HasEraValue(fn_name) {
		return cb.Scripter.EraCallBoolArgInt(fn_name, arg)
	}
	return false, nil
}

// Call script function and return error of
// calling result. If function is not found,
// return not found error.
func (cb callBacker) mustCall(fn_name string) error {
	return cb.Scripter.EraCall(fn_name)
}

// call script function if it exists.
// it returns script function is called?(bool) and
// error from calling result.
func (cb callBacker) checkCall(fn_name string) (bool, error) {
	if cb.Scripter.HasEraValue(fn_name) {
		err := cb.Scripter.EraCall(fn_name)
		return true, err
	}
	return false, nil
}

// call script function if it exists.
// it returns script function is called?(bool) and
// calling result.
func (cb callBacker) checkCallBoolArgInt(fn_name string, arg int64) (struct{ Called, Return bool }, error) {
	if cb.Scripter.HasEraValue(fn_name) {
		ret, err := cb.Scripter.EraCallBoolArgInt(fn_name, arg)
		return struct{ Called, Return bool }{true, ret}, err
	}
	return struct{ Called, Return bool }{}, nil
}

// call script function if it exists, and return
// call result. If not exists, print caution to screen.
func (cb callBacker) cautionCall(fn_name string) error {
	if called, err := cb.checkCall(fn_name); called {
		return err
	}
	return cb.printCaution(fn_name)
}

// call script function if it exists, and return
// call result. If not exists, print caution to screen.
func (cb callBacker) cautionCallBoolArgInt(fn_name string, arg int64) (bool, error) {
	if check, err := cb.checkCallBoolArgInt(fn_name, arg); check.Called {
		return check.Return, err
	}
	return false, cb.printCaution(fn_name)
}

func (cb callBacker) printCaution(fn_name string) error {
	return cb.io.PrintW("era." + fn_name + " is not found.")
}
