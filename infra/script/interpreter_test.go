package script

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/stub"
	lua "github.com/yuin/gopher-lua"
)

var (
	globalInterpreter          *Interpreter
	globalOlderDataInterpreter *Interpreter
)

func TestMain(m *testing.M) {
	globalInterpreter = newInterpreter()
	defer globalInterpreter.Quit()
	globalOlderDataInterpreter = newOlderDataInterpreter()
	defer globalOlderDataInterpreter.Quit()

	os.Exit(m.Run())
}

const scriptDir = "testing"

func newConfig() Config {
	return Config{
		LoadDir:             scriptDir,
		LoadPattern:         LoadPattern,
		CallStackSize:       CallStackSize,
		RegistrySize:        RegistrySize,
		IncludeGoStackTrace: true,
	}
}

func newInterpreter() *Interpreter {
	return newInterpreterWithConf(newConfig())
}

func newInterpreterWithConf(conf Config) *Interpreter {
	state, err := stub.GetGameState()
	if err != nil {
		panic(err)
	}
	return NewInterpreter(state, stub.NewScriptGameController(), conf)
}

func newOlderDataInterpreter() *Interpreter {
	state, err := stub.GetOlderGameState()
	if err != nil {
		panic(err)
	}
	return NewInterpreter(state, stub.NewScriptGameController(), newConfig())
}

func newInterpreterAndInputQueuer() (*Interpreter, InputQueuer) {
	state, err := stub.GetGameState()
	if err != nil {
		panic(err)
	}
	ctrlr := stub.NewScriptGameController()
	inputQ := stub.GetInputQueue(ctrlr)
	return NewInterpreter(state, ctrlr, newConfig()), inputQ
}

func TestInterpreterPath(t *testing.T) {
	ip := globalInterpreter

	got := ip.PathOf(ip.config.LoadPattern)
	expect := filepath.Join(ip.config.LoadDir, ip.config.LoadPattern)
	if got != expect {
		t.Log(got, expect)
		t.Errorf("got: %v expect %v", got, expect)
	}
}

func TestInterpreterDoFiles(t *testing.T) {
	ip := globalInterpreter

	for _, file := range []string{
		"era_module.lua",
		"preloads.lua",
		"require.lua",
		"lua_function.lua",
	} {
		if err := ip.DoFile(filepath.Join(scriptDir, file)); err != nil {
			t.Error(err)
		}
	}

}

func TestInterpreterOlderDataLoadTest(t *testing.T) {
	ip := globalInterpreter
	olderIpr := globalOlderDataInterpreter

	// Do on the older data context
	const olderSubDir = "on_olderdata"
	for _, file := range []string{
		"savedata.lua",
	} {
		if err := olderIpr.DoFile(filepath.Join(scriptDir, olderSubDir, file)); err != nil {
			t.Fatal(err)
		}
	}
	// Do on the newer data context
	for _, file := range []string{
		"loaddata.lua",
	} {
		if err := ip.DoFile(filepath.Join(scriptDir, file)); err != nil {
			t.Error(err)
		}
	}
}

