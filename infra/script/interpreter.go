package script

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/state"

	lua "github.com/yuin/gopher-lua"
)

// Interpreter parse script files.
// it runs user script in the strict environment.
//
// typical usage:
//
//	ip := NewInterpreter(...)
//	defer ip.Quit()
type Interpreter struct {
	vm        *lua.LState
	eraModule *lua.LTable

	state *state.GameState
	game  GameController

	customLoaders *customLoaders
	watchDogTimer *watchDogTimer

	config Config
}

// construct interpreter.
// must be call Interpreter.Quit after use this.
func NewInterpreter(s *state.GameState, g GameController, config Config) *Interpreter {
	vm := lua.NewState(lua.Options{
		CallStackSize:       config.CallStackSize,
		RegistrySize:        config.RegistrySize,
		IncludeGoStackTrace: config.IncludeGoStackTrace,
		SkipOpenLibs:        true,
	})

	ip := &Interpreter{
		vm:            vm,
		state:         s,
		game:          g,
		customLoaders: newCustomLoaders(),
		watchDogTimer: newWatchDogTimer(
			time.Duration(config.InfiniteLoopTimeoutSecond) * time.Second),
		config: config,
	}
	ip.init()
	return ip
}

// initialize sandbox environment and user defined data.
func (ip *Interpreter) init() {
	L := ip.vm
	// register bultin libraries which do not contain
	// the modules to access file system and OS.
	for _, pair := range []struct {
		name string
		open lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.MathLibName, lua.OpenMath},
		{lua.StringLibName, lua.OpenString},
		// {lua.ChannelLibName, lua.OpenChannel}, // Channel is not used.
		{lua.CoroutineLibName, lua.OpenCoroutine},
		{lua.DebugLibName, lua.OpenDebug},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.open),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.name)); err != nil {
			panic(err)
		}
	}
	knockoutUnsafeLibs(L)

	for _, mod := range []struct {
		Name   string
		Loader lua.LGFunction
	}{
		{bit32ModuleName, bit32Loader},
		{loggerModuleName, loggerLoader},
		{timeModuleName, timeLoader},
	} {
		L.PreloadModule(mod.Name, mod.Loader)
	}

	ip.eraModule = registerEraModule(L, ip.state, ip.game, ip.watchDogTimer)
	registerSystemParams(L, ip.state)
	registerCsvParams(L, ip.state.CSV)
	registerCharaParams(L, ip.state)
	registerMisc(L)
	registerExtPairs(L)

	// register load path which is limited under config.LoadDir only.
	// NOTE: bultin path is cleared. ( /usr/local/share/lua5.1  etc. are not available)
	reg_path, err := filesystem.ResolvePath(filepath.Join(ip.config.LoadDir, "?.lua"))
	if err != nil {
		panic(err) // it should never occurs
	}
	L.SetField(L.GetGlobal("package"), "path", lua.LString(reg_path))

	// register custom loader
	if err := ip.customLoaders.Register(L); err != nil {
		panic(err) // it never occurs
	}

	ip.config.register(L)
}

// set context to internal virtual machine.
// context must not be nil.
// It also starts Watch Dog Timer (WDT) which detect infinite loop
// on script and terminate execution with raise error.
func (ip Interpreter) SetContext(ctx context.Context) {
	ip.vm.SetContext(ctx)

	// start watch dog timer (WDT) and stops immidiately.
	// WDT re-starts at API call from external.
	// NOTE: If Run returns false, context is not propagates to watchDogTimer.
	startTimer := ip.watchDogTimer.Run(ctx)
	if startTimer {
		ip.watchDogTimer.Stop()
	}
}

// Quit quits virtual machine in Interpreter.
// use it for releasing resources.
func (ip *Interpreter) Quit() {
	ip.customLoaders.RemoveAll()
	ip.customLoaders.Unregister(ip.vm)
	ip.vm.Close()
	ip.watchDogTimer.Quit()
	ip.game = nil
	ip.state = nil
}

// DoString runs given src text as script.
func (ip Interpreter) DoString(src string) error {
	if fn, err := ip.vm.LoadString(src); err != nil {
		return err
	} else {
		err = ip.callByParam(fn, lua.MultRet)
		return ip.checkSpecialError(err)
	}
}

