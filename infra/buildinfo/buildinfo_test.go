package buildinfo

import (
	"os"
	"reflect"
	"testing"
)

func TestGetWithoutCompileTimeInfo(t *testing.T) {
	if os.Getenv("GOTEST_BUILDINFO_COMPILE_TIME_INFO") == "true" {
		t.Skip("Without compile time information test is skipped")
	}
	tests := []struct {
		name string
		want BuildInfo
	}{
		{name: "normal", want: BuildInfo{
			Version:    "dev",
			CommitHash: "none",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithCompileTimeInfo(t *testing.T) {
	if os.Getenv("GOTEST_BUILDINFO_COMPILE_TIME_INFO") != "true" {
		t.Skip("With compile time information test is skipped")
	}
	tests := []struct {
		name string
		want BuildInfo
	}{
		{name: "normal", want: BuildInfo{
			Version:    "v1.0.0",
			CommitHash: "34567#",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
