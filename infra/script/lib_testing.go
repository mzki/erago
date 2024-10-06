package script

import (
	"errors"
	"fmt"

	"github.com/mzki/erago/scene"
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

// +gendoc "Era Module"
// * var era.is_testing: boolean = false
//
// is_testing indicates current running mode is testing or not.
// Some feature is enabled only in testing mode.
//
// is_testing は現在のスクリプト実行がテスト環境か、ゲーム環境かを示します。
// いくつかの機能はテスト環境でのみ使用可能です。
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
	registerInputQueuer(L, tc)
	registerFlowPCall(L, ip)
}

// +gendoc "Era Module"
// * inputQ: InputQueue = era.inputQueue()
//
// It is enabled only in testing mode. It returns InputQueue object which can
// simulate pseudo user input and can be retrieved from era.inputXXX and its variants.
//
// この機能はテスト環境でのみ有効です。
// InputQueue オブジェクトを返します。inputQueueオブジェクトはユーザーの入力を疑似的にシミュレートし、
// era.inputXXX などの入力関数からユーザの入力を取り出すことができます。

func registerInputQueuer(L *lua.LState, inputQ InputQueuer) {
	funcMap := L.SetFuncs(L.NewTable(), inputQueueFuncMap)
	meta := getOrNewMetatable(L, "era_input_queue", map[string]lua.LValue{
		"__index": funcMap,
		// "__newindex": ,
		"__metatable": metaProtectObj,
	})

	ud := L.NewUserData()
	ud.Value = inputQ
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

// +gendoc: "Flow Module"
// * ok: boolean, ...: any = flow.pcall(fun: function, ...: any)
//
// It is enabled only in testing mode. It will call fun with args and catch the jumping execution step by calling
// flow control functions inside fun such as flow.gotoNextScene.
// It returns the execution status of fun at 1st result and the result of fun at 2nd result.
// When the 1st result is true, there is no any calling flow control functions. The 2nd result and later
// are the result of calling fun.
// when the 1st result is false, there is calling flow control functions.
// The 2nd result is string message that shows which flow control is happened.
// NOTE: flow control functions are flow.gotoNextScene, flow.longReturn and flow.quit .
//
// flow.saveScene and flow.LoadScene are complex case. when flow control functions are called inside save/loadScene
// flow.pcall will catch the jumping, In other case save/loadScene will back to the caller position and proceed next
// execution steps.
//
// この機能はテスト環境でのみ有効です。
// 受け取った fun を args を引数に呼び出します。呼び出し中に発生したフロー制御関数による実行ステップのジャンプ
// (例: flow.gotoNextScene) を補足し、実行ステータスとして通知します。
// 1番目の戻り値が true の場合、フロー制御関数の呼び出しはありません。 2番目以降の戻り値は fun を呼び出した結果です。
// 1番目の戻り値が false の場合、フロー制御関数が呼び出されています。 2番目の結果は、どのフロー制御が行われたかを示す文字列メッセージです。
// 備考: フロー制御関数とは flow.gotoNextScene, flow.longReturn, flow.quit です。
//
// flow.saveScene と flow.LoadScene は複雑です。 save/loadScene の中でフロー制御関数が実行された場合、実行ステップの
// ジャンプが補足されます。一方, save/loadScene の元の呼び出し位置に戻るケースでは、通常通り実行ステップが続行されます。

func registerFlowPCall(L *lua.LState, ip *Interpreter) {
	flowMod, ok := mustGetEraModule(L).RawGetString(eraFlowModName).(*lua.LTable)
	if !ok {
		panic(EraModuleName + "." + eraFlowModName + " is not a module")
	}

	flowMod.RawSetString(
		"pcall",
		L.NewFunction(wrapLFlowPCall(ip)),
	)
}

var inputQueueFuncMap = map[string]lua.LGFunction{
	"append":  linputQueueAppend,
	"prepend": linputQueuePrepend,
	"clear":   linputQueueClear,
	"size":    linputQueueSize,
}

// +gendoc "InputQueue"
// * n_inputs: integer = InputQueue:append(user_inputs: table)
//
// append() appends pseudo user inputs into the end of internal queue.
// user_inputs are array of string like {"0", "one", "2"}.
// The returned value is a size of internal queue after appending.
// Each element of user_inputs are retrived from every call of era.inputXXX by its order.
// Note that a number is needed to treat as string, like "0" for zero.
// If the next element of user_inputs in internal queue is not a number then calling inputNum() will
// stuck infinitely and raise timeout error by time expiration.
//
// append() はユーザーの疑似入力を内部の待ち行列の末端に追加します。
// user_inputs は {"0", "one", "2"} のような文字列の配列で表現します
// 返り値は、append()が成功した後の内部の待ち行列の要素数です。
// 待ち行列の要素は、 era.inputXXX によって、その順番通りに取得できます。
// 数値をユーザーの入力として表現する場合も、文字列で表現する必要があります。ゼロの場合は "0" です。
// 待ち行列から取得される次の要素が数値以外である場合、 inputNum() の呼び出しから永久戻ってくることができません。
// タイムアウトによるエラーが発生するまで、スクリプトの実行がブロックされます。
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
			L.ArgError(2, fmt.Sprintf("%d-th element should be string", i))
		}
	}
	L.Push(lua.LNumber(finalN))
	return 1
}

// +gendoc "InputQueue"
// * n_inputs: integer = InputQueue:prepend(user_inputs: table)
//
// prepend() is same as append() except that user_inputs is prepend to
// the begin of internal queue.
//
// prepend() は append() とほぼ同じ機能です。違いは user_inputs が内部の待ち行列の
// 先頭に追加されることです。
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

// +gendoc "InputQueue"
// * InputQueue:clear()
//
// clear() clears all of element in internal queue.
//
// clear() は内部の待ち行列を空にします。
func linputQueueClear(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	inputQ.Clear()
	return 0
}

// +gendoc "InputQueue"
// * n_inputs: integer = InputQueue:size()
//
// size() returns a number of element in internal queue.
//
// size() は内部の待ち行列の要素数を返します。
func linputQueueSize(L *lua.LState) int {
	inputQ := checkInputQueuer(L, 1)
	L.Push(lua.LNumber(inputQ.Size()))
	return 1
}

var flowControlFuncErrorMap = map[error]string{
	scene.ErrorQuit:      "flow.quit",
	scene.ErrorSceneNext: "flow.gotoNextScene",
	errScriptLongReturn:  "flow.longReturn",
}

func wrapLFlowPCall(ip *Interpreter) lua.LGFunction {
	return func(L *lua.LState) int {
		fun := L.CheckFunction(1)
		args := make([]lua.LValue, 0, 2)
		for i := 2; i <= L.GetTop(); i++ {
			v := L.Get(i)
			args = append(args, v)
		}
		base := L.GetTop()

		err := L.CallByParam(lua.P{
			Fn:      fun,
			NRet:    lua.MultRet,
			Protect: true,
		}, args...)
		if err == nil {
			L.Insert(lua.LTrue, base+1)
			return (L.GetTop() - base)
		} else { // something error happend
			if intErr := ip.extractScriptInterruptError(err.Error()); intErr != nil {
				for funcErr, msg := range flowControlFuncErrorMap {
					if errors.Is(intErr, funcErr) {
						L.Push(lua.LFalse)
						L.Push(lua.LString(msg))
						return 2
					}
				}
				// intErr is not result of flow control function. rethrow original error instead of intErr
				// to keep original context. Later it will be catched at root of script call stack.
			}
			raiseErrorE(L, err)
			return 0
		}
	}
}
