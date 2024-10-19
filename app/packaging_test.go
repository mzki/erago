package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackaging(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir("../stub"); err != nil {
		t.Fatalf("Need to change directory, error = %v", err)
	}
	defer os.Chdir("../app") // to back current dir

	appConf, err := LoadConfigOrDefault(ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	appConf.LogFile = "stdout" // to supress generate log file

	var appConfCSVInvalid Config = *appConf
	appConfCSVInvalid.Game.CSVConfig.Dir = "path/to/not-found-dir"

	type args struct {
		dstDir      string
		appConf     *Config
		appConfPath string
		extraFiles  []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal", args{tempDir, appConf, ConfigFile, []string{"doc.go"}}, true},
		{"error already exist", args{tempDir, appConf, ConfigFile, []string{"doc.go"}}, false},
		{"error CSV path is invalid", args{tempDir, &appConfCSVInvalid, ConfigFile, []string{"doc.go"}}, false},
		{"error extra file is not found", args{filepath.Join(tempDir, "extra-file-not-found"), appConf, ConfigFile, []string{"path/to/not-found"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Packaging(tt.args.dstDir, tt.args.appConf, tt.args.appConfPath, tt.args.extraFiles); got != tt.want {
				t.Errorf("Packaging() = %v, want %v", got, tt.want)
			}
		})
	}
}
