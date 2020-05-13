package script

import (
	"github.com/yuin/gopher-lua"
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
		L.RaiseError(err.Error())
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
		L.RaiseError(err.Error())
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

func lintstrIteratorMetaPairs(L *lua.LState) int {
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
