package script

import (
	"fmt"

	"github.com/mzki/erago/infra/loader"
	"github.com/mzki/erago/util/errutil"
	lua "github.com/yuin/gopher-lua"
)

// CustomLoaders holds platform depended file loaders which is used by package.loader
// to search module name and import additional script into the interpreter environtment.
type customLoaders struct {
	loaders map[loader.Loader]struct{}

	registry map[*lua.LState]*lua.LFunction
}

func newCustomLoaders() *customLoaders {
	return &customLoaders{
		loaders:  make(map[loader.Loader]struct{}, 4),
		registry: make(map[*lua.LState]*lua.LFunction),
	}
}

func (ldrs *customLoaders) LuaLoader(L *lua.LState) int {
	if len(ldrs.loaders) == 0 {
		L.Push(lua.LString("customLoaders: empty custom loaders"))
		return 1
	}

	modpath := L.CheckString(1)
	merr := errutil.NewMultiError()

	for ldr, _ := range ldrs.loaders {
		reader, err := ldr.Load(modpath)
		if err != nil {
			merr.Add(err)
			continue
		}
		defer reader.Close() // calling Close() is statcked with current reader.

		lfunc, err := L.Load(reader, modpath)
		if err != nil {
			merr.Add(err)
			continue
		}

		// load successfully
		L.Push(lfunc)
		return 1
	}

	// load failed or not found
	// TODO set cunstom loader error into VM, to retrieve it later
	// return just error message to adapt to package.loader specification.
	L.Push(lua.LString(merr.Err().Error()))
	return 1
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

func (ldrs *customLoaders) Add(ld loader.Loader) {
	ldrs.loaders[ld] = struct{}{}
}

func (ldrs *customLoaders) Remove(ld loader.Loader) {
	delete(ldrs.loaders, ld)
}

func (ldrs *customLoaders) RemoveAll() {
	for ld, _ := range ldrs.loaders {
		ldrs.Remove(ld)
	}
}

// -- Interprer APIs

func (ip *Interpreter) AddCustomLoader(ld loader.Loader) {
	ip.customLoaders.Add(ld)
}

func (ip *Interpreter) RemoveCustomLoader(ld loader.Loader) {
	ip.customLoaders.Remove(ld)
}
