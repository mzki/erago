package model

import (
	"embed"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/mzki/erago/filesystem"
)

//go:embed testdata/*
var zipTestdataFS embed.FS

func TestNewOSDirFileSystem(t *testing.T) {
	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		absPath string
	}
	tests := []struct {
		name string
		args args
		want filesystem.FileSystem
	}{
		{"normal", args{absTempDir}, filesystem.AbsDirFileSystem(absTempDir)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOSDirFileSystem(tt.args.absPath)
			w, err := got.Store("test-file.txt")
			if err != nil {
				t.Errorf("NewOSDirFileSystem() failed to Store with filesystem. error = %v", err)
			}
			w.Close()
		})
	}
}

func TestInstallPackage(t *testing.T) {
	zipBs, err := zipTestdataFS.ReadFile("testdata/archive-golden.zip")
	if err != nil {
		t.Fatal(err)
	}
	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		outFsys  filesystem.FileSystem
		zipBytes []byte
	}
	tests := []struct {
		name             string
		args             args
		outDir           string
		wantExtractedDir string
		wantErr          bool
	}{
		{"normal", args{filesystem.AbsDirFileSystem(absTempDir), zipBs}, absTempDir, "archive-golden", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExtractedDir, err := InstallPackage(FromGoFS(tt.args.outFsys), tt.args.zipBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("InstallPackage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExtractedDir != tt.wantExtractedDir {
				t.Errorf("InstallPackage() = %v, want %v", gotExtractedDir, tt.wantExtractedDir)
			}
			extractedPath := filepath.Join(tt.outDir, gotExtractedDir)
			if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
				t.Errorf("InstallPackage(), should exist extractedPath = %v, but not", extractedPath)
			}
		})
	}
}

var savTestdataFS = fstest.MapFS{
	"testdata/exportsav/erago.conf":     {Data: []byte("")}, // filled later.
	"testdata/exportsav/sav/save00.sav": {Data: []byte("save00.sav.content")},
	"testdata/exportsav/sav/save01.sav": {Data: []byte("save01.sav.content")},
	"testdata/exportsav/sav/save17.sav": {Data: []byte("save17.sav.content")},
	"testdata/exportsav/sav/share.sav":  {Data: []byte("share.sav.content")},
}

