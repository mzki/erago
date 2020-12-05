package script

import (
	"fmt"

	"github.com/mzki/erago/state"

	lua "github.com/yuin/gopher-lua"
)

// // register System Paramters

const (
	systemParamsModuleName = "system"
	sharedParamsModuleName = "share"

	saveInfoDataName = "saveinfo"

	intParamMetaName = "IntParam"
	strParamMetaName = "StrParam"
)

func registerSystemParams(L *lua.LState, gamestate *state.GameState) {
	era_module := mustGetEraModule(L)

	len_func := L.NewFunction(lenScalable)
	next_int_func := L.NewFunction(lnextIntPair)
	next_str_func := L.NewFunction(lnextStrPair)
	pairs_func := L.NewFunction(lintstrIteratorMetaPairs)
	intparam_meta := getOrNewMetatable(L, intParamMetaName, map[string]lua.LValue{
		"__index":     L.NewFunction(intParamMetaIndex),
		"__newindex":  L.NewFunction(intParamMetaNewIndex),
		"__len":       len_func,
		"__ipairs":    pairs_func,
		"__pairs":     pairs_func,
		"__next":      next_int_func,
		"__metatable": metaProtectObj,
	})
	L.SetFuncs(intparam_meta, intParamMethods)
	L.SetGlobal(intParamMetaName, intparam_meta)

	strparam_meta := getOrNewMetatable(L, strParamMetaName, map[string]lua.LValue{
		"__index":     L.NewFunction(strParamMetaIndex),
		"__newindex":  L.NewFunction(strParamMetaNewIndex),
		"__len":       len_func,
		"__ipairs":    pairs_func,
		"__pairs":     pairs_func,
		"__next":      next_str_func,
		"__metatable": metaProtectObj,
	})
	L.SetFuncs(strparam_meta, strParamMethods)
	L.SetGlobal(strParamMetaName, strparam_meta)

	// register system and shared params as userdata to module.
	for _, data := range []struct {
		modName string
		iface   interface {
			ForEachIntParam(func(key string, param state.IntParam))
			ForEachStrParam(func(key string, param state.StrParam))
		}
	}{
		{systemParamsModuleName, gamestate.SystemData},
		{sharedParamsModuleName, gamestate.ShareData},
	} {
		mod := L.NewTable()
		data.iface.ForEachIntParam(func(key string, param state.IntParam) {
			ud := newUserDataWithMt(L, param, intparam_meta)
			mod.RawSetString(key, ud)
		})
		data.iface.ForEachStrParam(func(key string, param state.StrParam) {
			ud := newUserDataWithMt(L, param, strparam_meta)
			mod.RawSetString(key, ud)
		})
		L.SetMetatable(mod, getStrictTableMetatable(L))
		era_module.RawSetString(data.modName, mod)
	}

	// register save informations
	ud := newUserDataWithMt(L, gamestate.SaveInfo, newGetterSetterMt(L, saveInfoDataName, getSetSaveInfo))
	era_module.RawSetString(saveInfoDataName, ud)
}

// save info user data
func getSetSaveInfo(L *lua.LState) int {
	ud := L.CheckUserData(1)
	info, ok := ud.Value.(*state.SaveInfo)
	if !ok {
		L.ArgError(1, "require Save info object")
		return 0
	}

	key := L.CheckString(2)
	if L.GetTop() == 3 {
		switch key {
		case "save_comment":
			info.SaveComment = L.CheckString(3)
		default:
			L.ArgError(2, "saveInfo can not be assigned field "+key)
		}
		return 0
	}
	switch key {
	case "save_comment":
		L.Push(lua.LString(info.SaveComment))
	case "load_comment":
		L.Push(lua.LString(info.LastLoadComment))
	case "load_ver":
		L.Push(lua.LNumber(info.LastLoadVer))
	default:
		L.ArgError(2, key+" is not found in saveInfo")
	}
	return 1
}

// // scalableValues

type scalableValues interface {
	Len() int // return its size
}

