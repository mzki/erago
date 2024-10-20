package model

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

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
			gotExtractedDir, err := InstallPackage(tt.args.outFsys, tt.args.zipBytes)
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
