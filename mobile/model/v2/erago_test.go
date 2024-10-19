package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mzki/erago/filesystem"
)

type stubUI struct{}

func (ui stubUI) OnPublishJson(_ string) (_ error) {
	return nil
}

func (ui stubUI) OnPublishJsonTemporary(_ string) (_ error) {
	return nil
}

func (ui stubUI) OnRemove(nParagraph int) (_ error) {
	return nil
}

func (ui stubUI) OnRemoveAll() (_ error) {
	return nil
}

// it is called when mobile.app requires inputting
// user's command.
func (ui stubUI) OnCommandRequested() {}

// it is called when mobile.app requires just input any command.
func (ui stubUI) OnInputRequested() {}

// it is called when mobile.app no longer requires any input,
// such as just-input and command.
func (ui stubUI) OnInputRequestClosed() {}

func TestInit(t *testing.T) {
	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	absCurrentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	absStubDir, err := filepath.Abs("../../../stub")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ui      UI
		baseDir string
		options InitOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "normal",
			args:    args{&stubUI{}, absStubDir, InitOptions{ImageFetchNone, filesystem.Desktop}},
			wantErr: false,
		},
		{
			name:    "normal with default filesystem",
			args:    args{&stubUI{}, absStubDir, InitOptions{ImageFetchNone, nil}},
			wantErr: false,
		},
		{
			name:    "normal with relative dir",
			args:    args{&stubUI{}, "../../../stub", InitOptions{ImageFetchNone, nil}},
			wantErr: true,
		},
		{
			name:    "error config nor script files not found",
			args:    args{&stubUI{}, absTempDir, InitOptions{ImageFetchNone, nil}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(tt.args.baseDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(absCurrentDir)

			if err := Init(tt.args.ui, tt.args.baseDir, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ensure all resource is released.
			Quit()
		})
	}
}

type stubAppContext struct{}

func (stubAppContext) NotifyQuit(error) {}

func TestMain(t *testing.T) {
	absCurrentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	absStubDir, err := filepath.Abs("../../../stub")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		appContext AppContext
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal", args{&stubAppContext{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(absCurrentDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(absCurrentDir)

			if err := Init(&stubUI{}, absStubDir, InitOptions{ImageFetchNone, nil}); err != nil {
				t.Fatal(err)
			}
			defer Quit()
			Main(tt.args.appContext)
		})
	}
}

func TestQuit(t *testing.T) {
	absCurrentDir, err := filepath.Abs("./")
	if err != nil {
		t.Fatal(err)
	}
	absStubDir, err := filepath.Abs("../../../stub")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		doInit    bool
		wantPanic bool
	}{
		{"normal", true, false},
		{"error not initialized", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if rec := recover(); (tt.wantPanic == false) && (rec != nil) {
					t.Errorf("%v", rec)
				}
			}()
			if err := os.Chdir(absCurrentDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(absCurrentDir)

			if tt.doInit {
				if err := Init(&stubUI{}, absStubDir, InitOptions{ImageFetchNone, nil}); err != nil {
					t.Fatal(err)
				}
			}
			Quit()
		})
	}
}

func TestSendCommand(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendCommand(tt.args.cmd)
		})
	}
}

func TestSendSkippingWait(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendSkippingWait()
		})
	}
}

func TestSendStopSkippingWait(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendStopSkippingWait()
		})
	}
}

func TestSetViewSize(t *testing.T) {
	type args struct {
		lineCount     int
		lineRuneWidth int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetViewSize(tt.args.lineCount, tt.args.lineRuneWidth); (err != nil) != tt.wantErr {
				t.Errorf("SetViewSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetTextUnitPx(t *testing.T) {
	type args struct {
		textUnitWidthPx  float64
		textUnitHeightPx float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetTextUnitPx(tt.args.textUnitWidthPx, tt.args.textUnitHeightPx); (err != nil) != tt.wantErr {
				t.Errorf("SetTextUnitPx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