func checkScalableValues(L *lua.LState, pos int) scalableValues {
	ud := L.CheckUserData(pos)
	if value, ok := ud.Value.(scalableValues); ok {
		return value
	}
	L.ArgError(pos, "require a object having method Len()")
	return nil
}

func lenScalable(L *lua.LState) int {
	data := checkScalableValues(L, 1)
	L.Push(lua.LNumber(data.Len()))
	return 1
}

// return whether given index is in range of scalableValues.
func indexIsInRange(index int, sv scalableValues) bool {
	return index >= 0 && index < sv.Len()
}

// // XXXParam

func checkNameIndexer(L *lua.LState, pos int) state.NameIndexer {
	ud := L.CheckUserData(pos)
	if indexer, ok := ud.Value.(state.NameIndexer); ok {
		return indexer
	}
	L.ArgError(pos, "require csvindex.* object")
	return nil
}

// check pos-th argument is int or string, then return as index,
func checkParamIndex(L *lua.LState, pos int, indexer state.NameIndexer) int {
	L.CheckTypes(pos, lua.LTNumber, lua.LTString)

	switch lval := L.Get(pos); lval.Type() {
	case lua.LTNumber:
		return int(lua.LVAsNumber(lval))
	case lua.LTString:
		key := lua.LVAsString(lval)
		index := indexer.GetIndex(key)
		if index < 0 {
			L.ArgError(pos, key+" is not found in csv")
		}
		return index
	}
	return -1
}

// return slicing index range, [from:to).
func checkIndexSliceRange(L *lua.LState, pos int, sv scalableValues) (from, to int) {
	from = L.CheckInt(pos)
	to = L.OptInt(pos+1, sv.Len())

	starting_negative := from < 0
	inversed_range := from > to
	ending_overrun := to > sv.Len()
	if starting_negative || inversed_range || ending_overrun {
		L.ArgError(pos, fmt.Sprintf("slice range: from(%d) ~ to(%d), must be in 0 ~ data-length(%d)", from, to, sv.Len()))
	}
	return from, to
}

// //  intParam

var intParamMethods = map[string]lua.LGFunction{
	"new":   intParamNew,
	"set":   intParamGetSet,
	"get":   intParamGetSet,
	"len":   lenScalable,
	"slice": intParamSlice,
	"fill":  intParamFill,
	// TODO: implement "ipairs", "pairs"
}

// construct intparam as lua object.
func newLIntParam(L *lua.LState, ip state.IntParam) *lua.LUserData {
	return newUserDataWithMt(L, ip, L.GetTypeMetatable(intParamMetaName))
}

// check whether pos-th argument is intParams as userdata?
func checkIntParam(L *lua.LState, pos int) state.IntParam {
	ud := L.CheckUserData(pos)
	if value, ok := ud.Value.(state.IntParam); ok {
		return value
	}
	L.ArgError(pos, "require IntParam object")
	return state.IntParam{}
}

const indexOutMessage = "index out of data range"

// metamethod __index for IntParam.
func intParamMetaIndex(L *lua.LState) int {
	data := checkIntParam(L, 1)
	L.CheckTypes(2, lua.LTNumber, lua.LTString)

	switch lvalue := L.Get(2); lvalue.Type() {
	case lua.LTNumber:
		index := int(lua.LVAsNumber(lvalue))
		if ok := indexIsInRange(index, data); !ok {
			L.ArgError(2, indexOutMessage)
		}
		L.Push(lua.LNumber(data.Get(index)))
		return 1
	case lua.LTString:
		key := lua.LVAsString(lvalue)
		// find methods
		mt := L.GetTypeMetatable(intParamMetaName).(*lua.LTable)
		if fn := mt.RawGetString(key); fn.Type() == lua.LTFunction {
			L.Push(fn)
			return 1
		}

		// find data
		if val, ok := data.GetByStr(key); ok {
			L.Push(lua.LNumber(val))
			return 1
		}
		L.ArgError(2, key+" is not found in csv")
	}

	L.Push(lua.LNil)
	return 1
}

