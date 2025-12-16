package script

import (
	lua "github.com/yuin/gopher-lua"
)

// references to [ lua-users wiki: Sand Boxes ] http://lua-users.org/wiki/SandBoxes
var unsafeLibs = []string{
	"print", // write stdout is not allowed
	"dofile",
	"dostring",
	"load",
	"loadfile",
	// "loadstring" // it is usable and not accessing filesystem.
}

func knockoutUnsafeLibs(L *lua.LState) {
	for _, func_name := range unsafeLibs {
		L.SetGlobal(func_name, lua.LNil)
	}
}

// extend ipairs to accept userdata with "__next" meta method
func registerExtPairs(L *lua.LState) {
	// assumes LState already registers "ipairs" fucntion
	for _, v := range []struct {
		RawFuncName string
		ExtendFunc  lua.LGFunction
	}{
		{"ipairs", extendIpairs},
		{"pairs", extendPairs},
	} {
		rawFunc := L.GetGlobal(v.RawFuncName)
		if rawFunc == lua.LNil {
			panic(v.RawFuncName + " is not found in LState")
		}
		lextendFunc := L.NewClosure(v.ExtendFunc, rawFunc)
		L.SetGlobal(v.RawFuncName, lextendFunc)
	}
}

func extendIpairs(L *lua.LState) int {
	arg := L.CheckAny(1)
	// lua5.2 meta method extension
	ipairsOp := L.GetMetaField(arg, "__ipairs")

	if ipairsOp == lua.LNil {
		if arg.Type() == lua.LTUserData {
			L.ArgError(1, "userdata has no __ipairs meta method")
			return 0
		}
		ipairsOp = L.CheckFunction(lua.UpvalueIndex(1)) // assumes raw ipairs given
	}

	if err := L.CallByParam(lua.P{
		Fn:      ipairsOp,
		NRet:    lua.MultRet,
		Protect: false,
	}, arg); err != nil {
		raiseErrorE(L, err)
	}
	return 3 // ipairs's nret
}

func extendPairs(L *lua.LState) int {
	arg := L.CheckAny(1)
	// lua5.2 meta method extension
	pairsOp := L.GetMetaField(arg, "__pairs")

	if pairsOp == lua.LNil {
		if arg.Type() == lua.LTUserData {
			L.ArgError(1, "userdata has no __pairs meta method")
			return 0
		}
		pairsOp = L.CheckFunction(lua.UpvalueIndex(1)) // assumes raw pairs given
	}

	if err := L.CallByParam(lua.P{
		Fn:      pairsOp,
		NRet:    lua.MultRet,
		Protect: false,
	}, arg); err != nil {
		raiseErrorE(L, err)
	}
	return 3 // pairs's nret
}

// // iteration interface
type nextIntIterator interface {
	scalableValues
	Get(i int) int64
}

func checkNextIntIterator(L *lua.LState, pos int) nextIntIterator {
	ud := L.CheckUserData(pos)
	if value, ok := ud.Value.(nextIntIterator); ok {
		return value
	}
	L.ArgError(pos, "require a object having method Len() and Get()")
	return nil
}

func nextIntPair(ni nextIntIterator, idx int) (int, int64, bool) {
	idx += 1
	if indexIsInRange(idx, ni) {
		return idx, ni.Get(idx), true
	} else {
		return -1, 0, false
	}
}

func lnextIntPair(L *lua.LState) int {
	ns := checkNextIntIterator(L, 1)
	idx := L.OptInt(2, -1)
	nextIdx, value, ok := nextIntPair(ns, idx)
	if !ok {
		return 0
	} else {
		L.Push(lua.LNumber(nextIdx))
		L.Push(lua.LNumber(value))
		return 2
	}
}

type nextStrIterator interface {
	scalableValues
	Get(i int) string
}

func checkNextStrIterator(L *lua.LState, pos int) nextStrIterator {
	ud := L.CheckUserData(pos)
	if value, ok := ud.Value.(nextStrIterator); ok {
		return value
	}
	L.ArgError(pos, "require a object having method Len() and Get()")
	return nil
}

func nextStrPair(ns nextStrIterator, idx int) (int, string, bool) {
	idx += 1
	if indexIsInRange(idx, ns) {
		return idx, ns.Get(idx), true
	} else {
		return -1, "", false
	}
}

func lnextStrPair(L *lua.LState) int {
	ns := checkNextStrIterator(L, 1)
	idx := L.OptInt(2, -1)
	nextIdx, value, ok := nextStrPair(ns, idx)
	if !ok {
		return 0
	} else {
		L.Push(lua.LNumber(nextIdx))
		L.Push(lua.LString(value))
		return 2
	}
}

