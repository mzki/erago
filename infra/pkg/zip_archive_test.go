package pkg

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/mzki/erago/filesystem"
)

var testZipArchiveFS fs.ReadDirFS = fstest.MapFS{
	"test.txt":              {Data: []byte("test-txt")},
	"test-dir/test2.txt":    {Data: []byte("test-dir-test2-txt")},
	"test-dir/test3.txt":    {Data: []byte("test-dir-test3-txt")},
	"test-dir/マルチバイト文字.txt": {Data: []byte("マルチバイト文字-txt")},
}

//go:embed testdata/*
var testdataFS embed.FS

func TestArchiveAsZip(t *testing.T) {
	tempDir := t.TempDir()
	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		outFsys       filesystem.FileSystem
		outPathPrefix string
		srcFsys       fs.FS
		targetFiles   []string
	}
	tests := []struct {
		name           string
		args           args
		wantOutputPath string
		wantErr        bool
	}{
		/*
			{
				name: "generate-golden",
				args: args{
					&filesystem.OSFileSystem{MaxFileSize: filesystem.DefaultMaxFileSize},
					"testdata/archive-golden.zip",
					testZipArchiveFS,
					CollectFiles(testZipArchiveFS, "."),
				},
				wantOutputPath: "testdata/archive-golden.zip",
				wantErr:        false,
			},
		*/
		{
			name: "normal",
			args: args{
				filesystem.AbsDirFileSystem(absTempDir),
				filepath.Join(absTempDir, "testdata/archive-golden.zip"),
				testZipArchiveFS,
				CollectFiles(testZipArchiveFS, "."),
			},
			wantOutputPath: filepath.Join(absTempDir, "testdata/archive-golden.zip"),
			wantErr:        false,
		},
		{
			name: "error empty-base-name",
			args: args{
				filesystem.AbsDirFileSystem(absTempDir),
				filepath.Join(absTempDir, "testdata") + string(os.PathSeparator), // ends by path separator.
				testZipArchiveFS,
				CollectFiles(testZipArchiveFS, "."),
			},
			wantOutputPath: "",
			wantErr:        true,
		},
		{
			name: "overwrite already exist",
			args: args{
				filesystem.AbsDirFileSystem(absTempDir),
				filepath.Join(absTempDir, "testdata/archive-golden.zip"), // same as normal case.
				testZipArchiveFS,
				CollectFiles(testZipArchiveFS, "."),
			},
			wantOutputPath: filepath.Join(absTempDir, "testdata/archive-golden.zip"),
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutputPath, err := ArchiveAsZip(tt.args.outFsys, tt.args.outPathPrefix, tt.args.srcFsys, tt.args.targetFiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveAsZip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutputPath != tt.wantOutputPath {
				t.Errorf("ArchiveAsZip() = %v, want %v", gotOutputPath, tt.wantOutputPath)
			}
		})
	}
}

func TestArchiveAsZipWriter(t *testing.T) {
	goldenBytes, err := testdataFS.ReadFile("testdata/archive-golden.zip")
	if err != nil {
		t.Fatal(err)
	}
	goldenStr := string(goldenBytes)
	localizedPath := func(ss []string) []string {
		sss := make([]string, 0, len(ss))
		for _, s := range ss {
			ls, err := filepath.Localize(s)
			if err != nil {
				t.Errorf("Localize failed for %v", s)
			}
			sss = append(sss, ls)
		}
		return sss
	}
	type args struct {
		archiveBaseName string
		srcFsys         fs.FS
		targetFiles     []string
	}
	tests := []struct {
		name    string
		args    args
		checkW  bool
		wantW   string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				"archive-golden",
				testZipArchiveFS,
				CollectFiles(testZipArchiveFS, "."),
			},
			checkW:  true,
			wantW:   goldenStr,
			wantErr: false,
		},
		{
			name: "normal even if separator is backslash",
			args: args{
				"archive-golden",
				testZipArchiveFS,
				localizedPath(CollectFiles(testZipArchiveFS, ".")),
			},
			checkW:  true,
			wantW:   goldenStr,
			wantErr: false,
		},
		{
			name: "normal empty target files",
			args: args{
				"archive-golden",
				testZipArchiveFS,
				[]string{},
			},
			checkW:  false,
			wantW:   "",
			wantErr: false,
		},
		{
			name: "error not found path",
			args: args{
				"archive-golden",
				testZipArchiveFS,
				[]string{"path/to/not-found", "p/t/n-f"},
			},
			checkW:  false,
			wantW:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := ArchiveAsZipWriter(w, tt.args.archiveBaseName, tt.args.srcFsys, tt.args.targetFiles); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveAsZipWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkW {
				if gotW := w.String(); gotW != tt.wantW {
					t.Errorf("ArchiveAsZipWriter() = %+v, want %+v", []byte(gotW), []byte(tt.wantW))
				}
			}
		})
	}
}

func TestCollectFiles(t *testing.T) {
	mustFs := func(fsys fs.FS, err error) fs.FS {
		if err != nil {
			t.Fatal(err)
		}
		return fsys
	}
	type args struct {
		fsys   fs.ReadDirFS
		relDir string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
	}{
		{"normal", args{testZipArchiveFS, "test-dir"}, []string{"test-dir/test2.txt", "test-dir/test3.txt", "test-dir/マルチバイト文字.txt"}},
		{"empty", args{testZipArchiveFS, "not-found-dir"}, []string{}},
		{"all", args{testZipArchiveFS, "."}, []string{"test-dir/test2.txt", "test-dir/test3.txt", "test-dir/マルチバイト文字.txt", "test.txt"}},
		{"sub-all", args{mustFs(fs.Sub(testZipArchiveFS, "test-dir")).(fs.ReadDirFS), "."}, []string{"test2.txt", "test3.txt", "マルチバイト文字.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFiles := CollectFiles(tt.args.fsys, tt.args.relDir); !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("CollectFiles() = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}
