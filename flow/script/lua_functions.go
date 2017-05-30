package script

import (
	"github.com/yuin/gopher-lua"
)

// references to [ lua-users wiki: Sand Boxes ] http://lua-users.org/wiki/SandBoxes
var unsafeLibs = []string{
	"print", // write stdout is not allowed
	"dofile",
	"dostring",
	"load",
	"loadfile",
	"loadstring",
}

func knockoutUnsafeLibs(L *lua.LState) {
	for _, func_name := range unsafeLibs {
		L.SetGlobal(func_name, lua.LNil)
	}
}
