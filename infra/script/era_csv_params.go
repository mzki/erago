package script

import (
	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
	lua "github.com/yuin/gopher-lua"
)

const (
	csvModuleName       = "csv"
	csvFieldsModuleName = "csvfields"
	csvIndexModuleName  = "csvindex"

	csvFieldsNumbersMetaName = csvIndexModuleName + ".numbers"
	csvFieldsStringsMetaName = csvIndexModuleName + ".strings"

	csvItemPriceMetaName = "item_price"
)

var (
	// it is used as name in script.
	csvBuiltinItemName      = csv.BuiltinItemName
	csvBuiltinItemPriceName = csv.BuiltinItemPriceName
)

func registerCsvParams(L *lua.LState, CSV *csv.CsvManager) {
	era_module := mustGetEraModule(L)
	if v := era_module.RawGetString(csvModuleName); lua.LVAsBool(v) {
		return // already exist
	}

	LLenFunction := L.NewFunction(lenScalable)
	{ // register csv names
		csv_module := L.NewTable()
		L.SetMetatable(csv_module, getStrictTableMetatable(L))
		era_module.RawSetString(csvModuleName, csv_module)

		csv_names_meta := newMetatable(L, csvModuleName, map[string]lua.LValue{
			"__index":     L.NewFunction(csvNamesMetaIndex),
			"__len":       LLenFunction,
			"__metatable": metaProtectObj,
		})

		// register names defined by csv.
		for key, c := range CSV.Constants() {
			ud := newUserDataWithMt(L, c.Names, csv_names_meta)
			csv_module.RawSetString(key, ud)
		}

		// register builtin names.
		for key, names := range map[string]csv.Names{
			csvBuiltinItemName: CSV.Item.Names,
		} {
			ud := newUserDataWithMt(L, names, csv_names_meta)
			csv_module.RawSetString(key, ud)
		}

		// register builtin constants, csv item price
		int_param := state.NewIntParam(CSV.ItemPrices, CSV.Item.NameIndex)
		meta := newMetatable(L, csvItemPriceMetaName, map[string]lua.LValue{
			"__index":     L.NewFunction(intParamMetaIndex),
			"__len":       LLenFunction,
			"__metatable": metaProtectObj,
		})
		item_prices := newUserDataWithMt(L, int_param, meta)
		csv_module.RawSetString(csvBuiltinItemPriceName, item_prices)
	}

	{ // register csv index
		csv_index_module := L.NewTable()
		L.SetMetatable(csv_index_module, getStrictTableMetatable(L))
		era_module.RawSetString(csvIndexModuleName, csv_index_module)

		csv_index_meta := newMetatable(L, csvIndexModuleName, map[string]lua.LValue{
			"__index":     L.NewFunction(csvIndexMetaIndex),
			"__len":       LLenFunction,
			"__metatable": metaProtectObj,
		})

		// register csv index deifined by user.
		for key, c := range CSV.Constants() {
			ud := newUserDataWithMt(L, c.NameIndex, csv_index_meta)
			csv_index_module.RawSetString(key, ud)
		}

		// register builtin csv index.
		for key, nidx := range map[string]csv.NameIndex{
			csvBuiltinItemName: CSV.Item.NameIndex,
		} {
			ud := newUserDataWithMt(L, nidx, csv_index_meta)
			csv_index_module.RawSetString(key, ud)
		}
	}

	{ // register csv fields
		registerCsvFieldsMeta(L) // must be first

		csv_fields_module := L.NewTable()
		L.SetMetatable(csv_fields_module, getStrictTableMetatable(L))
		era_module.RawSetString(csvFieldsModuleName, csv_fields_module)

		// register csv index deifined by user.
		csv_fields_meta := newMetatable(L, csvFieldsModuleName, map[string]lua.LValue{
			"__index":     L.NewFunction(csvFieldsMetaIndex),
			"__len":       LLenFunction,
			"__metatable": metaProtectObj,
		})
		for key, c := range CSV.Constants() {
			ud := newUserDataWithMt(L, c.CustomFields, csv_fields_meta)
			csv_fields_module.RawSetString(key, ud)
		}
	}
}

// TODO use struct with debug information, varname.
// type struct LCSVName {
//   varname string, // for debug
//   csv.Names
// }
//
// type struct LCSVIndex {
//   varname string, // for debug
//   csv.NameIndex
// }

