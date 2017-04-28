package script

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"local/erago/flow"
	"local/erago/stub"
)

var globalInterpreter *Interpreter

func TestMain(m *testing.M) {
	globalInterpreter = newInterpreter()
	defer globalInterpreter.Quit()

	os.Exit(m.Run())
}

const scriptDir = "testing"

func newConfig() Config {
	c := NewConfig("./")
	c.LoadDir = scriptDir
	return c
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
		if err := ip.DoFile(filepath.Join("./testing/", file)); err != nil {
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
		{"testquit", flow.ErrorQuit},
		{"testgoto", flow.ErrorSceneNext},
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

func BenchmarkScript(b *testing.B) {
	// TODO
}
