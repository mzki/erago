package script

import (
	"local/erago/flow"
	"math"

	"github.com/yuin/gopher-lua"
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

// +gendoc.set_section "Constant Value"

// +gendoc
// * MAX_INTEGER
//
// 精度を保てる最大の整数
//
//
// * MAX_NUMBER
//
// 最大の実数
//
//
// * PRINTC_WIDTH
//
// era.printc()のデフォルトで使用されるwidthの値
//
//
// * TEXTBAR_WIDTH
// * TEXTBAR_FG
// * TEXTBAR_BG
//
// era.printBar()のデフォルトで使用されるwidth, fg, bgの値
//
//
// * TEXTLINE_SYMBOL
//
// era.printLine()のデフォルトで使用されるsymbolの値
func registerMisc(L *lua.LState) {
	L.SetGlobal("MAX_INTEGER", lua.LNumber(MaxInteger))
	L.SetGlobal("MAX_NUMBER", lua.LNumber(MaxNumber))

	L.SetGlobal("PRINTC_WIDTH", lua.LNumber(flow.DefaultPrintCWidth))
	L.SetGlobal("TEXTBAR_WIDTH", lua.LNumber(flow.DefaultTextBarWidth))
	L.SetGlobal("TEXTBAR_FG", lua.LString(flow.DefaultTextBarFg))
	L.SetGlobal("TEXTBAR_BG", lua.LString(flow.DefaultTextBarBg))
	L.SetGlobal("TEXTLINE_SYMBOL", lua.LString(flow.DefaultLineSymbol))
}