// // csv names

func checkCsvNames(L *lua.LState, pos int) csv.Names {
	ud := L.CheckUserData(pos)
	if names, ok := ud.Value.(csv.Names); ok {
		return names
	}
	L.ArgError(pos, "require csv.* object")
	return nil
}

func csvNamesMetaIndex(L *lua.LState) int {
	names := checkCsvNames(L, 1)
	idx := L.CheckInt(2)

	if ok := indexIsInRange(idx, names); !ok {
		L.ArgError(2, indexOutMessage)
	}
	L.Push(lua.LString(names.Get(idx)))
	return 1
}

// // csv index

func checkCsvIndex(L *lua.LState, pos int) csv.NameIndex {
	ud := L.CheckUserData(pos)
	if nidx, ok := ud.Value.(csv.NameIndex); ok {
		return nidx
	}
	L.ArgError(pos, "require csvindex.* object")
	return nil
}

func csvIndexMetaIndex(L *lua.LState) int {
	nidx := checkCsvIndex(L, 1)
	key := L.CheckString(2)

	index := nidx.GetIndex(key)
	if index < 0 {
		L.ArgError(2, key+" is not found")
	}
	L.Push(lua.LNumber(index))
	return 1
}

// // csv fields

func registerCsvFieldsMeta(L *lua.LState) {
	if lv := L.GetTypeMetatable(csvFieldsNumbersMetaName); lua.LVAsBool(lv) {
		return // already exists
	}

	LLenFunction := L.NewFunction(lenScalable)

	// Numbers
	_ = newMetatable(L, csvFieldsNumbersMetaName, map[string]lua.LValue{
		"__index":     L.NewFunction(csvNumbersMetaIndex),
		"__len":       LLenFunction,
		"__metatable": metaProtectObj,
	})
	// Strings
	_ = newMetatable(L, csvFieldsStringsMetaName, map[string]lua.LValue{
		"__index":     L.NewFunction(csvStringsMetaIndex),
		"__len":       LLenFunction,
		"__metatable": metaProtectObj,
	})
}

func checkCsvFields(L *lua.LState, pos int) csv.CustomFields {
	ud := L.CheckUserData(pos)
	if cf, ok := ud.Value.(csv.CustomFields); ok {
		return cf
	}
	L.ArgError(pos, "require csvfields.* object")
	return csv.CustomFields{}
}

func csvFieldsMetaIndex(L *lua.LState) int {
	cf := checkCsvFields(L, 1)
	key := L.CheckString(2)

	switch cf.TypeOf(key) {
	case csv.CFIntType:
		nums := cf.MustNumbers(key)
		ud := newUserDataWithMt(L, nums, L.GetTypeMetatable(csvFieldsNumbersMetaName))
		L.Push(ud)
		return 1
	case csv.CFStrType:
		strs := cf.MustStrings(key)
		ud := newUserDataWithMt(L, strs, L.GetTypeMetatable(csvFieldsStringsMetaName))
		L.Push(ud)
		return 1
	default:
		L.ArgError(2, key+" is not found in csv fields")
		return 0
	}
}

func checkCsvNumbers(L *lua.LState, pos int) *csv.Numbers {
	ud := L.CheckUserData(pos)
	if nums, ok := ud.Value.(*csv.Numbers); ok {
		return nums
	}
	L.ArgError(pos, "require csvfields.numbers object")
	return nil
}

func checkCsvStrings(L *lua.LState, pos int) *csv.Strings {
	ud := L.CheckUserData(pos)
	if strs, ok := ud.Value.(*csv.Strings); ok {
		return strs
	}
	L.ArgError(pos, "require csvfields.numbers object")
	return nil
}

func csvNumbersMetaIndex(L *lua.LState) int {
	nums := checkCsvNumbers(L, 1)
	i := L.CheckInt(2)
	if i < 0 || i >= nums.Len() {
		L.ArgError(2, "index out of bounds")
	}
	L.Push(lua.LNumber(nums.Get(i)))
	return 1
}

func csvStringsMetaIndex(L *lua.LState) int {
	strs := checkCsvStrings(L, 1)
	i := L.CheckInt(2)
	if i < 0 || i >= strs.Len() {
		L.ArgError(2, "index out of bounds")
	}
	L.Push(lua.LString(strs.Get(i)))
	return 1
}
