package script

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yuin/gopher-lua"

	"local/erago/util"
)

// ScriptConfig holds Script parameters.
type Config struct {
	path util.PathManager

	LoadDir     string `toml:"load_dir"`
	LoadPattern string `toml:"load_pattern"`

	CallStackSize    int  `toml:"call_stack_size"`
	RegistrySize     int  `toml:"registry_size"`
	ShowGoStackTrace bool `toml:"debug"`
}

var (
	defaultScriptDir     = "ELA"
	defaultScriptPattern = "init.lua"

	defaultCallStackSize = lua.CallStackSize
	defaultRegistrySize  = lua.RegistrySize
)

func NewConfig(basedir string) Config {
	return Config{
		path:             util.NewPathManager(basedir),
		LoadDir:          defaultScriptDir,
		LoadPattern:      defaultScriptPattern,
		CallStackSize:    defaultCallStackSize,
		RegistrySize:     defaultRegistrySize,
		ShowGoStackTrace: false,
	}
}

// set base directory. All of Expoerted Field is prefixed by base dir.
func (c *Config) SetBaseDir(basedir string) {
	c.path = util.NewPathManager(basedir)
}

func (c Config) loadPattern() string {
	return c.path.Join(c.LoadDir, c.LoadPattern)
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
		{registryDebugEnableKey, lua.LBool(conf.ShowGoStackTrace)},
		{registryBaseDirKey, lua.LString(conf.path.Join(conf.LoadDir))},
	} {
		reg.RawSetString(set.key, set.val)
		L.SetGlobal(set.key, set.val)
	}
}

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
		return joined, fmt.Errorf("given path %s must be under %s", p, basedir)
	}
	return joined, nil
}