func TestExportSav(t *testing.T) {
	savDataFS, err := savTestdataFS.Sub("testdata/exportsav")
	if err != nil {
		t.Fatal(err)
	}

	absEragoDir, err := filepath.Abs("testdata/exportsav")
	if err != nil {
		t.Fatal(err)
	}

	var configOnlyFS = fstest.MapFS{}
	{
		appConfBs, err := zipTestdataFS.ReadFile("testdata/exportsav/erago.conf")
		if err != nil {
			t.Fatal(err)
		}
		savTestdataFS["testdata/exportsave/erago.conf"] = &fstest.MapFile{Data: appConfBs}
		configOnlyFS["erago.conf"] = &fstest.MapFile{Data: appConfBs}
	}

	goldenBs, err := zipTestdataFS.ReadFile("testdata/exportsav/golden.zip")
	if err != nil {
		t.Fatal(err)
	}

	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// skipBytes indicates byte comparision is skipped.
	var skipBytes = []byte{}

	type args struct {
		absEragoDir string
		eragoFsys   filesystem.FileSystemGlob
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// {"golden", args{absEragoDir, filesystem.FromFS(savDataFS)}, nil, false},
		{"normal", args{absEragoDir, filesystem.FromFS(savDataFS)}, goldenBs, false},
		{"normal with absFS, without byte result comparison", args{absEragoDir, filesystem.AbsDirFileSystem(absEragoDir)}, skipBytes, false},
		{"error config nor sav not found", args{absTempDir, filesystem.AbsDirFileSystem(absTempDir)}, nil, true},
		{"error config found but sav pattern not matched", args{absEragoDir, filesystem.FromFS(configOnlyFS)}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExportSav(tt.args.absEragoDir, FromGoFSGlob(tt.args.eragoFsys))
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportSav() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "golden" {
				f, err := os.Create("testdata/exportsav/golden.zip")
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if _, err := f.Write(got); err != nil {
					t.Fatal(err)
				}
				return
			}

			if !reflect.DeepEqual(tt.want, skipBytes) {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ExportSav() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMatchGlobPattern(t *testing.T) {
	type args struct {
		pattern string
		path    string
	}
	tests := []struct {
		name   string
		args   args
		wantFn func(args) (want bool, wantErr bool)
	}{
		{
			name: "normal",
			args: args{"*_test.go", "install_test.go"},
			wantFn: func(a args) (want bool, wantErr bool) {
				ok, err := filepath.Match(a.pattern, a.path)
				return ok, err != nil
			},
		},
		{
			name: "normal-os-specific",
			args: args{filepath.Join("test", "*_test.go"), filepath.Join("test", "install_test.go")},
			wantFn: func(a args) (want bool, wantErr bool) {
				ok, err := filepath.Match(a.pattern, a.path)
				return ok, err != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, wantErr := tt.wantFn(tt.args)
			got, err := MatchGlobPattern(tt.args.pattern, tt.args.path)
			if (err != nil) != wantErr {
				t.Errorf("MatchGlobPattern() error = %v, wantErr %v", err, wantErr)
				return
			}
			if got != want {
				t.Errorf("MatchGlobPattern() = %v, want %v", got, want)
			}
		})
	}
}

var logTestdataFS = fstest.MapFS{
	"testdata/exportlog/erago.conf": {Data: []byte("")}, // filled later.
	"testdata/exportlog/erago.log":  {Data: []byte("log file")},
}

func TestExportLog(t *testing.T) {
	logDataFS, err := logTestdataFS.Sub("testdata/exportlog")
	if err != nil {
		t.Fatal(err)
	}

	absEragoDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	absEragoDir = filepath.Join(absEragoDir, "testdata/exportlog")

	var configOnlyFS = fstest.MapFS{}
	{
		appConfBs, err := zipTestdataFS.ReadFile("testdata/exportsav/erago.conf")
		if err != nil {
			t.Fatal(err)
		}
		logTestdataFS["testdata/exportlog/erago.conf"] = &fstest.MapFile{Data: appConfBs}
		configOnlyFS["erago.conf"] = &fstest.MapFile{Data: appConfBs}
	}

	goldenBs, err := logTestdataFS.ReadFile("testdata/exportlog/erago.log")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		absEragoDir string
		eragoFsys   filesystem.FileSystemGlob
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"normal", args{absEragoDir, filesystem.FromFS(logDataFS)}, goldenBs, false},
		{"error no log", args{absEragoDir, filesystem.FromFS(configOnlyFS)}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExportLog(tt.args.absEragoDir, FromGoFSGlob(tt.args.eragoFsys))
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExportLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportSav(t *testing.T) {
	savDataFS, err := savTestdataFS.Sub("testdata/exportsav")
	if err != nil {
		t.Fatal(err)
	}

	absEragoDir, err := filepath.Abs("testdata/exportsav")
	if err != nil {
		t.Fatal(err)
	}

	var configOnlyFS = fstest.MapFS{}
	{
		appConfBs, err := zipTestdataFS.ReadFile("testdata/exportsav/erago.conf")
		if err != nil {
			t.Fatal(err)
		}
		savTestdataFS["testdata/exportsave/erago.conf"] = &fstest.MapFile{Data: appConfBs}
		configOnlyFS["erago.conf"] = &fstest.MapFile{Data: appConfBs}
	}

	goldenBs, err := zipTestdataFS.ReadFile("testdata/exportsav/golden.zip")
	if err != nil {
		t.Fatal(err)
	}

	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	storeFn := func(s string) (io.WriteCloser, error) {
		fpath := filepath.Join(absTempDir, s)
		dir, _ := filepath.Split(fpath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.Create(fpath)
	}

	withStoreFn := func(ifs *filesystem.InteropFileSystem) *filesystem.InteropFileSystem {
		ifs.StoreFn = storeFn
		return ifs
	}
	type args struct {
		absEragoDir string
		eragoFsys   filesystem.FileSystemGlob
		savZipBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{absEragoDir, withStoreFn(filesystem.FromFS(savDataFS)), goldenBs}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ImportSav(tt.args.absEragoDir, FromGoFSGlob(tt.args.eragoFsys), tt.args.savZipBytes); (err != nil) != tt.wantErr {
				t.Errorf("ImportSav() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
