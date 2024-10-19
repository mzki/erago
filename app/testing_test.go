package app

import (
	"os"
	"testing"
)

func TestTesting(t *testing.T) {
	if err := os.Chdir("../stub"); err != nil {
		t.Fatalf("Need to change directory, error = %v", err)
	}
	defer os.Chdir("../app") // to back current dir

	appConf, err := LoadConfigOrDefault(ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	appConf.LogFile = "stdout" // to supress generate log file

	type args struct {
		appConf     *Config
		scriptFiles []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal", args{appConf, []string{"../stub/ELA/game_test.lua", "../stub/ELA/game_test_input_request.lua"}}, true},
		{"error not found path", args{appConf, []string{"path/to/not-found.lua"}}, false},
		{"error test file not found", args{appConf, []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Testing(tt.args.appConf, tt.args.scriptFiles); got != tt.want {
				t.Errorf("Testing() = %v, want %v", got, tt.want)
			}
		})
	}
}
