package script

import (
	"github.com/yuin/gopher-lua"
)

// references to [ lua-users wiki: Sand Boxes ] http://lua-users.org/wiki/SandBoxes
var unsafeLibs = []string{
	"print", // write stdout is not allowed
}

func knockoutUnsafeLibs(L *lua.LState) {
	for _, func_name := range unsafeLibs {
		L.SetGlobal(func_name, lua.LNil)
	}
}

// references to [ lua-users wiki: Sand Boxes ] http://lua-users.org/wiki/SandBoxes
var safe_libs = []string{
	"error",
	"assert",
	"ipairs",
	"next",
	"pairs",
	"pcall",
	"select",
	"tonumber",
	"tostring",
	"type",
	"unpack",
	"_VERSION",
	"xpcall",
	"setmetatable",
	"getmetatable",
	"coroutine.create",
	"coroutine.resume",
	"coroutine.running",
	"coroutine.status",
	"coroutine.wrap",
	"coroutine.yield",
	"string.byte",
	"string.char",
	"string.dump",
	"string.find",
	"string.format",
	"string.gmatch",
	"string.gsub",
	"string.len",
	"string.lower",
	"string.match",
	"string.rep",
	"string.reverse",
	"string.sub",
	"string.upper",
	"table.insert",
	"table.maxn",
	"table.remove",
	"table.sort",
	"math.abs",
	"math.acos",
	"math.asin",
	"math.atan",
	"math.atan2",
	"math.ceil",
	"math.cos",
	"math.cosh",
	"math.deg",
	"math.exp",
	"math.floor",
	"math.fmod",
	"math.frexp",
	"math.huge",
	"math.ldexp",
	"math.log",
	"math.log10",
	"math.max",
	"math.min",
	"math.modf",
	"math.pi",
	"math.pow",
	"math.rad",
	"math.random",
	"math.randomseed", // - UNSAFE (maybe) - see math.random",
	"math.sin",
	"math.sinh",
	"math.sqrt",
	"math.tan",
	"math.tanh",
	"os.clock",
	"os.date",
	"os.difftime",
	"os.time",
}

