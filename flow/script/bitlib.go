package script

import (
	"github.com/yuin/gopher-lua"
)

const bit32ModuleName = "bit32"

// +gendoc.set_section "Builtin Module: bit32"

// +gendoc
// 32bitまでのビット演算を行うモジュール。
//
// Lua5.1では、ビット演算子が存在しないため、
// その代替として用意されています。
// また、Lua5.1では、and, or などが予約語になっているため、
// bit32.band, bit32.borのようにbが前置きされた名前を使用します。
// モジュールを利用する場合には以下のようにします。
//
//   bit32 = require "bit32"
//   bit32.band(3, 1) == 1  -- 0x11 AND 0x01 = 0x01
//   bit32.bor(3, 1) == 3   -- 0x11 OR 0x01 = 0x11

// references to http://lua.tips/download/func/bit32/lbitlib.c
// in which a part of Lua5.2's bit32 modules are used.
//
// Copyright (C) 1994-2013 Lua.org, PUC-Rio.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// bit32 module loader for lua interpreter.
func bit32Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), bit32Exports)
	L.Push(mod)
	return 1
}

var bit32Exports = map[string]lua.LGFunction{
	"band":     bitAnd,
	"bor":      bitOr,
	"bxor":     bitXor,
	"bnot":     bitNot,
	"rshift":   bitRShift,
	"lshift":   bitLShift,
	"set":      bitSet,
	"unset":    bitUnset,
	"get":      bitGet,
	"popcount": bitPopCount,
}

func checkUInt64(L *lua.LState, pos int) uint64 {
	n := L.CheckNumber(pos)
	if n < 0 {
		L.ArgError(pos, "bit argument must be n >= 0")
	}
	return uint64(n)
}

const mask32 = 0x00000000ffffffff

func trim32(n uint64) uint64 {
	return mask32 & n
}

const mask64 = 0xffffffffffffffff

// +gendoc
// * result = bit32.band(numbers...)
//
// 渡された数値全てをAND演算した結果を返します。
func bitAnd(L *lua.LState) int {
	top := L.GetTop()
	var r uint64 = mask64
	for i := 1; i <= top; i++ {
		r &= checkUInt64(L, i)
	}
	L.Push(lua.LNumber(trim32(r)))
	return 1
}

// +gendoc
// * result = bit32.bor(numbers...)
//
// 渡された数値全てをOR演算した結果を返します。
func bitOr(L *lua.LState) int {
	top := L.GetTop()
	r := uint64(0)
	for i := 1; i <= top; i++ {
		r |= checkUInt64(L, i)
	}
	L.Push(lua.LNumber(trim32(r)))
	return 1
}

// +gendoc
// * result = bit32.bxor(numbers...)
//
// 渡された数値全てをXOR演算した結果を返します。
func bitXor(L *lua.LState) int {
	top := L.GetTop()
	r := uint64(0)
	for i := 1; i <= top; i++ {
		r ^= checkUInt64(L, i)
	}
	L.Push(lua.LNumber(trim32(r)))
	return 1
}

// +gendoc
// * result = bit32.bnot(number)
//
// 渡された数値1つをNOT演算した結果を返します。
func bitNot(L *lua.LState) int {
	n := checkUInt64(L, 1)
	n = mask64 &^ n
	L.Push(lua.LNumber(trim32(n)))
	return 1
}

// +gendoc
// * result = bit32.rshift(number, offset)
//
// 渡された数値numberをoffset回右に論理シフトした結果を返します。
func bitRShift(L *lua.LState) int {
	n := checkUInt64(L, 1)
	shift := L.CheckInt(2)
	n = n >> uint(shift)
	L.Push(lua.LNumber(trim32(n)))
	return 1
}

// +gendoc
// * result = bit32.lshift(number, offset)
//
// 渡された数値numberをoffset回左に論理シフトした結果を返します。
func bitLShift(L *lua.LState) int {
	n := checkUInt64(L, 1)
	shift := L.CheckInt(2)
	n = n << uint(shift)
	L.Push(lua.LNumber(trim32(n)))
	return 1
}

// +gendoc
// * result = bit32.set(number, offset)
//
// 渡された数値numberの、2^offsetに対応するビットを1にした結果を返します。
func bitSet(L *lua.LState) int {
	n := checkUInt64(L, 1)
	pos := L.CheckInt(2)
	n |= 1 << uint(pos)
	L.Push(lua.LNumber(trim32(n)))
	return 1
}

// +gendoc
// * result = bit32.unset(number, offset)
//
// 渡された数値numberの、2^offsetに対応するビットを0にした結果を返します。
func bitUnset(L *lua.LState) int {
	n := checkUInt64(L, 1)
	pos := L.CheckInt(2)
	n &^= 1 << uint(pos) // AND NOT, NOT(000100) -> (111011) then AND.
	L.Push(lua.LNumber(trim32(n)))
	return 1
}

// +gendoc
// * 0_or_1 = bit32.get(number, offset)
//
// 渡された数値numberの、2^offsetに対応するビットを返します。
func bitGet(L *lua.LState) int {
	n := checkUInt64(L, 1)
	pos := L.CheckInt(2)
	n = (n >> uint(pos)) & 1
	L.Push(lua.LNumber(n))
	return 1
}

// +gendoc
// * number_of_1 = bit32.popcount(number)
//
// 渡された数値numberの、ビット単位の1の個数を返します。
// 例：
//   x = 0x03 -- 00000011
//   popcount(x) => 2
//
//   y = 0x10 -- 00010000
//   popcount(y) => 1
func bitPopCount(L *lua.LState) int {
	n := trim32(checkUInt64(L, 1))
	n = (n & 0x5555555555555555) + ((n & 0xAAAAAAAAAAAAAAAA) >> 1)
	n = (n & 0x3333333333333333) + ((n & 0xCCCCCCCCCCCCCCCC) >> 2)
	n = (n & 0x0F0F0F0F0F0F0F0F) + ((n & 0xF0F0F0F0F0F0F0F0) >> 4)
	// reduced sum per 8bit. references http://d.hatena.ne.jp/s-yata/20120419/1334845666
	n *= 0x0101010101010101
	// reduced sum is stored at top 8bits.
	n = (n >> 56) & 0xFF
	L.Push(lua.LNumber(n))
	return 1
}
