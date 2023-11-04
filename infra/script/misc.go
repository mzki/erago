package script

import (
	"math"

	"github.com/mzki/erago/scene"
	lua "github.com/yuin/gopher-lua"
)

const (
	// Maximum integer number to maintain accuracy in Interpreter.
	// Since log10(MaxInteger) = 15.95...,
	// a number upto 10^15 is completely supported, and
	// upto 10^16 is partialy supported.
	MaxInteger = (1 << 53) - 1

	// Maximum number in Interpreter.
	MaxNumber = math.MaxFloat64
)

// +gendoc.set_section "Era Module"

// +gendoc
// * var era.MAX_INTEGER: integer
// 精度を保てる最大の整数

// +gendoc
// * var era.MAX_NUMBER: number
// 最大の実数

// +gendoc
// * var era.PRINTC_WIDTH: integer
// era.printc()のデフォルトで使用されるwidthの値

// +gendoc
// * var era.TEXTBAR_WIDTH: integer
// era.printBar()のデフォルトで使用されるwidth, fg, bgの値

// +gendoc
// * var era.TEXTBAR_FG
// era.printBar()のデフォルトで使用されるwidth, fg, bgの値

// +gendoc
// * var era.TEXTBAR_BG
// era.printBar()のデフォルトで使用されるwidth, fg, bgの値

// +gendoc
// * var era.TEXTLINE_SYMBOL
// era.printLine()のデフォルトで使用されるsymbolの値

func registerMisc(L *lua.LState) {
	eraMod := mustGetEraModule(L)
	for _, s := range []struct {
		Key   string
		Value lua.LValue
	}{
		{"MAX_INTEGER", lua.LNumber(MaxInteger)},
		{"MAX_NUMBER", lua.LNumber(MaxNumber)},

		{"PRINTC_WIDTH", lua.LNumber(scene.DefaultPrintCWidth)},
		{"TEXTBAR_WIDTH", lua.LNumber(scene.DefaultTextBarWidth)},
		{"TEXTBAR_FG", lua.LString(scene.DefaultTextBarFg)},
		{"TEXTBAR_BG", lua.LString(scene.DefaultTextBarBg)},
		{"TEXTLINE_SYMBOL", lua.LString(scene.DefaultLineSymbol)},
	} {
		L.SetGlobal(s.Key, s.Value) // for backward compatibility
		eraMod.RawSetString(s.Key, s.Value)
	}
}
