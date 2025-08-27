package script

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/buildinfo"
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
	registryBaseDirKey        = "_BASE_DIRECTORY"
	registryDebugEnableKey    = "_DEBUG_ENABLE"
	registryRuntimeVersionKey = "_RUNTIME_VERSION"
)

func (conf Config) register(L *lua.LState) error {
	resLoadDir, err := filesystem.ResolvePath(filepath.Clean(conf.LoadDir))
	if err != nil {
		return fmt.Errorf("config value register failed by ResolvePath %s: %w", conf.LoadDir, err)
	}

	binfo := buildinfo.Get()
	reg := L.CheckTable(lua.RegistryIndex)
	for _, set := range []struct {
		key string
		val lua.LValue
	}{
		{registryDebugEnableKey, lua.LBool(conf.IncludeGoStackTrace)},
		{registryBaseDirKey, lua.LString(resLoadDir)},
		{registryRuntimeVersionKey, lua.LString(binfo.Version)},
	} {
		reg.RawSetString(set.key, set.val)
		L.SetGlobal(set.key, set.val)
	}
	return nil
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
	if err := validateScriptPath(joined, basedir); err != nil {
		return "", fmt.Errorf("make scriptPath failed: %w", err)
	}
	return joined, nil
}

func validateScriptPathL(L *lua.LState, p string) error {
	lv := L.CheckTable(lua.RegistryIndex).RawGetString(registryBaseDirKey)
	baseDir := lua.LVAsString(lv)
	return validateScriptPath(p, baseDir)
}

func validateScriptPath(p, baseDir string) error {
	cleanP := filepath.Clean(p)
	cleanD := filepath.Clean(baseDir) + string(os.PathSeparator) // ends with "/"
	if !strings.HasPrefix(cleanP, cleanD) {
		return fmt.Errorf("given path %s must be under %s", p, baseDir)
	}
	return nil
}
