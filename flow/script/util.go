package script

import "github.com/yuin/gopher-lua"

// // UserData and MetaTable utils

// new MetaTable with name and field values.
func newMetatable(L *lua.LState, mt_name string, fields map[string]lua.LValue) *lua.LTable {
	mt := L.NewTypeMetatable(mt_name)
	for key, lv := range fields {
		mt.RawSetString(key, lv)
	}
	return mt
}

// return registered index and newindex metatable.
func newGetterSetterMt(L *lua.LState, mt_name string, fn lua.LGFunction) *lua.LTable {
	lfn := L.NewFunction(fn)
	return newMetatable(L, mt_name, map[string]lua.LValue{
		"__index":     lfn,
		"__newindex":  lfn,
		"__metatable": metaProtectObj,
	})
}

// allocate User Data set metatable and data.
func newUserDataWithMt(L *lua.LState, data interface{}, mt lua.LValue) *lua.LUserData {
	ud := newUserData(L, data)
	L.SetMetatable(ud, mt)
	return ud
}

// allocate User Data set data.
func newUserData(L *lua.LState, v interface{}) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = v
	return ud
}

// must get era module in global scope. if not found panic.
func mustGetEraModule(L *lua.LState) *lua.LTable {
	era_module, ok := L.GetGlobal(EraModuleName).(*lua.LTable)
	if !ok {
		panic(EraModuleName + " is not found")
	}
	return era_module
}

// strict table occurs error when accessing undefined keys.
const strictTableMetaName = "stricttable"

// not use it directly, getting it by use getStrictTableMetaTable.
var strictTableMetatable *lua.LTable

func getStrictTableMetatable(L *lua.LState) *lua.LTable {
	if strictTableMetatable != nil {
		return strictTableMetatable
	}

	indexFunc := L.NewFunction(ltableNotFoundMetaIndex)
	strictTableMetatable = newMetatable(L, strictTableMetaName, map[string]lua.LValue{
		"__index": indexFunc,
		// __newindex is not handled
		"__metatable": metaProtectObj,
	})
	return strictTableMetatable
}

func ltableNotFoundMetaIndex(L *lua.LState) int {
	_ = L.CheckTable(1)
	key := L.CheckString(2)
	L.RaiseError(key + " is not found")
	return 0
}
