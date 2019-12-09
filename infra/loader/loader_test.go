package loader

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestOSLoader(t *testing.T) {
	reader, err := OS.Load("loader_test.go")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	content, err := ioutil.ReadAll(io.LimitReader(reader, 64))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "package loader") {
		t.Errorf("can not read first line of this file")
	}
}

func TestOSLoaderSizeExceed(t *testing.T) {
	osldr := &OSLoader{MaxFileSize: 10}
	reader, err := osldr.Load("loader_test.go")
	if err == nil {
		reader.Close()
		t.Fatalf("with max file size 10 byte, but loader reports no error")
	}
}

func TestStringLoader(t *testing.T) {
	const source = "abcdefghijklmnopqrstuvwxyz"
	reader, err := String.Load(source)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != source {
		t.Errorf("different content")
	}
}