func TestInterpreterLoadDataOnSandbox(t *testing.T) {
	ip := globalInterpreter

	// exists
	for _, file := range []string{
		"data_on_sandbox.lua",
	} {
		const key = "DATA"
		data, err := ip.LoadDataOnSandbox(filepath.Join(scriptDir, file), key)
		if err != nil {
			t.Fatal(err)
		}
		if data == nil {
			t.Fatalf("data name %s should exist but not on file %s", key, file)
		}
		for _, field := range []string{
			"DATA1",
			"data2",
			"Data3",
		} {
			text, ok := data[field]
			if !ok {
				t.Errorf("%s: data name %s should contain key %s for string value", file, key, field)
			}
			if len(text) == 0 {
				t.Errorf("%s: data name %s should contain valid string for key %s, got: %v", file, key, field, text)
			}
		}
	}

	// file eixsts but data key not exists
	for _, file := range []string{
		"data_on_sandbox.lua",
	} {
		const key = "__NO_DATA"
		data, err := ip.LoadDataOnSandbox(filepath.Join(scriptDir, file), key)
		if err != nil {
			t.Errorf("%s: querying no exist data name %s should not error, but %v", file, key, err)
		}
		if len(data) != 0 {
			t.Errorf("%s: no exist data name %s should empty but containing something: %#v", file, key, data)
		}
	}

	// empty_key should error
	for _, file := range []string{
		"data_on_sandbox.lua",
	} {
		const key = ""
		_, err := ip.LoadDataOnSandbox(filepath.Join(scriptDir, file), key)
		if err == nil {
			t.Errorf("empty key should not acceptable but no error!!!")
		}
	}

	// file not eixsts
	for _, file := range []string{
		"__not_exists_data_on_sandbox.lua",
	} {
		const key = "DATA"
		_, err := ip.LoadDataOnSandbox(filepath.Join(scriptDir, file), key)
		if err == nil {
			t.Errorf("file name isnt exists but no error!!!")
		}
	}
}

