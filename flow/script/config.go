package script

import (
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
	registryBaseDirKey     = "_base_directory"
	registryDebugEnableKey = "_debug_enable"
)

func (conf Config) register(L *lua.LState) {
	reg := L.CheckTable(lua.RegistryIndex)
	reg.RawSetString(registryDebugEnableKey, lua.LBool(conf.ShowGoStackTrace))
	reg.RawSetString(registryBaseDirKey, lua.LString(conf.path.Join(conf.LoadDir)))
}

// return path of p under script base directory.
// for example, base dir is "/dir" and p is "sub/file" then
// return "/dir/sub/file".
func scriptPath(L *lua.LState, p string) string {
	lv := L.CheckTable(lua.RegistryIndex).RawGetString(registryBaseDirKey)
	basedir := lua.LVAsString(lv)
	joined := filepath.Join(basedir, p)
	if !strings.HasPrefix(joined, basedir) {
		// TODO: joined references unexpected directory, should panic?
		joined = filepath.Join(basedir, filepath.Base(p))
	}
	return joined
}
