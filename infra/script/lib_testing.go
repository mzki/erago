package script

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// InputQueuer is an interface for the queue of user input commands.
// It's values can be retrieved from GameController.InputXXX.
type InputQueuer interface {
	Append(x string) (n int)
	Prepend(x string) (n int)
	Clear()
	Size() int
}

// TestingController is an interface to define features accesible only testing mode.
type TestingController interface {
	InputQueuer
}

// TODO gendoc
const eraModuleKeyIsTesting = "is_testing"

func registerIsTesting(L *lua.LState, isTesting bool) {
	eraMod := mustGetEraModule(L)
	eraMod.RawSetString(eraModuleKeyIsTesting, lua.LBool(isTesting))
}

const regKeyTestingController = "era_testing_controller"

// OpenTestingLibs enables the features used for only testing.
// Such features are useful for development purpose such as unit testing.
// The test feature is disabled at default.
func (ip *Interpreter) OpenTestingLibs(tc TestingController) {
	L := ip.vm
	registerIsTesting(L, true)

	funcMap := L.SetFuncs(L.NewTable(), inputQueueFuncMap)
	meta := getOrNewMetatable(L, "era_input_queue", map[string]lua.LValue{
		"__index": funcMap,
		// "__newindex": ,
		"__metatable": metaProtectObj,
	})

	ud := L.NewUserData()
	ud.Value = tc
	ud.Metatable = meta
	reg := L.Get(lua.RegistryIndex).(*lua.LTable)
	reg.RawSetString(regKeyTestingController, ud)

	mustGetEraModule(L).RawSetString(
		"inputQueue",
		L.NewFunction(
			lua.LGFunction(func(L *lua.LState) int {
				L.Push(ud)
				return 1
			}),
		),
	)
}

var inputQueueFuncMap = map[string]lua.LGFunction{
	"append":  linputQueueAppend,
	"prepend": linputQueuePrepend,
	"clear":   linputQueueClear,
	"size":    linputQueueSize,
}

func checkInputQueuer(L *lua.LState, pos int) InputQueuer {
	ud := L.CheckUserData(pos)
	tc, ok := ud.Value.(InputQueuer)
	if !ok {
		L.ArgError(pos, "Invalid type. Expect InputQueue, but not.")
	}
	return tc
}

func linputQueueAppend(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	commands := L.CheckTable(2)
	finalN := 0
	for i := 1; i <= commands.MaxN(); i++ {
		switch cmd := commands.RawGetInt(i); cmd.Type() {
		case lua.LTString:
			finalN = inputQ.Append(lua.LVAsString(cmd))
		default:
			L.ArgError(2, fmt.Sprintf("%d-th element is string", i))
		}
	}
	L.Push(lua.LNumber(finalN))
	return 1
}

func linputQueuePrepend(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	commands := L.CheckTable(2)
	finalN := 0
	// reversed order. {1, 2, 3} -> 3 + {}, 2 + {3}, 1 + {2, 3}
	for i := commands.MaxN(); i >= 1; i-- {
		switch cmd := commands.RawGetInt(i); cmd.Type() {
		case lua.LTString:
			finalN = inputQ.Prepend(lua.LVAsString(cmd))
		default:
			L.ArgError(2, fmt.Sprintf("%d-th element is not string", i))
		}
	}
	L.Push(lua.LNumber(finalN))
	return 1
}

func linputQueueClear(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	inputQ.Clear()
	return 0
}

func linputQueueSize(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	L.Push(lua.LNumber(inputQ.Size()))
	return 1
}
