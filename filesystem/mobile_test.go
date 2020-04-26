package filesystem

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

const MobileLoadSource = "mobile_test.go"

func TestMobileLoader(t *testing.T) {
	var FS FileSystem = Mobile
	reader, err := FS.Load(MobileLoadSource)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	content, err := ioutil.ReadAll(io.LimitReader(reader, 64))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "package filesystem") {
		t.Errorf("can not read first line of this file")
	}
}

func TestAbsPathFileSystemResolvePath(t *testing.T) {
	parentPath, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	filesystem := &AbsPathFileSystem{CurrentDir: parentPath}
	fpath, err := filesystem.ResolvePath(MobileLoadSource)
	if err != nil {
		t.Fatal(err)
	}

	if !filepath.IsAbs(fpath) {
		t.Fatalf("ResolvePath() returns a not absolute path. got: %v", fpath)
	}

	// case2: returns absolute path even if empty current dir.
	{
		EmptyFS := &AbsPathFileSystem{CurrentDir: ""}
		fpath, err := EmptyFS.ResolvePath(MobileLoadSource)
		if err != nil {
			t.Fatal(err)
		}

		if !filepath.IsAbs(fpath) {
			t.Fatalf("EmptyFS.ResolvePath() returns a not absolute path. got: %v", fpath)
		}
	}

	// case3: changing abspath by user
	{
		const SubDirName = "subdir"
		FS := &AbsPathFileSystem{CurrentDir: filepath.Join(parentPath, SubDirName)}
		fpath, err := FS.ResolvePath(MobileLoadSource)
		if err != nil {
			t.Fatal(err)
		}

		if !filepath.IsAbs(fpath) {
			t.Fatalf("EmptyFS.ResolvePath() returns a not absolute path. got: %v", fpath)
		}
		dir, base := filepath.Split(fpath)
		if got := filepath.Base(dir); got != SubDirName {
			t.Errorf("user specified subdir missing: got %v, expect %v", got, SubDirName)
		}
		if base != MobileLoadSource {
			t.Errorf("user specified base name missing: got %v, expect %v", base, MobileLoadSource)
		}
	}

	// finally test reading content
	reader, err := filesystem.Load(MobileLoadSource)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	content, err := ioutil.ReadAll(io.LimitReader(reader, 64))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "package filesystem") {
		t.Errorf("can not read first line of this file")
	}
}
