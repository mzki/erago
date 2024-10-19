package app

import (
	"embed"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mzki/erago/infra/serialize/toml"
)

//go:embed testdata/*
var testdataFS embed.FS

func TestGenGolden(t *testing.T) {
	t.Skip("golden data generaten is not a test.")
	_, err := LoadConfigOrDefault(filepath.Join("./testdata", ConfigFile))
	if errors.Is(err, ErrDefaultConfigGenerated) {
		// OK
	} else if err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigOrDefault(t *testing.T) {
	confFile, err := testdataFS.Open(filepath.Join("testdata", ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	defer confFile.Close()

	var confGolden Config
	if err := toml.Decode(confFile, &confGolden); err != nil {
		t.Fatal(err)
	}

	var confGoldenLack = confGolden
	confGoldenLack.FontSize = DefaultFontSize
	confGoldenLack.LogLimitMegaByte = DefaultLogLimitMegaByte

	tempDir := t.TempDir()

	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
		err     error
	}{
		{"exist config", args{filepath.Join("./testdata", ConfigFile)}, &confGolden, false, nil},
		{"exist config but lack of some value", args{filepath.Join("./testdata", ConfigFile+".lack")}, &confGoldenLack, false, nil},
		{"not exist config", args{filepath.Join(tempDir, ConfigFile)}, NewConfig(DefaultBaseDir), true, ErrDefaultConfigGenerated},
		{"can not create config", args{filepath.Join(tempDir)}, NewConfig(DefaultBaseDir), true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfigOrDefault(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigOrDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.err != nil && (!errors.Is(err, tt.err)) {
					t.Errorf("LoadConfigOrDefault() error = %v, should same as %v", err, tt.err)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfigOrDefault() = %v, want %v", got, tt.want)
			}
			if _, err := os.Stat(tt.args.file); os.IsNotExist(err) {
				t.Errorf("LoadConfigOrDefault() file %v should exist but not", tt.args.file)
			}

		})
	}
}