// metamethod __newindex for IntParam.
func intParamMetaNewIndex(L *lua.LState) int {
	data := checkIntParam(L, 1)
	index := checkParamIndex(L, 2, data)
	if ok := indexIsInRange(index, data); !ok {
		L.ArgError(2, indexOutMessage)
	}
	new_value := L.CheckInt64(3)
	data.Set(index, new_value)
	return 0
}

// +gendoc "XXXParam"
// * new_intparam = IntParam.new(size, [name_indexer])
//
// 新しいIntParamを作成します。IntParamはsizeの長さのInt64配列と同じように振る舞います。
// Indexは0から始まることに注意が必要です。
// name_indexerは省略できます。name_indexerは、文字列を渡すとindex番号を返すオブジェクトです。
// 具体的には、csvindexモジュール以下のオブジェクトです。
// name_indexerを渡した場合には、indexとして文字列も使用することができます。
// つまり、
//   intparam = IntParam.new(100)
//   index = csvindex.item["道具"]
//   value = intparam[index]
//
// という操作を、
//   intparam = IntParam.new(100, csvindex.item)
//   value = intparam["道具"]
//
// というように、代行してくれます。
func intParamNew(L *lua.LState) int {
	size := L.CheckInt(1)
	var indexer state.NameIndexer
	if L.GetTop() == 2 {
		indexer = checkNameIndexer(L, 2)
	} else {
		indexer = state.NoneNameIndexer{}
	}
	ip := state.NewIntParam(make([]int64, size), indexer)
	L.Push(newLIntParam(L, ip))
	return 1
}

// +gendoc "XXXParam"
// * value = IntParam:get(key),  IntParam:set(key, new_value)
//
// key番目の値を取得/設定します。keyにはインデックス番号あるいは文字列を指定します。
// 例えば、keyの名前とIntParamのメソッド名が被っている場合には、メソッドが優先されてしまいます。
//   intparam:len()
//   len_method = intparam["len"]
//   not_a_value = intparam["len"]
//
// その場合（ここでは、"len"というkey）でも、このget/setによって、値の取得/設定が可能です。
func intParamGetSet(L *lua.LState) int {
	ip := checkIntParam(L, 1)
	index := checkParamIndex(L, 2, ip)
	if ok := indexIsInRange(index, ip); !ok {
		L.ArgError(2, indexOutMessage)
	}

	if L.GetTop() == 3 {
		// set
		ip.Set(index, L.CheckInt64(3))
		return 0
	}
	// get
	L.Push(lua.LNumber(ip.Get(index)))
	return 1
}

// +gendoc "XXXParam"
// * sliced_intparam = IntParam:slice(from,[to = max_length])
//
// fromからtoまでのデータ範囲を切り出します。切り出したデータは再び0から始まり、その長さは(to - from)になります。
// toは省略可能です。省略したときには、現在のデータの最大の長さがtoとして使用されます。
func intParamSlice(L *lua.LState) int {
	ip := checkIntParam(L, 1)
	from, to := checkIndexSliceRange(L, 2, ip)
	L.Push(newLIntParam(L, ip.Slice(from, to)))
	return 1
}

// +gendoc "XXXParam"
// * IntParam:fill(new_value)
//
// 現在のデータ範囲すべてをnew_valueで初期化します。
func intParamFill(L *lua.LState) int {
	ip := checkIntParam(L, 1)
	val := L.CheckInt64(2)
	ip.Fill(val)
	return 0
}

// // strParan

var strParamMethods = map[string]lua.LGFunction{
	"new":   strParamNew,
	"set":   strParamGetSet,
	"get":   strParamGetSet,
	"len":   lenScalable,
	"slice": strParamSlice,
	"fill":  strParamFill,
	// TODO: implement "ipairs", "pairs"
}

// construct strparam as lua object.
func newLStrParam(L *lua.LState, sp state.StrParam) *lua.LUserData {
	return newUserDataWithMt(L, sp, L.GetTypeMetatable(strParamMetaName))
}

