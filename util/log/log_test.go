package log

import (
	"bytes"
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