func TestInterpreterLoadFileFromFS(t *testing.T) {
	ip := globalInterpreter

	// exists
	for _, file := range []string{
		"data_on_sandbox.lua",
	} {
		lfunc, err := ip.loadFileFromFS(filepath.Join(scriptDir, file))
		if err != nil {
			t.Fatal(err)
		}
		if err := ip.vm.CallByParam(lua.P{
			Fn:      lfunc,
			NRet:    0,
			Protect: true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	backupFS := filesystem.Default
	defer func() { filesystem.Default = backupFS }()

	filesystem.Default = &filesystem.AbsPathFileSystem{}

	// exists
	for _, file := range []string{
		"data_on_sandbox.lua",
	} {
		lfunc, err := ip.loadFileFromFS(filepath.Join(scriptDir, file))
		if err != nil {
			t.Fatal(err)
		}
		if err := ip.vm.CallByParam(lua.P{
			Fn:      lfunc,
			NRet:    0,
			Protect: true,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInterpreterLoadSystemRelativePath(t *testing.T) {
	const fileName = "init_loadsystem"
	const testPath = "testing/" + fileName + ".lua"
	ip := newInterpreterWithConf(Config{
		LoadDir:             "testing",
		LoadPattern:         fileName + ".lua",
		CallStackSize:       CallStackSize,
		RegistrySize:        RegistrySize,
		IncludeGoStackTrace: true,
	})
	defer ip.Quit()

	if err := os.WriteFile(testPath, []byte(`era.twait(100)`), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testPath)

	oldFS := filesystem.Default
	defer func() { filesystem.Default = oldFS }()
	filesystem.Default = filesystem.Desktop
	if err := ip.LoadSystem(); err != nil {
		t.Errorf("Failed to LoadSystem: %v", err)
	}
}

func TestInterpreterLoadSystemAbsPath(t *testing.T) {
	const fileName = "init_loadsystem"
	const testPath = "testing/" + fileName + ".lua"
	ip := newInterpreterWithConf(Config{
		LoadDir:             "testing",
		LoadPattern:         fileName + ".lua",
		CallStackSize:       CallStackSize,
		RegistrySize:        RegistrySize,
		IncludeGoStackTrace: true,
	})
	defer ip.Quit()

	if err := os.WriteFile(testPath, []byte(`era.twait(100)`), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testPath)

	oldFS := filesystem.Default
	defer func() { filesystem.Default = oldFS }()
	absDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	filesystem.Default = &filesystem.AbsPathFileSystem{
		CurrentDir: absDir,
		Backend:    filesystem.Desktop,
	}
	if err := ip.LoadSystem(); err != nil {
		t.Errorf("Failed to LoadSystem: %v", err)
	}
}

func TestInterpreterLoadSystemNotFound(t *testing.T) {
	const fileName = "load_system_not_found"
	ip := newInterpreterWithConf(Config{
		LoadDir:             "testing",
		LoadPattern:         fileName + ".lua",
		CallStackSize:       CallStackSize,
		RegistrySize:        RegistrySize,
		IncludeGoStackTrace: true,
	})
	defer ip.Quit()

	err := ip.LoadSystem()
	if notFoundErr, ok := err.(*LoadPatternNotFoundError); !ok {
		t.Errorf("LoadSystem should return LoadPatternNotFoundError but got: %v", err)
	} else {
		t.Logf("Intended Erorr: %v", notFoundErr)
	}
}

func TestInterpreterLoadSystemUpperLoadDirFail(t *testing.T) {
	const upperTestDir = "testing2"
	const testPath = "testing2/invalid_basedir.lua"
	ip := newInterpreterWithConf(Config{
		LoadDir:             "testing",
		LoadPattern:         "../" + testPath,
		CallStackSize:       CallStackSize,
		RegistrySize:        RegistrySize,
		IncludeGoStackTrace: true,
	})
	defer ip.Quit()

	// create partially exist script file
	if err := os.MkdirAll(upperTestDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, []byte(`era.printl "hello"`), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(upperTestDir)

	if err := ip.LoadSystem(); err == nil {
		t.Fatal("LoadSystem should fail due to access upper LoadDir, breaking rule, but not error")
	} else {
		t.Logf("Intended Error: %v", err)
	}
}

func TestInterpreterEraCall(t *testing.T) {
	ip := globalInterpreter

	for _, file := range []string{
		"eracall.lua",
	} {
		if err := ip.DoFile(filepath.Join(scriptDir, file)); err != nil {
			t.Error(err)
		}
	}

	// call era.XXX function
	err := ip.EraCall("testing")
	if err != nil {
		t.Error(err)
	}

	b, err := ip.EraCallBool("testing_bool")
	if !b {
		t.Error("must be true but false")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestInterpreterSpecialErrors(t *testing.T) {
	ip := globalInterpreter

	if err := ip.DoFile(filepath.Join(scriptDir, "quit.lua")); err != nil {
		t.Fatal(err)
	}
	for _, testcase := range []struct {
		FuncName string
		Error    error
	}{
		{"testquit", scene.ErrorQuit},
		{"testgoto", scene.ErrorSceneNext},
		{"testlongreturn", nil},
		{"testpcall_quit", scene.ErrorQuit},
		{"testpcall_gotoNextScene", scene.ErrorSceneNext},
		{"testpcall_longReturn", nil},
		{"testpcall_something", nil},
		{"testxpcall_quit", scene.ErrorQuit},
		{"testxpcall_gotoNextScene", scene.ErrorSceneNext},
		{"testxpcall_longReturn", nil},
		{"testxpcall_something", nil},
		{"testxpcall_something2", nil},
		{"testxpcall_something_error_handler", nil},
	} {
		if err := ip.EraCall(testcase.FuncName); err != testcase.Error {
			if err == nil {
				err = fmt.Errorf("error(nil) is invalid in this test")
			}
			t.Errorf("in func era.%v, Error expect: %v, got: %v", testcase.FuncName, testcase.Error, err)
		}
	}
}

func TestInterpreterContextCancel(t *testing.T) {
	t.Parallel()
	conf := newConfig()
	conf.InfiniteLoopTimeoutSecond = 10

	ip := newInterpreterWithConf(conf)
	defer ip.Quit()
	ctx, cancel := context.WithCancel(context.Background())
	ip.SetContext(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- ip.DoFile(filepath.Join(scriptDir, "infinite_loop.lua"))
		close(errCh)
	}()

	time.Sleep(1 * time.Millisecond)
	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestInterpreterWatchDogTimerCancel(t *testing.T) {
	t.Parallel()
	conf := newConfig()
	conf.InfiniteLoopTimeoutSecond = 1

	ip := newInterpreterWithConf(conf)
	defer ip.Quit()
	ctx := context.Background()
	ip.SetContext(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- ip.DoFile(filepath.Join(scriptDir, "infinite_loop.lua"))
		close(errCh)
	}()

	time.Sleep(2 * time.Second)
	if err := <-errCh; !errors.Is(err, ErrWatchDogTimerExpired) {
		t.Fatal(err)
	}
}

func TestInterpreterWatchDogTimerNotExpired(t *testing.T) {
	t.Parallel()
	conf := newConfig()
	conf.InfiniteLoopTimeoutSecond = 1

	ip := newInterpreterWithConf(conf)
	defer ip.Quit()
	ctx, cancel := context.WithCancel(context.Background())
	ip.SetContext(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- ip.DoFile(filepath.Join(scriptDir, "infinite_loop_ok.lua"))
		close(errCh)
	}()

	time.Sleep(time.Duration(conf.InfiniteLoopTimeoutSecond+1) * time.Second)
	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestInterpreterWatchDogTimerNotExpiredAfterInitialization(t *testing.T) {
	t.Parallel()

	conf := newConfig()
	conf.InfiniteLoopTimeoutSecond = 1

	ip := newInterpreterWithConf(conf)
	defer ip.Quit()
	ctx := context.Background()
	ip.SetContext(ctx)

	// large sleep than InfiniteLoopTimeoutSecond
	// may be watch dog timer is expired if something wrong.
	time.Sleep(2 * time.Second)

	errCh := make(chan error, 1)
	go func() {
		errCh <- ip.DoFile(filepath.Join(scriptDir, "lua_function.lua"))
		close(errCh)
	}()

	if err := <-errCh; err == ErrWatchDogTimerExpired {
		t.Fatal(errors.New("No request WatchDogTimer, but got it"))
	}
}

func TestInterpreterTestingLibs(t *testing.T) {
	ip, inputQ := newInterpreterAndInputQueuer()
	ip.OpenTestingLibs(inputQ)

	for _, testcase := range []struct {
		FileName string
		Error    error
	}{
		{"era_input_queue.lua", nil},
	} {
		if err := ip.DoFile(filepath.Join(scriptDir, testcase.FileName)); err != testcase.Error {
			t.Errorf("in %v, Error expect: %v, got: %v", testcase.FileName, testcase.Error, err)
		}
	}
}

func TestInterpreterTimeoutByContext(t *testing.T) {
	ip := newInterpreter()

	for _, testcase := range []struct {
		FileName string
		Error    error
	}{
		{"infinite_loop_ok.lua", context.DeadlineExceeded},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		ip.SetContext(ctx)
		if err := ip.DoFile(filepath.Join(scriptDir, testcase.FileName)); !errors.Is(err, testcase.Error) {
			t.Errorf("in %v, Error expect: %v, got: %v", testcase.FileName, testcase.Error, err)
		}
		cancel()
	}
}

type gameControllerWithInputError struct {
	GameController
}

var errInputErrorSomehow = errors.New("input error some how")

func (gameControllerWithInputError) RawInput() (string, error) { return "", errInputErrorSomehow }
func (gameControllerWithInputError) RawInputWithTimeout(context.Context, time.Duration) (string, error) {
	return "", errInputErrorSomehow
}
func (gameControllerWithInputError) Command() (string, error) { return "", errInputErrorSomehow }
func (gameControllerWithInputError) CommandWithTimeout(context.Context, time.Duration) (string, error) {
	return "", errInputErrorSomehow
}
func (gameControllerWithInputError) CommandNumber() (int, error) { return 0, errInputErrorSomehow }
func (gameControllerWithInputError) CommandNumberWithTimeout(context.Context, time.Duration) (int, error) {
	return 0, errInputErrorSomehow
}
func (gameControllerWithInputError) CommandNumberRange(ctx context.Context, min, max int) (int, error) {
	return 0, errInputErrorSomehow
}
func (gameControllerWithInputError) CommandNumberSelect(context.Context, ...int) (int, error) {
	return 0, errInputErrorSomehow
}
func (gameControllerWithInputError) Wait() error { return errInputErrorSomehow }
func (gameControllerWithInputError) WaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	return errInputErrorSomehow
}

func newInterpreterWithInputError() *Interpreter {
	state, err := stub.GetGameState()
	if err != nil {
		panic(err)
	}
	ctrlr := stub.NewScriptGameController()
	return NewInterpreter(state, &gameControllerWithInputError{ctrlr}, newConfig())
}

func TestInterpreterInputErrorSomehow(t *testing.T) {
	ip := newInterpreterWithInputError()

	for _, testcase := range []struct {
		Src   string
		Error error
	}{
		{`era.input()`, errInputErrorSomehow},
		{`era.tinput(1000)`, errInputErrorSomehow},
		{`era.rawInput()`, errInputErrorSomehow},
		{`era.trawInput(1000)`, errInputErrorSomehow},
		{`era.inputNum()`, errInputErrorSomehow},
		{`era.tinputNum(1000)`, errInputErrorSomehow},
		{`era.wait()`, errInputErrorSomehow},
		{`era.twait(1000)`, errInputErrorSomehow},
		{`era.inputRange(0, 10)`, errInputErrorSomehow},
		{`era.inputSelect(0, 1, 2)`, errInputErrorSomehow},
	} {
		if err := ip.DoString(testcase.Src); !errors.Is(err, testcase.Error) {
			t.Errorf("in %q, Error expect: %v, got: %v", testcase.Src, testcase.Error, err)
		}
	}
}

func TestInterpreterInternalErrorInPcall(t *testing.T) {
	ip := newInterpreterWithInputError()

	for _, testcase := range []struct {
		Src      string
		NotError error
	}{
		{`pcall(era.input); error("something wrong except input error")`, errInputErrorSomehow},
		{`xpcall(era.input); error("something wrong except input error")`, errInputErrorSomehow},
		{`xpcall(era.input, era.input); error("something wrong except input error")`, errInputErrorSomehow},
	} {
		if err := ip.DoString(testcase.Src); errors.Is(err, testcase.NotError) {
			t.Errorf("in %q, Error NOT expect: %v, got: %v", testcase.Src, testcase.NotError, err)
		}
	}

	for _, testcase := range []struct {
		Src   string
		Error error
	}{
		{`pcall(error, "something wrong except input error"); era.input()`, errInputErrorSomehow},
		{`xpcall(function() error("something wrong except input error") end); era.input()`, errInputErrorSomehow},
		{`xpcall(function() error("fn") end, function() error("something wrong except input error") end); era.input()`, errInputErrorSomehow},
	} {
		if err := ip.DoString(testcase.Src); !errors.Is(err, testcase.Error) {
			t.Errorf("in %q, Error expect: %v, got: %v", testcase.Src, testcase.Error, err)
		}
	}

}

// ------------------------
// benchmarking
// ------------------------

func newBenchInterpreter(b *testing.B) *Interpreter {
	ip := newInterpreter()
	if err := ip.DoFile(filepath.Join(scriptDir, "bench.lua")); err != nil {
		ip.Quit()
		b.Fatal(err)
	}
	return ip
}

func BenchmarkScriptWithoutContext(b *testing.B) {
	ip := newBenchInterpreter(b)
	defer ip.Quit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ip.EraCall("bench1"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScriptWithContext(b *testing.B) {
	ip := newBenchInterpreter(b)
	defer ip.Quit()

	ip.SetContext(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ip.EraCall("bench2"); err != nil {
			b.Fatal(err)
		}
	}
}
