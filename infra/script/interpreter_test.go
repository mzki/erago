package script

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/stub"
)

var globalInterpreter *Interpreter

func TestMain(m *testing.M) {
	globalInterpreter = newInterpreter()
	defer globalInterpreter.Quit()

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
	state, err := stub.GetGameState()
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
	} {
		if err := ip.DoFile(filepath.Join(scriptDir, file)); err != nil {
			t.Error(err)
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
	ip := newInterpreter()
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