// // references to [花映塚AI自作ツール] http://www.usamimi.info/~ide/programe/touhouai/report-20140705.pdf
// var lib_regexp = regexp.MustCompile(`([a-zA-Z0-9_]+)\.?([a-zA-Z0-9_]+)?`)
//
// func (ip Interpreter) registerSafeLuaLibs(env *lua.LTable) error {
// 	L := ip.vm
//
// 	// register safe lua standard lib
// 	for _, lib_name := range safe_libs {
// 		matches := lib_regexp.FindStringSubmatch(lib_name)
// 		if len(matches) != 3 {
// 			panic(lib_name + " does not 2-SubMatchString")
// 			//continue
// 		}
//
// 		if len(matches[2]) != 0 { // case module.function
// 			mod_name, fn_name := matches[1], matches[2]
// 			mod := L.GetGlobal(mod_name)
// 			if mod.Type() != lua.LTTable {
// 				continue
// 			}
// 			if env_mod := L.GetField(env, mod_name); env_mod.Type() != lua.LTTable {
// 				L.SetField(env, mod_name, L.NewTable())
// 			}
//
// 			fn := L.GetField(mod, fn_name)
// 			if fn.Type() != lua.LTFunction {
// 				continue
// 			}
// 			env_mod := L.GetField(env, mod_name)
// 			if table, ok := env_mod.(*lua.LTable); ok {
// 				L.SetField(table, fn_name, fn)
// 			}
//
// 		} else if len(matches[1]) != 0 { // case function
// 			val_name := matches[1]
// 			val := L.GetGlobal(val_name)
// 			L.SetField(env, val_name, val)
//
// 		} else {
// 			panic("invalid string " + lib_name + " to match")
// 		}
// 	}
//
// 	// register limited lua lib
// 	L.SetField(env, "loadfile", L.NewFunction(ip.safeLoadFile))
// 	L.SetField(env, "dofile", L.NewFunction(ip.safeDoFile))
// 	L.SetField(env, "require", L.NewFunction(ip.safeRequire))
// 	L.SetField(env, "module", L.NewFunction(libModule))
//
// 	// // strict io open
// 	// if io_module := L.GetField(env, "io"); io_module.Type() == lua.LTTable {
// 	// 	L.SetField(io_module, "open", L.NewFunction(safeIoOpen))
// 	// } else {
// 	// 	new_io_module := L.NewTable()
// 	// 	L.SetField(new_io_module, "open", L.NewFunction(safeIoOpen))
// 	// 	L.SetField(env, "io", new_io_module)
// 	// }
//
// 	L.SetField(env, "_G", env) // as Global environmet
// 	return nil
// }
//
// // loadfile
// func (ip Interpreter) safeLoadFile(L *lua.LState) int {
// 	file := checkFilePath(L, 1)
//
// 	fn, err := L.LoadFile(ip.PathOf(file))
// 	if err != nil {
// 		L.ArgError(1, "can not compile script: "+err.Error())
// 	}
//
// 	L.Push(fn)
// 	return 1
// }
//
// // dofile
// func (ip Interpreter) safeDoFile(L *lua.LState) int {
// 	ip.safeLoadFile(L)
// 	L.Call(0, 0)
// 	return 0
// }
//
// var whileloading = &lua.LUserData{}
//
// // require
// // references to gopher-lua/loadlib.go#loRequire()
// func (ip Interpreter) safeRequire(L *lua.LState) int {
// 	name := L.CheckString(1)
// 	loaded := L.GetField(L.Get(lua.RegistryIndex), "_LOADED") // TODO: use ip.loaded
//
// 	// check module is already loaded?
// 	if lv := L.GetField(loaded, name); lua.LVAsBool(lv) {
// 		if lv == whileloading {
// 			L.RaiseError("loop or previous error loading module: %s", name)
// 		}
// 		L.Push(lv)
// 		return 1
// 	}
//
// 	// restrict load file path
// 	const LuaExtentsion = ".lua"
// 	loadpath := ip.PathOf(strings.Replace(name, ".", string(os.PathSeparator), -1))
// 	loadpath += LuaExtentsion
// 	modfn, err := L.LoadFile(loadpath)
// 	if err != nil {
// 		L.RaiseError(err.Error())
// 		return 0
// 	}
//
// 	// call module function
// 	L.SetField(loaded, name, whileloading)
// 	L.Push(modfn)
// 	L.Push(lua.LString(name))
// 	L.Call(1, 1)
// 	ret := L.Get(-1)
// 	L.Pop(1)
// 	modv := L.GetField(loaded, name)
// 	if ret != lua.LNil && modv == whileloading {
// 		L.SetField(loaded, name, ret)
// 		L.Push(ret)
// 	} else if modv == whileloading {
// 		L.SetField(loaded, name, lua.LTrue)
// 		L.Push(lua.LTrue)
// 	} else {
// 		L.Push(modv)
// 	}
// 	return 1
// }
//
// // module NOTE: not supported
// func libModule(L *lua.LState) int {
// 	L.RaiseError(`module() is not supported.
// to define module, use lua5.2 style:
// 	local _M = {}
// 	_M._NAME = (...)
// 	_M.func1 = function(arg) {}
// 	_M.private_value = value
// 	return _M
// 	`)
// 	return 0
// }
//
// // io.open
// func safeIoOpen(L *lua.LState) int {
// 	file := checkFilePath(L, 1)
// 	open := L.GetField(L.GetGlobal("io"), "open")
//
// 	L.Push(open)
// 	L.Push(lua.LString(file))
// 	if err := L.PCall(1, 1, nil); err != nil {
// 		L.ArgError(1, "io.open error: "+err.Error())
// 	}
// 	return 1
// }
//
// func checkFilePath(L *lua.LState, pos int) string {
// 	if file := L.CheckString(pos); validateExtension(file) {
// 		return file
// 	}
// 	L.ArgError(pos, "invalid file path")
// 	return ""
// }
//
// // end with .lua
// const ScriptExtension = ".lua"
//
// // path is vaild?
// func validateExtension(file string) bool {
// 	return strings.HasSuffix(file, ScriptExtension)
// }
