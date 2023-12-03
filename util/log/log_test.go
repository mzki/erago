package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := New(buf, "", LstdFlags)

	logger.Debug("debug text")
	if bs := buf.Bytes(); len(bs) != 0 {
		t.Error("On InfoLevel, Debug outputs some text")
	}

	buf.Reset()
	logger.Info("info text")
	if bs := buf.Bytes(); len(bs) == 0 {
		t.Error("On InfoLevel, Info outputs nothing")
	}

	logger.SetLevel(DebugLevel)

	buf.Reset()
	logger.Info("info text")
	if bs := buf.Bytes(); len(bs) == 0 {
		t.Error("On DebugLevel, Info outputs nothing")
	}

	buf.Reset()
	logger.Debug("debug text")
	if bs := buf.Bytes(); len(bs) == 0 {
		t.Error("On DebugLevel, Debug outputs nothing")
	}
}

func TestLoggerError(t *testing.T) {
	buf := new(bytes.Buffer)
	limitW := LimitWriter(buf, int64(len("2024/12/24 12:00:00 ")+len(DebugPrefix)+1))
	logger := New(limitW, "", LstdFlags)
	logger.SetLevel(DebugLevel)

	if err := logger.Err(); err != nil {
		t.Errorf("newly logger should not have any error, but got error: %v", err)
	}

	logger.Debug("")
	if err := logger.Err(); err != nil {
		t.Logf("Written log: %v", buf.String())
		t.Fatalf("Debug must be succeeded, but got error: %v", err)
	}

	logger.Debug("some text")
	if err := logger.Err(); !errors.Is(err, io.EOF) {
		t.Logf("Written log: %v", buf.String())
		t.Fatalf("Debug reaches maximum bytes limit and must be EOF error, but got other error: %v", err)
	}

	buf.Reset()
	logger.SetOutput(LimitWriter(buf, 1024))
	logger.Info("info text")
	if err := logger.Err(); err != nil {
		t.Fatalf("After buffer and limit is reset, Info should be suceeded and reutrn nil, but got error: %v", err)
	}
}

func ExampleLogger() {
	buf := new(bytes.Buffer)
	logger := New(buf, "", Lshortfile)
	logger.SetLevel(DebugLevel)

	logger.Debug("debug text")
	logger.Info("info text")

	SetOutput(buf)
	SetLevel(DebugLevel)
	SetFlags(Lshortfile)

	Debug("debug from function")
	Info("info from function")
	fmt.Println(buf.String())

	// Output:
	// log_test.go:76: DEBUG: debug text
	// log_test.go:77: info text
	// log_test.go:83: DEBUG: debug from function
	// log_test.go:84: info from function
}
