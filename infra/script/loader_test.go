package script

import (
	"fmt"
	"io"
	"testing"
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

	if mockLoader.LastLoadName != requirePath {
		t.Error("custom loader is not called")
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
