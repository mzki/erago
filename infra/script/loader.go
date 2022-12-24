package script

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/util/errutil"
	lua "github.com/yuin/gopher-lua"
)

type customLoaderConfig struct {
	watchChange bool
}

// CustomLoaders holds platform depended file loaders which is used by package.loader
// to search module name and import additional script into the interpreter environtment.
type customLoaders struct {
	loaders  map[filesystem.RFileSystemPR]*loaderHelper
	registry map[*lua.LState]*lua.LFunction
	config   customLoaderConfig
}

func newCustomLoaders(watchChange bool) *customLoaders {
	return &customLoaders{
		loaders:  make(map[filesystem.RFileSystemPR]*loaderHelper, 4),
		registry: make(map[*lua.LState]*lua.LFunction),
		config:   customLoaderConfig{watchChange},
	}
}

func (ldrs *customLoaders) LuaLoader(L *lua.LState) int {
	if len(ldrs.loaders) == 0 {
		L.Push(lua.LString("customLoaders: empty custom loaders"))
		return 1
	}

	name := L.CheckString(1)
	merr := errutil.NewMultiError()

	for ldr, ldrHelper := range ldrs.loaders {
		modpath, msg := ldrs.findFile(L, name, ldr)
		if len(modpath) == 0 {
			merr.Add(fmt.Errorf(msg))
			continue
		}
		lfunc, err := ldrs.loadLFunc(L, ldr, modpath, name)
		if err != nil {
			merr.Add(err)
			continue
		}

		// load successfully

		// watch ldrHelper
		if ldrs.ShouldWatch(ldrHelper.watcher) {
			if err := ldrHelper.watcher.Watch(modpath); err != nil {
				merr.Add(err)
				continue
			}
		}
		ldrHelper.pathAndNameMap[modpath] = name
		L.Push(lfunc)
		return 1
	}

	// load failed or not found
	// TODO set cunstom loader error into VM, to retrieve it later
	// return just error message to adapt to package.loader specification.
	L.Push(lua.LString(merr.Err().Error()))
	return 1
}

