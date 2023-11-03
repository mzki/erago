package script

import lua "github.com/yuin/gopher-lua"

// // UserData and MetaTable utils

// new MetaTable with name and field values.
func newMetatable(L *lua.LState, mt_name string, fields map[string]lua.LValue) *lua.LTable {
	mt := L.NewTypeMetatable(mt_name)
	for key, lv := range fields {
		mt.RawSetString(key, lv)
	}
	return mt
}

// get Metatable with mtName if exist and return it,
// if not exist create new Metatable with name and field values.
func getOrNewMetatable(L *lua.LState, mtName string, fields map[string]lua.LValue) *lua.LTable {
	if mt, ok := L.GetTypeMetatable(mtName).(*lua.LTable); ok {
		return mt
	}
	return newMetatable(L, mtName, fields)
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
	raiseErrorf(L, key+" is not found")
	return 0
}

// ----------- Lua Registery utils --------------------

func setRegGValue[T any](L *lua.LState, key string, value T) {
	ud := L.NewUserData()
	ud.Value = value
	setRegValue(L, key, ud)
}

func getRegGValue[T any](L *lua.LState, key string) T {
	ud := getRegValue(L, key).(*lua.LUserData)
	return ud.Value.(T)
}

func setRegValue(L *lua.LState, key string, value lua.LValue) {
	L.Get(lua.RegistryIndex).(*lua.LTable).RawSetString(key, value)
}

func getRegValue(L *lua.LState, key string) lua.LValue {
	return L.Get(lua.RegistryIndex).(*lua.LTable).RawGetString(key)
}
