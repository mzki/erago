package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const testFileInteropTest = "./interop_test.go"

func TestFromFS(t *testing.T) {
	type args struct {
		fsys fs.FS
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal case", args{os.DirFS("./")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsystem := FromFS(tt.args.fsys)
			if ok := fsystem.Exist(testFileInteropTest); !ok {
				t.Errorf("%v should exist, but fsys reports not exist.", testFileInteropTest)
			}
		})
	}
}

func TestInteropFileSystem_Load(t *testing.T) {
	type fields struct {
		Backend fs.FS
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal case", fields{os.DirFS("./")}, args{testFileInteropTest}, false},
		{"non-exist case", fields{os.DirFS("./")}, args{"path/to/not-exist.file"}, true},
		{"non-exist case2", fields{os.DirFS("./not-exist-dir")}, args{testFileInteropTest}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			got, err := ifs.Load(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteropFileSystem.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				defer got.Close()
			}
			if tt.wantErr || err != nil {
				return
			}
			// try read
			bs := make([]byte, 8)
			if _, err := got.Read(bs); (err != nil) != tt.wantErr {
				t.Errorf("try read data failed. error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInteropFileSystem_Store(t *testing.T) {
	tempDir := t.TempDir()
	type fields struct {
		Backend fs.FS
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"fs.FS backend will fail for Store", fields{os.DirFS(tempDir)}, args{"newfile.txt"}, true},
		{"nested-dir-not-found", fields{os.DirFS(tempDir)}, args{"test/newfile.txt"}, true},
		{"AbsPathFileSystem backend will success for Store", fields{AbsDirFileSystem(tempDir)}, args{"test/newfile2.txt"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			got, err := ifs.Store(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteropFileSystem.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				defer got.Close()
			}
			if tt.wantErr || err != nil {
				return
			}
			// try write
			bs := []byte{0x00, 0x01, 0x02, 0x03}
			if _, err := got.Write(bs); (err != nil) != tt.wantErr {
				t.Errorf("try write failed. error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInteropFileSystem_Exist(t *testing.T) {
	type fields struct {
		Backend fs.FS
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"normal case", fields{os.DirFS("./")}, args{testFileInteropTest}, true},
		{"non-exist case", fields{os.DirFS("./")}, args{"path/to/not-exist.file"}, false},
		{"non-exist case2", fields{os.DirFS("./not-exist-dir")}, args{testFileInteropTest}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			if got := ifs.Exist(tt.args.path); got != tt.want {
				t.Errorf("InteropFileSystem.Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteropFileSystem_Open(t *testing.T) {
	type fields struct {
		Backend fs.FS
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal case", fields{os.DirFS("./")}, args{testFileInteropTest}, false},
		{"non-exist case", fields{os.DirFS("./")}, args{"path/to/not-exist.file"}, true},
		{"non-exist case2", fields{os.DirFS("./not-exist-dir")}, args{testFileInteropTest}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			got, err := ifs.Open(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteropFileSystem.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				defer got.Close()
			}
			if tt.wantErr {
				return
			}
		})
	}
}

type globDirFS struct {
	fs.FS
	dir string
}

func newGlobDirFS(dir string) *globDirFS {
	return &globDirFS{os.DirFS(dir), dir}
}

func (fsys *globDirFS) Glob(pattern string) ([]string, error) {
	pattern = filepath.Join(fsys.dir, pattern)
	return filepath.Glob(pattern)
}

func TestInteropFileSystem_Glob(t *testing.T) {
	absCurrentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		Backend fs.FS
	}
	type args struct {
		pattern string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{"normal", fields{newGlobDirFS("./")}, args{"interop_test.go"}, []string{"interop_test.go"}, false},
		{"error glob not support", fields{&AbsPathFileSystem{absCurrentDir, &emptyFileSystem{}}}, args{"interop_test.go"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			got, err := ifs.Glob(tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteropFileSystem.Glob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteropFileSystem.Glob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteropFileSystem_ResolvePath(t *testing.T) {
	absCurrentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		Backend fs.FS
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"normal abs path resolved", fields{&AbsPathFileSystem{absCurrentDir, &emptyFileSystem{}}}, args{"interop_test.go"}, filepath.Join(absCurrentDir, "interop_test.go"), false},
		{"normal but not resolved", fields{newGlobDirFS("./")}, args{"interop_test.go"}, "interop_test.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifs := &InteropFileSystem{
				Backend: tt.fields.Backend,
			}
			got, err := ifs.ResolvePath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteropFileSystem.ResolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InteropFileSystem.ResolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
