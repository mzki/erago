package script

import (
	"time"

	"github.com/yuin/gopher-lua"
)

// +gendoc.set_section "Builtin Module: time"

// +gendoc
// 時間に関係する処理を行うモジュール。
//
// このモジュールでは、数値をナノ秒と捉えます。
// 以下の定数との掛け算によって、数値の単位を変換することが
// できます。
//
// * time.NANOSECOND  =  1
// * time.MICROSECOND = time.NANOSECOND * 1000
// * time.MILLISECOND = time.MICROSECOND * 1000
// * time.SECOND      = time.MILLISECOND * 1000
//
// Example:
//   time = require "time"
//   local one_nsec = 1 * time.NANOSECOND -- １ナノ秒
//   local one_usec = 1 * time.MICROSECOND -- １マイクロ秒
//   local one_msec = 1 * time.MILLISECOND -- １ミリ秒
//   local one_sec = 1 * time.SECOND -- １秒

const timeModuleName = "time"

// time module for lua interpreter.
func timeLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), timeExports)
	mod.RawSetString("SECOND", lua.LNumber(time.Second))
	mod.RawSetString("MILLISECOND", lua.LNumber(time.Millisecond))
	mod.RawSetString("MICROSECOND", lua.LNumber(time.Microsecond))
	mod.RawSetString("NANOSECOND", lua.LNumber(time.Nanosecond))
	L.SetMetatable(mod, getStrictTableMetatable(L))
	L.Push(mod)
	return 1
}

var timeExports = map[string]lua.LGFunction{
	"now":      timeNow,
	"since":    timeSince,
	"year":     timeYear,
	"month":    timeMonth,
	"day":      timeDay,
	"hour":     timeHour,
	"minute":   timeMinute,
	"second":   timeSecond,
	"weekday":  timeWeekday,
	"format":   timeFormat,
	"tostring": timeToString,
}

func checkDuration(L *lua.LState, pos int) time.Duration {
	return time.Duration(L.CheckInt64(pos))
}

func checkTime(L *lua.LState, pos int) time.Time {
	ud := L.CheckUserData(pos)
	t, ok := ud.Value.(time.Time)
	if !ok {
		L.ArgError(pos, "require time userdata")
	}
	return t
}

// +gendoc
// * t = time.now(["*t"])
//
// 現在時刻を取得します。返り値tは通常timeモジュールでしか
// 扱うことはできません。しかし、この関数の引数として、
// "*t"を渡すと、返り値はテーブル型で返され、自由に
// 中身を見ることができます。ただし、テーブル型にした場合、
// timeモジュールで扱うことはできないことに注意してください。
//
// テーブル型の中身は以下の通り：
//   year:  年,
//   month: 月(1-12),
//   day:   日(1-31),
//   hour:  時(0-23),
//   min:   分(0-59),
//   sec:   秒(0-59),
//   wday:  曜日(0-6), -- 日曜を0として加算

// return time object or time as table if "*t" is passed.
func timeNow(L *lua.LState) int {
	if L.GetTop() == 0 {
		ud := L.NewUserData()
		ud.Value = time.Now()
		L.Push(ud)
		return 1
	}

	// unpack fields to lua.Table
	if fmt := L.CheckString(1); fmt != "*t" {
		L.ArgError(1, "invalid argument: "+fmt)
	}
	// references gopher-lua: oslib.go: osData()
	t := time.Now()
	ret := L.NewTable()
	ret.RawSetString("year", lua.LNumber(t.Year()))
	ret.RawSetString("month", lua.LNumber(t.Month()))
	ret.RawSetString("day", lua.LNumber(t.Day()))
	ret.RawSetString("hour", lua.LNumber(t.Hour()))
	ret.RawSetString("min", lua.LNumber(t.Minute()))
	ret.RawSetString("sec", lua.LNumber(t.Second()))
	ret.RawSetString("wday", lua.LNumber(t.Weekday()))
	// TODO yday & dst
	ret.RawSetString("yday", lua.LNumber(0))
	ret.RawSetString("isdst", lua.LFalse)
	L.Push(ret)
	return 1
}

// +gendoc
// * nsec = time.since(t)
//
// 時刻tからの経過時間をナノ秒nsecで返します。時刻tはtime.now()で
// 得られた結果のみを受け付けます。
//
// この関数を使って、スクリプトの実行時間を計ることが可能です。
//
//   local now = time.now()          -- 現在時刻を取得
//   heavy_process()                 -- 重たい処理
//   local delta_t = time.since(now) -- ナノ秒単位の経過時間delta_t
func timeSince(L *lua.LState) int {
	t := checkTime(L, 1)
	L.Push(lua.LNumber(time.Since(t)))
	return 1
}

// +gendoc
// * year = time.year(t)
//
// 時刻tの中身から年yearを取り出します。
func timeYear(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Year()))
	return 1
}

// +gendoc
// * month = time.month(t)
//
// 時刻tの中身から月monthを取り出します。
// monthは[1, 12]の範囲の数値です。
func timeMonth(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Month()))
	return 1
}

// +gendoc
// * day = time.day(t)
//
// 時刻tの中身から日dayを取り出します。
// dayは[1, 31]の範囲の数値です。
func timeDay(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Day()))
	return 1
}

// +gendoc
// * hour = time.hour(t)
//
// 時刻tの中身から時間hourを取り出します。
// hourは[0, 23]の範囲の数値です。
func timeHour(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Hour()))
	return 1
}

// +gendoc
// * minute = time.minute(t)
//
// 時刻tの中身から分minuteを取り出します。
// minuteは[0, 59]の範囲の数値です。
func timeMinute(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Minute()))
	return 1
}

// +gendoc
// * second = time.second(t)
//
// 時刻tの中身から秒secondを取り出します。
// secondは[0, 59]の範囲の数値です。
func timeSecond(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Second()))
	return 1
}

// +gendoc
// * weekday = time.weekday(t)
//
// 時刻tの中身から曜日weekdayを取り出します。
// weekdayは日曜日を0とした[0, 6]の範囲の数値です。
func timeWeekday(L *lua.LState) int {
	L.Push(lua.LNumber(checkTime(L, 1).Weekday()))
	return 1
}

// +gendoc
// * time_string = time.format(t, [fmt_string])
//
// 時刻tを文字列表現に変換します。
// デフォルトで"年/月/日 時:分:秒"の形式に変換します。
// fmt_stringを与えることで、変換形式を変えることができます。
// fmt_stringの書式はGolang timeパッケージを参照 (https://godoc.org/time)
func timeFormat(L *lua.LState) int {
	t := checkTime(L, 1)
	str := t.Format(L.OptString(2, "2006/01/02 15:04:05"))
	L.Push(lua.LString(str))
	return 1
}

// +gendoc
// * sec_string = time.tostring(nanosecond)
//
// ナノ秒を、人に見やすい形式の文字列に変換します。
// Example:
//   local second = 1 * time.SECOND  -- 1秒をナノ秒単位で
//   log.info(second)                -- "1000000000"と出力されてしまう
//   str = time.tostring(second)     -- 1秒を文字列形式に変換
//   log.info(str)                   -- "1s"と出力
func timeToString(L *lua.LState) int {
	d := checkDuration(L, 1)
	L.Push(lua.LString(d.String()))
	return 1
}
