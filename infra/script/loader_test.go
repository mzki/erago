package script

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mzki/erago/filesystem"
)

type MockReader struct {
	Ignore    bool
	I         int
	ReadCount int
	Closed    bool
}

var readContent = []byte(`-- lua comment
local g = "hello"
`)

func (m *MockReader) Read(b []byte) (int, error) {
	if m.Ignore {
		return 0, fmt.Errorf("ignore reading")
	}

	i := copy(b, readContent[m.I:])
	m.I += i
	m.ReadCount += 1

	if m.I == len(readContent) {
		return i, io.EOF
	}
	return i, nil
}

func (m *MockReader) Close() error {
	m.I = 0
	m.Closed = true
	return nil
}

type MockLoader struct {
	filesystem.NopPathResolver
	Ignore       bool
	LastLoadName string
	Reader       MockReader
}

func (m *MockLoader) Load(modname string) (io.ReadCloser, error) {
	if m.Ignore {
		return nil, fmt.Errorf("ignore loading %s", modname)
	}
	m.LastLoadName = modname
	return &m.Reader, nil
}

func (m *MockLoader) Exist(modname string) bool {
	return true
}

func TestCallCustomLoader(t *testing.T) {
	const requirePath = "must/not/be/present"
	mockLoader := &MockLoader{}
	globalInterpreter.AddCustomLoader(mockLoader)
	defer globalInterpreter.RemoveCustomLoader(mockLoader)

	err := globalInterpreter.DoString(`require "` + requirePath + `"` + "\n")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(mockLoader.LastLoadName, requirePath) {
		t.Errorf("custom loader is not called. expected to contain: %s, got: %s", requirePath, mockLoader.LastLoadName)
	}

	if mockLoader.Reader.ReadCount == 0 {
		t.Error("custom loader returns reader but it's not used")
	}

	if mockLoader.Reader.Closed == false {
		t.Error("custom loader returns reader but it's not closed")
	}
}

func TestCallCustomLoaderExistFile(t *testing.T) {
	const requirePath = "require"
	ip := newInterpreter()

	// partially disabled default custom loader to use mock loader.
	ip.RemoveCustomLoader(filesystem.Default)
	mockLoader := &MockLoader{}
	ip.AddCustomLoader(mockLoader)

	err := ip.DoString(`require "` + requirePath + `"` + "\n")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(mockLoader.LastLoadName, requirePath) {
		t.Errorf("custom loader is not called. expected to contain: %s, got: %s", requirePath, mockLoader.LastLoadName)
	}

	if mockLoader.Reader.ReadCount == 0 {
		t.Error("custom loader returns reader but it's not used")
	}

	if mockLoader.Reader.Closed == false {
		t.Error("custom loader returns reader but it's not closed")
	}
}

func TestCloseMultipleCustomReader(t *testing.T) {
	const requirePath = "must/not/be/present/multi"
	const multiNum = 5

	mockLoaders := make([]*MockLoader, 0, multiNum)
	for i := 0; i < multiNum; i++ {
		// all of loader loads successfull
		ldr := &MockLoader{Ignore: false}
		// but reader always reads fail...
		ldr.Reader.Ignore = true
		mockLoaders = append(mockLoaders, ldr)
	}

	for _, ldr := range mockLoaders {
		globalInterpreter.AddCustomLoader(ldr)
		defer globalInterpreter.RemoveCustomLoader(ldr)
	}

	err := globalInterpreter.DoString(`require "` + requirePath + `"` + "\n")
	if err == nil {
		t.Fatal("error reader used but no error, something wrong")
	}

	// finally check whether all of readers are properly handled
	for i, mockLoader := range mockLoaders {
		if mockLoader.Reader.ReadCount != 0 {
			t.Errorf("custom loader (%d) returns error reader but it's used", i)
		}

		if mockLoader.Reader.Closed == false {
			t.Errorf("custom loader (%d) returns error reader but it's not closed", i)
		}
	}
}

func TestCustomLoaderWatchFileChange(t *testing.T) {
	const testPath = "./testing/watch_change.lua"
	const requirePath = "watch_change"

	ipr := newInterpreterWithConf(Config{
		LoadDir:                   "testing",
		LoadPattern:               "*",
		CallStackSize:             CallStackSize,
		RegistrySize:              RegistrySize,
		IncludeGoStackTrace:       true,
		InfiniteLoopTimeoutSecond: 0,
		ReloadFileChange:          true,
	})
	defer ipr.Quit()

	fs := filesystem.Desktop
	err := ipr.AddCustomLoader(fs)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := ipr.RemoveCustomLoader(fs)
		if err != nil {
			t.Fatal(err)
		}
	}()

	const beforeCode = `
return {
	func1 = function() end
}
`
	const afterCode = `
return {
	func1 = "func1_string",
	func2 = function() end
}
`
	if err := os.WriteFile(testPath, []byte(beforeCode), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testPath) // to remove dust file.

	if err := ipr.DoString(fmt.Sprintf(`watch_change = require "%s"`, requirePath)); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(testPath, []byte(afterCode), 0755); err != nil {
		t.Fatal(err)
	}
	// reflect change by call era function, so it places first
	doSrc := fmt.Sprintf(`
era.printl "reflect change"
-- check existant module has be reflected reload result.
assert(type(watch_change.func1), "string")
assert(type(watch_change.func2), "function")
-- check require module has be reflected reload result.
local watch_change2 = require "%s"
assert(type(watch_change2.func1), "string")
assert(type(watch_change2.func2), "function")
`, requirePath)
	if err := ipr.DoString(doSrc); err != nil {
		t.Fatal(err)
	}
}

func TestCustomLoaderRequireUpperBaseDirFailed(t *testing.T) {
	const testPath = "./testing/../testing2/invalid_basedir.lua"
	const requirePath = "../invalid_basedir"
	const upperTestDir = "./testing2"

	ipr := newInterpreterWithConf(Config{
		LoadDir:                   "testing",
		LoadPattern:               "*",
		CallStackSize:             CallStackSize,
		RegistrySize:              RegistrySize,
		IncludeGoStackTrace:       true,
		InfiniteLoopTimeoutSecond: 0,
		ReloadFileChange:          false,
	})
	defer ipr.Quit()

	// create partially exist script file
	if err := os.MkdirAll(upperTestDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, []byte(`era.print "hello"`), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(upperTestDir)

	fs := filesystem.Desktop
	err := ipr.AddCustomLoader(fs)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := ipr.RemoveCustomLoader(fs)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := ipr.DoString(fmt.Sprintf(`require "%s"`, requirePath)); err == nil {
		t.Fatalf("call require with upper base directory, breaks sandbox, but no error")
	} else {
		t.Log(err)
	}
}