func lpairsWithMetaNext(L *lua.LState) int {
	ud := L.CheckUserData(1)
	nextOp := L.GetMetaField(ud, "__next")
	if nextOp == lua.LNil {
		L.ArgError(1, "__next meta function is not found")
	}
	L.Push(nextOp)
	L.Push(ud)
	L.Push(lua.LNumber(-1)) // to adjust to 0 on nextOp
	return 3
}

// extend pcall and xpcall to raise un-recoverable errors without any hooks in script layer.
func registerExtPCall(L *lua.LState, ip *Interpreter) {
	// assumes LState already registers "pcall" fucntion
	for _, v := range []struct {
		RawFuncName string
		ExtendFunc  lua.LGFunction
	}{
		{"pcall", ip.extendPcall},
		{"xpcall", ip.extendXPcall},
	} {
		rawFunc := L.GetGlobal(v.RawFuncName)
		if rawFunc == lua.LNil {
			panic(v.RawFuncName + " is not found in LState")
		}
		lextendFunc := L.NewClosure(v.ExtendFunc, rawFunc)
		L.SetGlobal(v.RawFuncName, lextendFunc)
	}
}

func (ip *Interpreter) extendPcall(L *lua.LState) int {
	rawFunc := L.CheckFunction(lua.UpvalueIndex(1)) // assumes raw pcall given
	args := []lua.LValue{}
	for i := 1; i <= L.GetTop(); i++ {
		args = append(args, L.Get(i))
	}
	orgTop := L.GetTop()

	err := L.CallByParam(lua.P{
		Fn:      rawFunc,
		NRet:    lua.MultRet,
		Protect: false,
	}, args...)
	raiseErrorIf(L, err)

	// clear internal error which is catched by Pcall protection, not used anywhere.
	_ = getAndClearRaisedError(L)

	// re-throw special errors so that script ignores Non-local exits by runtime.
	if ok := L.OptBool(orgTop+1, true); !ok {
		msg := L.OptString(orgTop+2, "nothing")
		err := ip.extractScriptInterruptError(msg)
		if err != nil {
			L.Pop(2)                // to reset rawFunc returns
			L.RaiseError("%s", msg) // re-throw msg itself to keep special error context in message.
			return 0
		}
	}
	return L.GetTop() - orgTop
}

func (ip *Interpreter) extendXPcall(L *lua.LState) int {
	rawFunc := L.CheckFunction(lua.UpvalueIndex(1)) // assumes raw xpcall given
	fn := L.CheckFunction(1)
	errHandler := L.OptFunction(2, nil)
	if errHandler != nil {
		errHandler = ip.wrapXPcallErrorHandler(L, errHandler)
	}
	orgTop := L.GetTop()

	err := L.CallByParam(lua.P{
		Fn:      rawFunc, // xpcall
		NRet:    lua.MultRet,
		Protect: false,
	}, fn, errHandler)
	raiseErrorIf(L, err) // it should be something fatal which breaks pcall protection,

	// clear internal error which is catched by Pcall protection, not used anywhere.
	_ = getAndClearRaisedError(L)

	// re-throw special errors such as gotoNextScene so that user can not handle any errors which
	// used by implementation internally. In other case, error handler is used as original XMCall does
	if ok := L.ToBool(orgTop + 1); !ok {
		msg := L.OptString(orgTop+2, "nothing")
		if err := ip.extractScriptInterruptError(msg); err != nil {
			L.RaiseError("%s", msg) // re-throw msg itself to keep special error context in message.
			return 0
		}
	}
	return L.GetTop() - orgTop
}

func (ip *Interpreter) wrapXPcallErrorHandler(L *lua.LState, orgHandler *lua.LFunction) *lua.LFunction {
	return L.NewFunction(func(L *lua.LState) int {
		// error in error handler breaks Lua runtime. must avoid any error in this function scope.
		// See. https://github.com/yuin/gopher-lua/issues/452
		errObj := L.CheckAny(1)
		msg := lua.LVAsString(errObj)
		if err := ip.extractScriptInterruptError(msg); err != nil {
			// reuse error msg itself to keep error context and immediately return
			// so that user can not access interruption raised by erago runtime.
			L.Push(lua.LString(msg))
			return 1
		} else {
			top := L.GetTop()
			L.Push(orgHandler)
			L.Push(errObj)
			if err := L.PCall(1, lua.MultRet, nil); err != nil {
				if apiErr, ok := err.(*lua.ApiError); ok {
					msg = lua.LVAsString(apiErr.Object)
				} else {
					// This case should not occured with PCall contract.
					msg = err.Error()
				}
				L.Push(lua.LString(msg))
				return 1
			}
			// need to return exactly 1 value, by Lua runtime restriction.
			if nret := L.GetTop() - top; nret > 1 {
				L.Pop(nret - 1)
			} else if nret == 1 {
				// do nothing
			} else {
				L.Push(lua.LNil)
			}
			return 1
		}
	})
}