// do given script on internal VM.
func (ip *Interpreter) DoFile(file string) error {
	if fn, err := ip.loadFileFromFS(file); err != nil {
		return err
	} else {
		err = ip.callByParam(fn, lua.MultRet)
		return ip.checkSpecialError(err)
	}
}

// lua.LoadFile with filesystem API.
func (ip *Interpreter) loadFileFromFS(file string) (*lua.LFunction, error) {
	fp, err := filesystem.Load(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return ip.vm.Load(fp, file)
}

// do given script file on internal VM with sandbox environment.
// Return data table queried by the dataKey.
func (ip *Interpreter) LoadDataOnSandbox(file, dataKey string) (map[string]string, error) {
	if len(file) == 0 {
		return nil, fmt.Errorf("empty file name")
	}
	if len(dataKey) == 0 {
		return nil, fmt.Errorf("empty data key")
	}

	lfunc, err := ip.loadFileFromFS(file)
	if err != nil {
		return nil, err
	}

	// do script on empty environment for only loading data.
	// TODO use data load enviornment held by the interpreter rather
	// than new empty environment?
	vm := ip.vm
	emptyEnv := vm.NewTable()
	vm.SetFEnv(lfunc, emptyEnv)
	if err := ip.callByParam(lfunc, 0); err != nil {
		return nil, err
	}

	ldata := emptyEnv.RawGetString(dataKey)

	var data = make(map[string]string)
	ltbl, ok := ldata.(*lua.LTable)
	if !ok {
		// data not found
		return data, nil
	}

	// coverts Ltable data into go map
	ltbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString || value.Type() != lua.LTString {
			return
		}
		data[lua.LVAsString(key)] = lua.LVAsString(value)
	})
	return data, nil
}

// return Path of Under Script Dir
func (ip Interpreter) PathOf(file string) string {
	return filepath.Join(ip.config.LoadDir, file)
}

// load all files matched to config pattern.
// it is used for loading user scirpts under specified directory.
func (ip Interpreter) LoadSystem() error {
	path := ip.config.loadPattern()
	files, err := filesystem.Glob(path)
	if err != nil {
		return err
	}
	for _, match := range files {
		if err := ip.DoFile(match); err != nil {
			return err
		}
	}
	return nil
}

func (ip Interpreter) getEraValue(vname string) lua.LValue {
	return ip.eraModule.RawGetString(vname)
}

// VM has given value name in era module?
func (ip Interpreter) HasEraValue(vname string) bool {
	return lua.LVAsBool(ip.getEraValue(vname))
}

// call funtion vname in era module.
func (ip Interpreter) EraCall(vname string) error {
	fn := ip.getEraValue(vname)
	err := ip.callByParam(fn, 0)
	return ip.checkSpecialError(err)
}

func (ip Interpreter) callByParam(fn lua.LValue, nret int, args ...lua.LValue) error {
	if parentCtx := ip.vm.Context(); parentCtx != nil {
		cancelCtx, cancel := context.WithCancel(parentCtx)
		defer cancel()
		ip.vm.SetContext(cancelCtx)
		defer func() { ip.vm.SetContext(parentCtx) }() // recover original state

		ip.watchDogTimer.Reset()
		go func() {
			select {
			case <-ip.watchDogTimer.Expired():
				cancel() // for script execution
			case <-cancelCtx.Done():
				// do nothing
			}
		}()
		defer ip.watchDogTimer.Stop()
	}

	return ip.vm.CallByParam(lua.P{
		Fn:      fn,
		NRet:    nret,
		Protect: true,
	}, args...)
}

// call funtion vname in era module and return a bool value.
func (ip Interpreter) EraCallBool(vname string) (bool, error) {
	fn := ip.getEraValue(vname)
	err := ip.callByParam(fn, 1)
	if err = ip.checkSpecialError(err); err != nil {
		return false, err
	}
	ret := ip.vm.Get(-1)
	ip.vm.Pop(1)
	return lua.LVAsBool(ret), nil
}

// call funtion vname in era module with argument int64,
// and return a bool value.
// To cover all range of int, argument requires int64
func (ip Interpreter) EraCallBoolArgInt(func_name string, arg int64) (bool, error) {
	fn := ip.getEraValue(func_name)
	err := ip.callByParam(fn, 1, lua.LNumber(arg))
	if err = ip.checkSpecialError(err); err != nil {
		return false, err
	}
	ret := ip.vm.Get(-1)
	ip.vm.Pop(1)
	return lua.LVAsBool(ret), nil
}
