package script

import (
	"github.com/mzki/erago/state/csv"
	lua "github.com/yuin/gopher-lua"
)

// +gendoc.set_section "Builtin Module: csv"

// +gendoc
// CSVファイルの読み込みを行うモジュール。

const builtinCSVModuleName = "csv"

// builtin csv module for lua interpreter.
func builtinCSVLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), builtinCSVExports)
	L.SetMetatable(mod, getStrictTableMetatable(L))
	L.Push(mod)
	return 1
}

var builtinCSVExports = map[string]lua.LGFunction{
	"readFunc": builtinCSVReadFunc,
}

// +gendoc
// * csv.readFunc(file_name, record_func)
//
// csvファイルであるfile_nameを読み込みます。csvファイルの各行は
// record_func(i, record) によって処理されます。ここで、
// i は現在の行数、record は現在の行のフィールド値の配列です。
//
// Example:
//
//	local filename = "/path/to/any.csv"
//	csv.readFunc(filename, function(i, record)
//	  local str1 = record[1] -- １列目の文字列
//	  local str2 = record[2] -- ２列目の文字列
//	end)
func builtinCSVReadFunc(L *lua.LState) int {
	file := checkFilePath(L, 1)
	luaP := lua.P{
		Fn:      L.CheckFunction(2),
		NRet:    0,
		Protect: false,
	}
	var i int = 0
	err := csv.ReadFileFunc(file, func(record []string) error {
		lt := L.CreateTable(len(record), 0)
		for _, r := range record {
			lt.Append(lua.LString(r))
		}
		i += 1
		return L.CallByParam(luaP, lua.LNumber(i), lt)
	})
	raiseErrorIf(L, err)
	return 0
}
