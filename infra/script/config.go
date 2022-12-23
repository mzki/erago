package script

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ScriptConfig holds Script parameters.
type Config struct {
	LoadDir     string
	LoadPattern string

	CallStackSize       int
	RegistrySize        int
	IncludeGoStackTrace bool

	InfiniteLoopTimeoutSecond int
	ReloadFileChange          bool
}

var (
	// default paramters for script VM.
	LoadPattern               = "init.lua"
	CallStackSize             = lua.CallStackSize
	RegistrySize              = lua.RegistrySize
	InfiniteLoopTimeoutSecond = 10 * time.Second
)

func (c Config) loadPattern() string {
	return filepath.Join(c.LoadDir, c.LoadPattern)
}

const (
	registryBaseDirKey     = "_BASE_DIRECTORY"
	registryDebugEnableKey = "_DEBUG_ENABLE"
)

func (conf Config) register(L *lua.LState) {
	reg := L.CheckTable(lua.RegistryIndex)
	for _, set := range []struct {
		key string
		val lua.LValue
	}{
		{registryDebugEnableKey, lua.LBool(conf.IncludeGoStackTrace)},
		{registryBaseDirKey, lua.LString(filepath.Clean(conf.LoadDir))},
	} {
		reg.RawSetString(set.key, set.val)
		L.SetGlobal(set.key, set.val)
	}
}

// check filepath at argument i.
func checkFilePath(L *lua.LState, i int) string {
	path, err := scriptPath(L, L.CheckString(i))
	if err != nil {
		L.ArgError(i, err.Error())
	}
	return path
}

// return path of p under script base directory.
// for example, base dir is "/dir" and p is "sub/file" then
// return "/dir/sub/file".
// if resulted path indicates above base directory, e.g. including  "../",
// it will return error.
func scriptPath(L *lua.LState, p string) (string, error) {
	lv := L.CheckTable(lua.RegistryIndex).RawGetString(registryBaseDirKey)
	basedir := lua.LVAsString(lv)
	joined := filepath.Clean(filepath.Join(basedir, p))
	if !strings.HasPrefix(joined, basedir) {
		return joined, fmt.Errorf("given path %s must be under %s, but you specifies %s", p, basedir, joined)
	}
	return joined, nil
}