func (ldrs *customLoaders) loadLFunc(L *lua.LState, ldr filesystem.RFileSystemPR, path, name string) (*lua.LFunction, error) {
	reader, err := ldr.Load(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return L.Load(reader, name)
}

func (ldrs *customLoaders) reload(L *lua.LState, ldr filesystem.RFileSystemPR, path string) error {
	var name string
	if ldrHelper, ok := ldrs.loaders[ldr]; !ok {
		return fmt.Errorf("reload: unknown FileSystem")
	} else {
		name, ok = ldrHelper.pathAndNameMap[path]
		if !ok {
			return fmt.Errorf("reload: file path %s is not loaded", path)
		}
	}

	lfunc, err := ldrs.loadLFunc(L, ldr, path, name)
	if err != nil {
		return err
	}

	L.Push(lfunc)
	L.Push(lua.LString(name))
	err = L.PCall(1, 1, nil)
	if err != nil {
		return err
	}
	lfuncRet := L.Get(1)
	L.Pop(1)

	// Update also loaded result to reflect change to Lua world.
	loaded := L.GetField(L.Get(lua.RegistryIndex), "_LOADED").(*lua.LTable)
	if lfuncRet == lua.LNil {
		L.SetField(loaded, name, lua.LTrue)
	} else if lfuncRet.Type() == lua.LTTable {
		// update old table
		oldValue := L.GetField(loaded, name)
		if oldValue.Type() != lua.LTTable {
			L.SetField(loaded, name, lfuncRet)
		} else {
			oldTable := oldValue.(*lua.LTable)
			newTable := lfuncRet.(*lua.LTable)
			newTable.ForEach(func(k, v lua.LValue) {
				L.RawSet(oldTable, k, v)
			})
		}
	} else {
		L.SetField(loaded, name, lfuncRet)
	}
	return nil
}

// find lua file path from package.path. See
// https://github.com/yuin/gopher-lua/blob/c841877397d8e2ef0bd755390798e6cb957590a9/loadlib.go#L30
func (ldrs *customLoaders) findFile(L *lua.LState, name string, ldr filesystem.RFileSystemPR) (string, string) {
	name = strings.Replace(name, ".", string(os.PathSeparator), -1)
	lv := L.GetField(L.GetField(L.Get(lua.EnvironIndex), "package"), "path")
	path, ok := lv.(lua.LString)
	if !ok {
		return "", "package.path must be a string"
	}
	messages := []string{}
	for _, pattern := range strings.Split(string(path), ";") {
		luapath := strings.Replace(pattern, "?", name, -1)
		if err := validateScriptPathL(L, luapath); err != nil {
			messages = append(messages, fmt.Sprintf("invalid script name %s: %s", name, err.Error()))
		} else {
			if ok := ldr.Exist(luapath); ok {
				return luapath, ""
			} else {
				messages = append(messages, fmt.Sprintf("%s is not found in %s", name, pattern))
			}
		}
	}
	return "", strings.Join(messages, "\n\t")
}

// Max loader count. Since default VM may have 2 builtin loaders,
// user can actually add loaderRegistryMax - 2 loaders.
const loaderRegistryMax = 4

// register custom loader to given lua state.
// It return nil when registeration is succeeded or erorr if failed.
func (ldrs *customLoaders) Register(L *lua.LState) error {
	if _, ok := ldrs.registry[L]; ok {
		return nil
	}

	// NOTE: get package.loader from registery index, since package module may be removed
	loaders, ok := L.GetField(L.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
	if !ok {
		panic("package.loader is not found on this VM")
	}

	for i := 1; i <= loaderRegistryMax; i++ {
		if ldr := L.RawGetInt(loaders, i); ldr == lua.LNil {
			// insert point found
			customLoader := L.NewFunction(ldrs.LuaLoader)
			L.RawSetInt(loaders, i, customLoader)
			ldrs.registry[L] = customLoader
			return nil
		}
	}

	return fmt.Errorf("maximum loader count is exceeded")
}

func (ldrs *customLoaders) Unregister(L *lua.LState) {
	registerFunc, ok := ldrs.registry[L]
	if !ok {
		// not registered
		return
	}
	// should have been registered

	// NOTE: get package.loader from registery index, since package module may be removed
	loaders, ok := L.GetField(L.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
	if !ok {
		panic("package.loader is not found on this VM")
	}

	for i := 1; i <= loaderRegistryMax; i++ {
		if ldr := L.RawGetInt(loaders, i); ldr == registerFunc {
			// remove point found
			L.RawSetInt(loaders, i, lua.LNil)
			return
		}
	}

	panic("never reached")
}

func (ldrs *customLoaders) ShouldWatch(w filesystem.Watcher) bool {
	return ldrs.config.watchChange && w != nil
}

func (ldrs *customLoaders) Add(ld filesystem.RFileSystemPR, fn onFileChangedFunc, errFn onFileChangeErrorFunc) error {
	var watcher filesystem.Watcher = nil
	if ldrs.config.watchChange {
		w, err := filesystem.OpenWatcherPR(ld)
		if err != nil {
			return fmt.Errorf("AddCustomLoader failed: %w", err)
		}
		watcher = w
	}
	ldrHelper := &loaderHelper{
		watcher:           watcher,
		pathAndNameMap:    make(map[string]string, 32),
		onFileChanged:     fn,
		onFileChangeError: errFn,
	}
	if ldrs.config.watchChange {
		closeCh := make(chan struct{})
		done := ldrHelper.watchFileChange(closeCh)
		ldrHelper.watchCloseReq = closeCh
		ldrHelper.watchDone = done
	}
	ldrs.loaders[ld] = ldrHelper
	return nil
}

func (ldrs *customLoaders) Remove(ld filesystem.RFileSystemPR) error {
	ldrHelper, ok := ldrs.loaders[ld]
	if !ok {
		return fmt.Errorf("remove requested RFileSystemPR is not found")
	}
	if ldrs.ShouldWatch(ldrHelper.watcher) {
		if err := ldrHelper.watcher.Close(); err != nil {
			return fmt.Errorf("remove requested Watcher failed: %w", err)
		}
		close(ldrHelper.watchCloseReq)
		select {
		case <-ldrHelper.watchDone:
			// do nothing
		case <-time.After(3 * time.Second):
			return fmt.Errorf("watchFileChange never end")
		}
	}
	delete(ldrs.loaders, ld)
	return nil
}

func (ldrs *customLoaders) RemoveAll() error {
	merr := errutil.NewMultiError()
	for ld, _ := range ldrs.loaders {
		err := ldrs.Remove(ld)
		merr.Add(err)
	}
	return merr.Err()
}

type onFileChangedFunc = func(path string)
type onFileChangeErrorFunc = func(error)

type loaderHelper struct {
	watcher           filesystem.Watcher
	pathAndNameMap    map[string]string
	onFileChanged     onFileChangedFunc
	onFileChangeError onFileChangeErrorFunc

	watchCloseReq chan<- struct{}
	watchDone     <-chan struct{}
}

func (lh *loaderHelper) watchFileChange(closech <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case ev, ok := <-lh.watcher.Events():
				if !ok {
					return
				}
				if ev.Has(filesystem.WatchOpWrite) {
					if lh.onFileChanged != nil {
						lh.onFileChanged(ev.Name)
					}
				}
			case err, ok := <-lh.watcher.Errors():
				if !ok {
					return
				}
				if lh.onFileChangeError != nil {
					lh.onFileChangeError(err)
				}
			case <-closech:
				return
			}
		}
	}()
	return done
}

// -- Interprer APIs

func (ip *Interpreter) AddCustomLoader(ld filesystem.RFileSystemPR) error {
	return ip.customLoaders.Add(ld, onFileChangedFunc(func(path string) {
		ip.appendTask(func() error { return ip.reload(ld, path) })
	}), onFileChangeErrorFunc(func(err error) {
		ip.appendTask(createTerminateTask(err))
	}))
}

func (ip *Interpreter) RemoveCustomLoader(ld filesystem.RFileSystemPR) error {
	return ip.customLoaders.Remove(ld)
}

func (ip *Interpreter) reload(ldr filesystem.RFileSystemPR, path string) error {
	return ip.customLoaders.reload(ip.vm, ldr, path)
}

func knockoutDefaultLuaLoader(L *lua.LState) {
	// NOTE: get package.loader from registery index, since package module may be removed
	loaders := L.GetField(L.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
	// Remove default lua loader. This is based on assumption package.loader registers {preload, lualoader}.
	// The assumption is valid if LState is just initialized, not modifying any to default package.loader.
	L.RawSetInt(loaders, 2, lua.LNil)
}
