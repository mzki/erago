package filesystem

import (
	"io"
	"io/fs"
	"strings"
	"testing"
)

const OSLoadSource = "desktop_test.go"

func TestOSLoader(t *testing.T) {
	var FS FileSystem = Desktop
	reader, err := FS.Load(OSLoadSource)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	content, err := io.ReadAll(io.LimitReader(reader, 64))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "package filesystem") {
		t.Errorf("can not read first line of this file")
	}
}

func TestOSLoaderSizeExceed(t *testing.T) {
	osldr := &OSFileSystem{MaxFileSize: 10}
	reader, err := osldr.Load(OSLoadSource)
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

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != source {
		t.Errorf("different content")
	}
}

func TestOSOpenFS(t *testing.T) {
	var fsys FileSystem = &OSFileSystem{MaxFileSize: DefaultMaxFileSize}
	fsfs, ok := fsys.(fs.FS)
	if !ok {
		t.Fatal("OSFileSystem does not implement fs.FS interface")
	}

	const fpath = "./desktop_test.go"
	file, err := fsfs.Open(fpath)
	if err != nil {
		t.Fatalf("failed to Open: %v", fpath)
	}
	defer file.Close()
}
