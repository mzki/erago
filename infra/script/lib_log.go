package script

import (
	"fmt"

	"github.com/mzki/erago/util/log"
	lua "github.com/yuin/gopher-lua"
)

// TODO: split logger between script and system?

// +gendoc.set_section "Builtin Module: log"

// +gendoc
// スクリプトの途中経過を出力するモジュール。
//
// 開発のデバッグ用に用意されたモジュールです。
// 変数の中身を知りたいとき、このモジュール
// の関数を通してログファイルに出力することで、
// 確認することができます。
//
// このモジュールでは、2つのロギングレベルが用意されています。
//
// 1. Informationレベル
//   このレベルでは、全てのメッセージが出力されます。
//   なぜなら、このレベルが扱う内容は誰にとっても
//   有益な情報(informative)であるべきだからです。
//   エラーメッセージや、スクリプトの流れを確認するためのメッセージ
//   などが、ここに含まれます。
//
// 2. Debugレベル
//   このレベルでは、システムがデバッグモードのときのみ出力されます。
//   ここでは、値の変化を細かく追っていくような、開発者による精査が
//   必要な場合を想定しています。
//
// 参考1: https://dave.cheney.net/2015/11/05/lets-talk-about-logging
// 参考2: http://qiita.com/methane/items/cedbf546ae2db8a63c3d
//
//
// Example:
//   log = require "log"
//   log.info("1+1は", 1+1, "です") -- "1+1は 2 です"とログファイルに出力
//   log.infof("1+1は%dです", 1+1) -- "1+1は2です"とログファイルに出力
//

const loggerModuleName = "log"

func loggerLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), loggerExports)
	L.Push(mod)
	return 1
}

var loggerExports = map[string]lua.LGFunction{
	"debug":  logDebug,
	"debugf": LogDebugf,
	"info":   logInfo,
	"infof":  logInfof,
}

// push value v into head of values vs.
func unshiftDebugValues(vs []interface{}, v interface{}) []interface{} {
	if len(vs) == 0 {
		return []interface{}{v}
	}
	vs = append(vs[:1], vs[0:]...)
	vs[0] = v
	return vs
}

// If Registry table has registryDebugEnableKey and stored value is true
// then logDebugXXX funtions do, otherwise do nothing.
func debugEnable(L *lua.LState) bool {
	lv := L.CheckTable(lua.RegistryIndex).RawGetString(registryDebugEnableKey)
	return lua.LVAsBool(lv)
}

// +gendoc
// * log.debug(any...)
//
// Debugレベルで、要素anyをログファイルに出力します。
// 出力結果は、各要素の間に１文字ぶんの空白をいれて、
// １行で出力されます。
// この関数はデバッグモード時のみ動作します。
func logDebug(L *lua.LState) int {
	if !debugEnable(L) {
		return 0
	}
	vs := getDebugValues(L, 1)
	header := getDebugHeader(L)
	vs = unshiftDebugValues(vs, header)
	log.Debugln(vs...)
	return 0
}

// +gendoc
// * log.debugf(fmt_string, [any...])
//
// Debugレベルで、fmt_stringをログファイルに出力します。
// fmt_string中の特殊文字(例えば、%v)は、anyで置き換えられます。
// 特殊文字の詳細は、Golangのfmtパッケージを参照 (https://godoc.org/fmt)。
// この関数はデバッグモード時のみ動作します。
func LogDebugf(L *lua.LState) int {
	if !debugEnable(L) {
		return 0
	}
	fmt_str := L.CheckString(1)
	vs := getDebugValues(L, 2)
	header := getDebugHeader(L)
	vs = unshiftDebugValues(vs, header)
	log.Debugf("%s "+fmt_str, vs...)
	return 0
}

func getDebugValues(L *lua.LState, start int) []interface{} {
	n := L.GetTop()
	if start < 1 || start > n {
		return []interface{}{}
	}

	// 1 free space to set header at later.
	vs := make([]interface{}, 0, n+1)
	for i := start; i <= n; i++ {
		var msg string
		switch lv := L.Get(i).(type) {
		case lua.LNumber, lua.LBool, lua.LString:
			msg = lv.String()
		case *lua.LFunction:
			p := lv.Proto
			msg = fmt.Sprintf("function: %s:%d", p.SourceName, p.LineDefined)
		case *lua.LTable:
			msg = fmt.Sprintf("table: size %d", lv.Len())
		case *lua.LUserData:
			msg = fmt.Sprintf("userdata: %v", lv.Value)
		case nil:
			msg = "nil"
		default:
			msg = lv.String()
		}
		vs = append(vs, msg)
	}
	return vs
}

func getDebugHeader(L *lua.LState) string {
	dbg, ok := L.GetStack(1)
	if !ok {
		raiseErrorf(L, "getDebugHeader: can not get current call stack.")
	}
	_, err := L.GetInfo("Sln", dbg, lua.LNil)
	if err != nil {
		raiseErrorf(L, "getDebugHeader: %w", err)
	}
	header := fmt.Sprintf("Script: %s:%d:", dbg.Source, dbg.CurrentLine)
	return header
}

// +gendoc
// * log.info(any...)
//
// Informationレベルで、要素anyをログファイルに出力します。
// 各要素anyを文字列化したのち、空白１つで１行にしたものを出力します。
func logInfo(L *lua.LState) int {
	vs := getDebugValues(L, 1)
	vs = unshiftDebugValues(vs, "Script:")
	log.Infoln(vs...)
	return 0
}

// +gendoc
// * log.infof(fmt_string, any...)
//
// Informationレベルで、fmt_stringをログファイルに出力します。
// fmt_string中の特殊文字(例えば、%v)は、anyで置き換えられます。
// 特殊文字の詳細は、Golangのfmtパッケージを参照 (https://godoc.org/fmt)。
func logInfof(L *lua.LState) int {
	fmt_str := L.CheckString(1)
	vs := getDebugValues(L, 2)
	vs = unshiftDebugValues(vs, "Script:")
	log.Infof("%s "+fmt_str, vs...)
	return 0
}
