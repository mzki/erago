package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const MobileLoadSource = "mobile_test.go"

func TestAbsDirFileSystem(t *testing.T) {
	dirPath, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	dirFs := AbsDirFileSystem(dirPath)
	reader, err := dirFs.Load(MobileLoadSource)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
}

func TestAbsDirFileSystemAtUnknownLocation(t *testing.T) {
	const unknownDir = "/path/to/unknown"
	dirFs := AbsDirFileSystem(unknownDir)
	reader, err := dirFs.Load(MobileLoadSource)
	if err == nil {
		defer reader.Close()
		t.Fatalf("Expected to raise some error for unknown path(%v), but no error", filepath.Join(unknownDir, MobileLoadSource))
	}
	// do not need to call reader.Close as err is not nil here.
}

func TestMobileLoader(t *testing.T) {
	var FS FileSystem = Mobile
	reader, err := FS.Load(MobileLoadSource)
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
		var EmptyFS PathResolver = &AbsPathFileSystem{CurrentDir: ""}
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

	content, err := io.ReadAll(io.LimitReader(reader, 64))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(content), "package filesystem") {
		t.Errorf("can not read first line of this file")
	}
}

func TestAbsFileSystemOpenFS(t *testing.T) {
	parentPath, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	var filesystem FileSystem = &AbsPathFileSystem{CurrentDir: parentPath}
	fsfs, ok := filesystem.(fs.FS)
	if !ok {
		t.Fatal("AbsPathFileSystem should support fs.FS interface, but not")
	}

	const fpath = "./mobile_test.go"
	file, err := fsfs.Open(fpath)
	if err != nil {
		t.Fatalf("failed to Open: %v", fpath)
	}
	defer file.Close()
}

var errNotSupportedForTest = fmt.Errorf("not supported")

type emptyFileSystem struct{}

func (emptyFileSystem) Load(fpath string) (io.ReadCloser, error) { return nil, errNotSupportedForTest }
func (emptyFileSystem) Store(fpath string) (io.WriteCloser, error) {
	return nil, errNotSupportedForTest
}
func (emptyFileSystem) Exist(fpath string) bool { return false }

func TestAbsFileSystemOpenFSNotSupport(t *testing.T) {
	parentPath, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	var filesystem FileSystem = &AbsPathFileSystem{
		CurrentDir: parentPath,
		Backend:    &emptyFileSystem{},
	}
	fsfs, ok := filesystem.(fs.FS)
	if !ok {
		t.Fatal("AbsPathFileSystem should support fs.FS interface, but not")
	}

	const fpath = "./mobile_test.go"
	file, err := fsfs.Open(fpath)
	if err == nil {
		file.Close()
		t.Fatalf("expected to have error but got nil for path: %v", fpath)
	}
	t.Logf("expected err content: %v", err)
}

func TestAbsPathFileSystem_Open(t *testing.T) {
	currentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		CurrentDir string
		Backend    FileSystem
	}
	type args struct {
		fpath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{currentDir, Desktop}, args{"mobile_test.go"}, false},
		{"not found", fields{currentDir, Desktop}, args{"path/to/not-found"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absfs := &AbsPathFileSystem{
				CurrentDir: tt.fields.CurrentDir,
				Backend:    tt.fields.Backend,
			}
			gotFile, err := absfs.Open(tt.args.fpath)
			if (err != nil) != tt.wantErr {
				t.Errorf("AbsPathFileSystem.Open() error = %v, wantErr %v", err, tt.wantErr)
				gotFile.Close()
				return
			}
			if err == nil {
				defer gotFile.Close()
			}
		})
	}
}
