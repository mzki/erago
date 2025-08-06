package model

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mzki/erago/app"
	"github.com/mzki/erago/filesystem"
)

type stubUI struct {
	cbOnCommandRequested func()
}

func (ui stubUI) OnPublishBytes(_ []byte) (_ error) {
	return nil
}

func (ui stubUI) OnPublishBytesTemporary(_ []byte) (_ error) {
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
func (ui stubUI) OnCommandRequested() {
	if cb := ui.cbOnCommandRequested; cb != nil {
		cb()
	}
}

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
			args:    args{&stubUI{}, absStubDir, InitOptions{ImageFetchNone, MessageByteEncodingJson, FromGoFSGlob(filesystem.Desktop)}},
			wantErr: false,
		},
		{
			name:    "normal with default filesystem",
			args:    args{&stubUI{}, absStubDir, InitOptions{ImageFetchNone, MessageByteEncodingJson, nil}},
			wantErr: false,
		},
		{
			name:    "normal with relative dir",
			args:    args{&stubUI{}, "../../../stub", InitOptions{ImageFetchNone, MessageByteEncodingJson, nil}},
			wantErr: true,
		},
		{
			name:    "error config nor script files not found",
			args:    args{&stubUI{}, absTempDir, InitOptions{ImageFetchNone, MessageByteEncodingJson, nil}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(tt.args.baseDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(absCurrentDir)

			if err := Init(tt.args.ui, tt.args.baseDir, &tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ensure all resource is released.
			Quit()
		})
	}
}

type stubAppContext struct {
	ctx  context.Context
	err  error
	quit chan error
}

func newStubAppContext(ctx context.Context) *stubAppContext {
	return &stubAppContext{
		ctx:  ctx,
		err:  nil,
		quit: make(chan error),
	}
}

func (stub *stubAppContext) NotifyQuit(err error) {
	stub.err = err
	stub.quit <- err
	close(stub.quit)
}

func (stub *stubAppContext) Done() <-chan error {
	return stub.quit
}

func (stub *stubAppContext) Err() error {
	return stub.err
}

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
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal", args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(absCurrentDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(absCurrentDir)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmdReqCh := make(chan struct{})
			cbCmdReq := func() {
				go func() { //invoke goroutine to avoid blocking callback.
					select {
					case cmdReqCh <- struct{}{}:
					case <-ctx.Done():
					}
				}()
			}

			if err := Init(&stubUI{cbOnCommandRequested: cbCmdReq}, absStubDir, &InitOptions{ImageFetchNone, MessageByteEncodingJson, nil}); err != nil {
				t.Fatal(err)
			}
			appContext := newStubAppContext(ctx)
			Main(appContext)
			select {
			case <-cmdReqCh:
				// OK, go next step
			case <-ctx.Done():
				t.Errorf("Main(), exceed timelimit to receive cmdReq, %v", ctx.Err())
				// NG, but need to continue to call Quit().
			}
			Quit() // immediately
			// below cases can be happened together, in such case the result is undeterminded.
			// someitimes OK, and the others NG.
			// In case of context done is first, its error is already captured at above. So we can take any result at here.
			// In case of appContext done is first, it is nomarl terminataion or abonormal one. decided at following steps.
			select {
			case <-ctx.Done():
				t.Errorf("Main(), failed to quit corretly. Canceled by parent context, %v", ctx.Err())
			case <-appContext.Done():
				switch err := appContext.Err(); {
				case err == nil || errors.Is(err, context.Canceled):
					// OK
				default:
					t.Errorf("Main(), failed to quit correctly. error = %v", err)
				}
			}
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
				if err := Init(&stubUI{}, absStubDir, &InitOptions{ImageFetchNone, MessageByteEncodingJson, nil}); err != nil {
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

func Test_disableDesktopFeatures(t *testing.T) {
	type args struct {
		appConfFn func() *app.Config
	}
	tests := []struct {
		name        string
		args        args
		wantChanged bool
		wantMessage string
	}{
		{
			name: "Game.Script.ReloadFileChanged",
			args: args{func() *app.Config {
				appConf := app.NewConfig("./")
				appConf.Game.ScriptConfig.ReloadFileChange = true
				return appConf
			}},
			wantChanged: true,
			wantMessage: fmt.Sprintf("Game.Script.ReloadFileChange = %v", false),
		},
		{
			name: "LogFile",
			args: args{func() *app.Config {
				appConf := app.NewConfig("./")
				appConf.LogFile = "stdout"
				return appConf
			}},
			wantChanged: true,
			wantMessage: "LogFile = " + app.DefaultLogFile,
		},
		{
			name: "LogLimitMegaByte",
			args: args{func() *app.Config {
				appConf := app.NewConfig("./")
				appConf.LogLimitMegaByte = app.DefaultLogLimitMegaByte + 1
				return appConf
			}},
			wantChanged: true,
			wantMessage: fmt.Sprintf("LogLimitMegaByte = %v", app.DefaultLogLimitMegaByte),
		},
		{
			name: "no change",
			args: args{func() *app.Config {
				appConf := app.NewConfig("./")
				appConf.Game.ScriptConfig.ReloadFileChange = false
				appConf.LogFile = app.DefaultLogFile
				appConf.LogLimitMegaByte = app.DefaultLogLimitMegaByte
				return appConf
			}},
			wantChanged: false,
			wantMessage: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConf := tt.args.appConfFn()
			gotChanged, gotMessage := disableDesktopFeatures(appConf)
			if gotChanged != tt.wantChanged {
				t.Errorf("disableDesktopFeatures() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if !strings.Contains(gotMessage, tt.wantMessage) {
				t.Errorf("disableDesktopFeatures() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}
