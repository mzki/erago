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
	if err := <-errCh; err != context.Canceled {
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
	if err := <-errCh; err != ErrWatchDogTimerExpired {
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
	if err := <-errCh; err != context.Canceled {
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