func checkStrParam(L *lua.LState, pos int) state.StrParam {
	ud := L.CheckUserData(pos)
	if value, ok := ud.Value.(state.StrParam); ok {
		return value
	}
	L.ArgError(pos, "require StrParam object")
	return state.StrParam{}
}

// metamethod __index for StrParam.
func strParamMetaIndex(L *lua.LState) int {
	data := checkStrParam(L, 1)
	L.CheckTypes(2, lua.LTNumber, lua.LTString)

	switch lvalue := L.Get(2); lvalue.Type() {
	case lua.LTNumber:
		index := int(lua.LVAsNumber(lvalue))
		if ok := indexIsInRange(index, data); !ok {
			L.ArgError(2, indexOutMessage)
		}
		L.Push(lua.LString(data.Get(index)))
		return 1
	case lua.LTString:
		key := lua.LVAsString(lvalue)
		// find methods
		mt := L.GetTypeMetatable(strParamMetaName).(*lua.LTable)
		if fn := mt.RawGetString(key); fn.Type() == lua.LTFunction {
			L.Push(fn)
			return 1
		}

		// find data
		if val, ok := data.GetByStr(key); ok {
			L.Push(lua.LString(val))
			return 1
		}
		L.ArgError(2, key+" is not found in csv")
	}

	L.Push(lua.LNil)
	return 1
}

// metamethod __newindex for StrParam.
func strParamMetaNewIndex(L *lua.LState) int {
	data := checkStrParam(L, 1)
	index := checkParamIndex(L, 2, data)
	if ok := indexIsInRange(index, data); !ok {
		L.ArgError(2, indexOutMessage)
	}
	new_value := L.CheckString(3)
	data.Set(index, new_value)
	return 0
}

// +gendoc "XXXParam"
// * new_strparam = StrParam.new(size, [name_indexer])
//
// 新しいStrParamを作成します。StrParamはsizeの長さの文字列の配列と同じように振る舞います。
// 詳細はIntParamに同じ。
func strParamNew(L *lua.LState) int {
	size := L.CheckInt(1)
	var indexer state.NameIndexer
	if L.GetTop() == 2 {
		indexer = checkNameIndexer(L, 2)
	} else {
		indexer = state.NoneNameIndexer{}
	}
	ip := state.NewStrParam(make([]string, size), indexer)
	L.Push(newLStrParam(L, ip))
	return 1
}

// +gendoc "XXXParam"
// * value = StrParam:get(key),  StrParam:set(key, new_value)
//
// key番目の値を取得/設定します。keyにはインデックス番号あるいは文字列を指定します。
func strParamGetSet(L *lua.LState) int {
	sp := checkStrParam(L, 1)
	index := checkParamIndex(L, 2, sp)
	if ok := indexIsInRange(index, sp); !ok {
		L.ArgError(2, indexOutMessage)
	}

	if L.GetTop() == 3 {
		// set
		sp.Set(index, L.CheckString(3))
		return 0
	}
	// get
	L.Push(lua.LString(sp.Get(index)))
	return 1
}

// +gendoc "XXXParam"
// * sliced_strparam = StrParam:slice(from,[to = max_length])
//
// fromからtoまでのデータ範囲を切り出します。切り出したデータは再び0から始まり、その長さは(from - to)になります。
// toは省略可能です。省略したときには、現在のデータの最大の長さがtoとして使用されます。
func strParamSlice(L *lua.LState) int {
	sp := checkStrParam(L, 1)
	from, to := checkIndexSliceRange(L, 2, sp)
	L.Push(newLStrParam(L, sp.Slice(from, to)))
	return 1
}

// +gendoc "XXXParam"
// * StrParam:fill(new_value)
//
// 現在のデータ範囲すべてをnew_valueで初期化します。
func strParamFill(L *lua.LState) int {
	ip := checkStrParam(L, 1)
	val := L.CheckString(2)
	ip.Fill(val)
	return 0
}
