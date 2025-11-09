package publisher_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
	mock_publisher "github.com/mzki/erago/view/exp/text/publisher/mock"
	gomock "go.uber.org/mock/gomock"
	"golang.org/x/image/math/fixed"
)

type globals struct {
	editor *publisher.Editor
	ctrl   *gomock.Controller
	cancel context.CancelFunc
}

func setupGlobals(t *testing.T, opts ...publisher.EditorOptions) struct {
	editor *publisher.Editor
	ctrl   *gomock.Controller
	cancel context.CancelFunc
} {
	ctx, cancel := context.WithCancel(context.Background())
	editor := publisher.NewEditor(ctx, opts...)
	if err := editor.SetViewSize(10, 100); err != nil {
		t.Fatal(err)
	}
	if err := editor.SetTextUnitPx(fixed.Point26_6{X: fixed.I(8), Y: fixed.I(14)}); err != nil {
		t.Fatal(err)
	}
	ctrl := gomock.NewController(t)
	cancelFunc := context.CancelFunc(func() {
		ctrl.Finish()
		editor.Close()
		cancel()
	})
	return globals{
		editor: editor,
		ctrl:   ctrl,
		cancel: cancelFunc,
	}
}

func TestEditor_Print(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	for _, testcase := range []struct {
		text            string
		newMockCallback func() publisher.Callback
	}{
		{
			"あいうえお",
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(0)
				return cb
			},
		},
		{
			"abcdef",
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(0)
				return cb
			},
		},
		{
			"a_b_c_\n_d_e_f_g\n_a",
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(2).Return(nil)
				return cb
			},
		},
		{
			strings.Repeat("long-text,", 100),
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(0)
				return cb
			},
		},
	} {
		s := testcase.text
		cb := testcase.newMockCallback()
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v, with s: %v", err, s)
		}
		if err := editor.Print(s); err != nil {
			t.Errorf("Can not print text: %v, with s: %v", err, s)
		}
	}
}

func TestEditor_Print_PublishedData(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	var publishedCount int64 = 0

	for _, testcase := range []struct {
		text            string
		newMockCallback func(s string) publisher.Callback
	}{
		{
			"newline\nnewline\n",
			func(s string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(2).DoAndReturn(func(p *pubdata.Paragraph) error {
					if v := p.Id; v != publishedCount {
						t.Errorf("Paragraph have invalid ID. want: %v, got: %v", publishedCount, v)
					}
					publishedCount++

					if v := p.Fixed; v != true {
						t.Errorf("Paragraph should have been fixed. want: %v, got:%v", true, v)
					}

					if v := len(p.Lines); v != 1 {
						t.Logf("%#v", p)
						t.Fatalf("Paragraph should have 1 line but not. want: %v, got: %v", 1, v)
					}
					line := p.Lines[0]
					if v := len(line.Boxes); v != 1 {
						t.Logf("%#v", line)
						t.Fatalf("Line should have 1 box but not. want: %v, got: %v", 1, v)
					}
					box := line.Boxes[0]
					if v := box.ContentType; v != pubdata.ContentType_CONTENT_TYPE_TEXT {
						t.Logf("%#v", box)
						t.Fatalf("Box should have ContentTypeText but not. want: %v, got: %v", pubdata.ContentType_CONTENT_TYPE_TEXT, v)
					}
					data := box.Data.(*pubdata.Box_TextData).TextData
					if data.Text != "newline" {
						t.Errorf("Box.Text not match. want: %v, got: %v", "newline", data.Text)
					}
					return nil
				})
				// Sync expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Return(nil)
				cb.EXPECT().OnSync().Return(nil)
				return cb
			},
		},
	} {
		s := testcase.text
		cb := testcase.newMockCallback(s)
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v, with s: %v", err, s)
		}
		if err := editor.Print(s); err != nil {
			t.Errorf("Can not print text: %v, with s: %v", err, s)
		}
		if err := editor.Sync(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEditor_PrintError(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	var errOnPublish error = errors.New("Error OnPublish")

	for _, testcase := range []struct {
		text            string
		newMockCallback func() publisher.Callback
		err             error
	}{
		{
			"newline\n",
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).Return(errOnPublish)
				return cb
			},
			errOnPublish,
		},
	} {
		s := testcase.text
		cb := testcase.newMockCallback()
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v, with s: %v", err, s)
		}
		if err := editor.Print(s); err != nil {
			t.Fatal(err)
		}
		// next call caused error since async error raised before calling.
		if err := editor.Print(s); !errors.Is(err, testcase.err) {
			t.Errorf("InternalError want: %v, got: %v", testcase.err, err)
		}
	}
}

func TestEditor_PrintButton(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	editor := publisher.NewEditor(ctx)
	defer editor.Close()
	if err := editor.SetViewSize(10, 100); err != nil {
		t.Fatal(err)
	}

	for _, s := range [][]string{
		{"あいうえお", "100"},
		{"abcdefg", "200"},
		{"a_b_c_", "300"},
		{strings.Repeat("long-text,", 100), "-1"},
	} {
		if err := editor.PrintButton(s[0], s[1]); err != nil {
			t.Errorf("Can not print text and command: %v", s)
		}
	}
}

func TestEditorSync(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	doneCh := make(chan struct{}, 1) // to pass async call
	for _, testcase := range []struct {
		newMockCallback func() publisher.Callback
	}{
		{
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(p *pubdata.Paragraph) error {
					if p.Id != 0 {
						t.Errorf("Sync() returns invalid ID. want: %v, got:%v", 0, p.Id)
					}
					if v := p.Fixed; v != false {
						t.Errorf("Sync() returns non-fixed Paragraph. want:%v, got:%v", false, v)
					}
					return nil
				})
				cb.EXPECT().OnSync().Times(1).Do(func() {
					doneCh <- struct{}{}
				})
				return cb
			},
		},
		{
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(p *pubdata.Paragraph) error {
					if p.Id != 0 {
						t.Errorf("Sync() returns invalid ID. want: %v, got:%v", 0, p.Id)
					}
					return nil
				})
				cb.EXPECT().OnSync().Times(1).Do(func() {
					doneCh <- struct{}{}
				})
				return cb
			},
		},
	} {
		cb := testcase.newMockCallback()
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v", err)
		}
		if err := editor.Sync(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-doneCh:
			// OK
		default:
			t.Fatal("Synchronous call Sync() but done channel not respond")
		}
	}
}

func TestEditorSyncTimeout(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type precond struct {
		doneCh chan struct{}
	}
	for _, testcase := range []struct {
		name            string
		newPrecond      func() *precond
		newMockCallback func(cond *precond) publisher.Callback
	}{
		{
			name: "timeout at OnPublishTemporary",
			newPrecond: func() *precond {
				return &precond{doneCh: make(chan struct{})}
			},
			newMockCallback: func(cond *precond) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					<-cond.doneCh // infinite loop
					return nil
				})
				cb.EXPECT().OnSync().Times(1) // to be called after initite loop ends.
				return cb
			},
		},
		{
			name: "timeout at OnSync",
			newPrecond: func() *precond {
				return &precond{doneCh: make(chan struct{})}
			},
			newMockCallback: func(cond *precond) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1).DoAndReturn(func() error {
					<-cond.doneCh // infinite loop
					return nil
				})
				return cb
			},
		},
	} {
		cond := testcase.newPrecond()
		cb := testcase.newMockCallback(cond)
		t.Run(testcase.name, func(t *testing.T) {
			defer close(cond.doneCh)
			if err := editor.SetCallback(cb); err != nil {
				t.Fatalf("Can not set callbakc: %v", err)
			}
			if err := editor.Sync(); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Sync() with not respond from Callback, wantErr: %v, got: %v", context.DeadlineExceeded, err)
			}
		})
	}
}

func TestEditorSyncError(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	var errOnPublishTemporary = errors.New("errOnPublishTemporary")
	var errOnSync = errors.New("errOnSync")

	type precond struct {
		doneCh chan struct{}
	}
	for _, testcase := range []struct {
		newPrecond      func() *precond
		newMockCallback func(cond *precond) publisher.Callback
		err             error
	}{
		{
			newPrecond: func() *precond {
				return &precond{doneCh: make(chan struct{}, 1)}
			},
			newMockCallback: func(cond *precond) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					cond.doneCh <- struct{}{}
					return errOnPublishTemporary
				})
				cb.EXPECT().OnSync().Times(0)
				return cb
			},
			err: errOnPublishTemporary,
		},
		{
			newPrecond: func() *precond {
				return &precond{doneCh: make(chan struct{}, 1)}
			},
			newMockCallback: func(cond *precond) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					return nil
				})
				cb.EXPECT().OnSync().Times(1).DoAndReturn(func() error {
					cond.doneCh <- struct{}{}
					return errOnSync
				})
				return cb
			},
			err: errOnSync,
		},
	} {
		cond := testcase.newPrecond()
		cb := testcase.newMockCallback(cond)
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v", err)
		}
		if err := editor.Sync(); !errors.Is(err, testcase.err) {
			t.Errorf("Unexpected Sync() error: want: %v, got:%v", testcase.err, err)
		}
		select {
		case <-cond.doneCh:
			// OK
		default:
			t.Fatal("Synchronous call Sync() but done channel not respond")
		}
	}
}

func TestEditor_PrintLabel(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type args struct {
		s string
	}
	tests := []struct {
		name            string
		e               *publisher.Editor
		args            args
		wantErr         bool
		newMockCallback func() publisher.Callback
	}{
		{
			name:    "short-label",
			e:       editor,
			args:    args{"short-label"},
			wantErr: false,
			newMockCallback: func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(0)
				return cb
			},
		},
		{
			name:    "long-label",
			e:       editor,
			args:    args{strings.Repeat("long-label", 10)},
			wantErr: false,
			newMockCallback: func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(0)
				return cb
			},
		},
	}
	for _, tt := range tests {
		tt.e.SetCallback(tt.newMockCallback())
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.PrintLabel(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Editor.PrintLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditor_PrintLine(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type args struct {
		sym string
	}
	tests := []struct {
		name            string
		e               *publisher.Editor
		args            args
		wantErr         bool
		newMockCallback func() publisher.Callback
		setup           func(e *publisher.Editor)
	}{
		{
			name:    "hold no content",
			e:       editor,
			args:    args{"#"},
			wantErr: false,
			newMockCallback: func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).Return(nil)
				return cb
			},
			setup: func(e *publisher.Editor) {},
		},
		{
			name:    "hold some content",
			e:       editor,
			args:    args{"?"},
			wantErr: false,
			newMockCallback: func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(2).Return(nil)
				return cb
			},
			setup: func(e *publisher.Editor) {
				_ = e.Print("some text")
			},
		},
	}
	for _, tt := range tests {
		tt.e.SetCallback(tt.newMockCallback())
		tt.setup(tt.e)
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.PrintLine(tt.args.sym); (err != nil) != tt.wantErr {
				t.Errorf("Editor.PrintLine() error = %v, wantErr %v", err, tt.wantErr)
			}
			if width, err := tt.e.CurrentRuneWidth(); err != nil {
				t.Fatal(err)
			} else if width != 0 {
				t.Errorf("After Editor.PrintLine() should not have some content.")
			}
		})
	}
}

func TestEditor_SetColor(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type args struct {
		color uint32
	}
	tests := []struct {
		name            string
		e               *publisher.Editor
		args            args
		wantErr         bool
		newMockCallback func(as args) publisher.Callback
	}{
		{
			name:    "Valid case",
			e:       editor,
			args:    args{0xffccaa},
			wantErr: false,
			newMockCallback: func(args args) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				// Print() expectation
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(p *pubdata.Paragraph) error {
					if v := len(p.Lines); v != 1 {
						t.Fatalf("Paragraph should have 1 line but not. want: %v, got: %v", 1, v)
					}
					line := p.Lines[0]
					if v := len(line.Boxes); v != 1 {
						t.Logf("%#v", p)
						t.Fatalf("Line should have 1 box but not. want: %v, got: %v", 1, v)
					}
					box := line.Boxes[0]
					expect := int32(args.color)
					if v := box.Data.(*pubdata.Box_TextData).TextData.Fgcolor; v != expect {
						t.Errorf("Published paragraph should have color by SetColor(). want: %v, got: %v", expect, v)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.SetCallback(tt.newMockCallback(tt.args)); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.SetColor(tt.args.color); (err != nil) != tt.wantErr {
				t.Errorf("Editor.SetColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.e.Print("newline\n"); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_GetColor(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name      string
		e         *publisher.Editor
		wantColor uint32
		wantErr   bool
		setup     func(e *publisher.Editor, wantColor uint32)
	}{
		{
			name:      "FirstGetColor",
			e:         editor,
			wantColor: publisher.ColorRGBAToUInt32RGB(publisher.ResetColor),
			wantErr:   false,
			setup:     func(e *publisher.Editor, wantColor uint32) {},
		},
		{
			name:      "GetColorAfterSetColor",
			e:         editor,
			wantColor: 0xffccaa,
			wantErr:   false,
			setup: func(e *publisher.Editor, wantColor uint32) {
				if err := e.SetColor(wantColor); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.wantColor)
			gotColor, err := tt.e.GetColor()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.GetColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotColor != tt.wantColor {
				t.Errorf("Editor.GetColor() = %v, want %v", gotColor, tt.wantColor)
			}
		})
	}
}

func TestEditor_ResetColor(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name    string
		e       *publisher.Editor
		wantErr bool
		setup   func(e *publisher.Editor)
	}{
		{
			name:    "AfterSetColor",
			e:       editor,
			wantErr: false,
			setup: func(e *publisher.Editor) {
				if err := e.SetColor(0xffccaa); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e)
			if err := tt.e.ResetColor(); (err != nil) != tt.wantErr {
				t.Errorf("Editor.ResetColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got, err := tt.e.GetColor(); err != nil {
				t.Fatal(err)
			} else if got != publisher.ColorRGBAToUInt32RGB(publisher.ResetColor) {
				t.Errorf("Editor.ResetColor() then GetColor() should return ResetColor. want: %v, got: %v",
					publisher.ColorRGBAToUInt32RGB(publisher.ResetColor), got)
			}
		})
	}
}

func TestEditor_SetAlignment(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type args struct {
		align attribute.Alignment
	}
	tests := []struct {
		name            string
		e               *publisher.Editor
		args            args
		wantErr         bool
		newMockCallback func(args) publisher.Callback
	}{
		{
			name:    "Valid case",
			e:       editor,
			args:    args{attribute.AlignmentCenter},
			wantErr: false,
			newMockCallback: func(args args) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				// Print() expectation
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(p *pubdata.Paragraph) error {
					expect := publisher.PdAlignment(args.align)
					if v := p.Alignment; v != expect {
						t.Errorf("Published paragraph should have alignment by SetAlignment(). want: %v, got: %v", expect, v)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.SetCallback(tt.newMockCallback(tt.args)); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.SetAlignment(tt.args.align); (err != nil) != tt.wantErr {
				t.Errorf("Editor.SetAlignment() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.e.Print("newline\n"); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_GetAlignment(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name      string
		e         *publisher.Editor
		wantAlign attribute.Alignment
		wantErr   bool
		setup     func(e *publisher.Editor, wantAlign attribute.Alignment)
	}{
		{
			name:      "FirstGet",
			e:         editor,
			wantAlign: attribute.AlignmentLeft,
			wantErr:   false,
			setup:     func(e *publisher.Editor, wantAlign attribute.Alignment) {},
		},
		{
			name:      "GetAfterSet",
			e:         editor,
			wantAlign: attribute.AlignmentCenter,
			wantErr:   false,
			setup: func(e *publisher.Editor, wantAlign attribute.Alignment) {
				if err := e.SetAlignment(wantAlign); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.wantAlign)
			gotAlign, err := tt.e.GetAlignment()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.GetAlignment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAlign, tt.wantAlign) {
				t.Errorf("Editor.GetAlignment() = %v, want %v", gotAlign, tt.wantAlign)
			}
		})
	}
}

func TestEditor_NewPage(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	tests := []struct {
		name            string
		e               *publisher.Editor
		wantErr         bool
		nlines          int
		setup           func(e *publisher.Editor, nlines int)
		newMockCallback func(nlines int) publisher.Callback
	}{
		{
			name:    "Valid case",
			e:       editor,
			wantErr: false,
			setup: func(e *publisher.Editor, nlines int) {
				if err := e.SetViewSize(nlines, 100); err != nil {
					t.Fatal(err)
				}
			},
			newMockCallback: func(nlines int) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(nlines).Return(nil)
				// Sync expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.nlines)
			if err := tt.e.SetCallback(tt.newMockCallback(tt.nlines)); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.NewPage(); (err != nil) != tt.wantErr {
				t.Errorf("Editor.NewPage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.e.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_ClearLine(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	type args struct {
		nline int
	}
	tests := []struct {
		name            string
		e               *publisher.Editor
		args            args
		wantErr         bool
		newMockCallback func(args) publisher.Callback
	}{
		{
			name:    "Valid case",
			e:       editor,
			args:    args{21},
			wantErr: false,
			newMockCallback: func(args args) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnRemove(args.nline - 1).Times(1).Return(nil)
				// Sync expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.SetCallback(tt.newMockCallback(tt.args)); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.ClearLine(tt.args.nline); (err != nil) != tt.wantErr {
				t.Errorf("Editor.ClearLine() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.e.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_ClearLineAll(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	tests := []struct {
		name            string
		e               *publisher.Editor
		wantErr         bool
		newMockCallback func() publisher.Callback
	}{
		{
			name:    "Valid case",
			e:       editor,
			wantErr: false,
			newMockCallback: func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnRemoveAll().Times(1).Return(nil)
				// Sync expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.SetCallback(tt.newMockCallback()); err != nil {
				t.Fatal(err)
			}
			if err := tt.e.ClearLineAll(); (err != nil) != tt.wantErr {
				t.Errorf("Editor.ClearLineAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.e.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_WindowRuneWidth(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name    string
		e       *publisher.Editor
		want    int
		wantErr bool
		setup   func(e *publisher.Editor, want int)
	}{
		{
			name:    "Valid case",
			e:       editor,
			want:    23,
			wantErr: false,
			setup: func(e *publisher.Editor, want int) {
				if err := e.SetViewSize(100, want); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.want)
			got, err := tt.e.WindowRuneWidth()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.WindowRuneWidth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Editor.WindowRuneWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEditor_WindowLineCount(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name    string
		e       *publisher.Editor
		want    int
		wantErr bool
		setup   func(e *publisher.Editor, want int)
	}{
		{
			name:    "Valid case",
			e:       editor,
			want:    23,
			wantErr: false,
			setup: func(e *publisher.Editor, want int) {
				if err := e.SetViewSize(want, 10); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.want)
			got, err := tt.e.WindowLineCount()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.WindowLineCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Editor.WindowLineCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEditor_CurrentRuneWidth(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name    string
		e       *publisher.Editor
		want    int
		wantErr bool
		setup   func(e *publisher.Editor, want int)
	}{
		{
			name:    "Valid case",
			e:       editor,
			want:    23,
			wantErr: false,
			setup: func(e *publisher.Editor, want int) {
				s := strings.Repeat("a", want)
				if err := e.Print(s); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.want)
			got, err := tt.e.CurrentRuneWidth()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.CurrentRuneWidth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Editor.CurrentRuneWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEditor_LineCount(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	tests := []struct {
		name    string
		e       *publisher.Editor
		want    int
		wantErr bool
		setup   func(e *publisher.Editor, want int)
	}{
		{
			name:    "First case",
			e:       editor,
			want:    0,
			wantErr: false,
			setup:   func(e *publisher.Editor, want int) {},
		},
		{
			name:    "After Print",
			e:       editor,
			want:    11,
			wantErr: false,
			setup: func(e *publisher.Editor, want int) {
				s := strings.Repeat("newline\n", want)
				if err := e.Print(s); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.e, tt.want)
			got, err := tt.e.LineCount()
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.LineCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Editor.LineCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEditor_PrintImage(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	type args struct {
		file       string
		widthInRW  int
		heightInLC int
	}
	tests := []struct {
		name    string
		e       *publisher.Editor
		args    args
		wantErr bool
		setup   func(*publisher.Editor)
	}{
		{
			name:    "First case",
			e:       editor,
			wantErr: false,
			args:    args{"/path/to/image/file", 10, 11},
		},
		{
			name:    "Succesive image",
			e:       editor,
			wantErr: false,
			args:    args{"/path/to/image/file2", 15, 20},
		},
		{
			name:    "text after image",
			e:       editor,
			wantErr: false,
			args:    args{"/path/to/image/file3", 0, 0},
			setup: func(e *publisher.Editor) {
				e.Print("some text")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if setup := tt.setup; setup != nil {
				setup(tt.e)
			}
			if err := tt.e.PrintImage(tt.args.file, tt.args.widthInRW, tt.args.heightInLC); (err != nil) != tt.wantErr {
				t.Errorf("Editor.PrintImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditor_MeasureImageSize(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	const testImagePath = "../../image/testdata/color.png"
	const printWidth = 10
	const textUnitW = 8
	const textUnitH = 14
	const printWPx = printWidth * textUnitW
	const expectW = printWidth
	var expectH = int(math.Ceil(float64(256) * float64(printWPx) / float64(512) / float64(textUnitH))) // 512x256 size

	type args struct {
		file       string
		widthInRW  int
		heightInLC int
	}
	tests := []struct {
		name     string
		e        *publisher.Editor
		args     args
		wantRetW int
		wantRetH int
		wantErr  bool
	}{
		{"measure test", editor, args{testImagePath, printWidth, 0}, expectW, expectH, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetW, gotRetH, err := tt.e.MeasureImageSize(tt.args.file, tt.args.widthInRW, tt.args.heightInLC)
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.MeasureImageSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRetW != tt.wantRetW {
				t.Errorf("Editor.MeasureImageSize() gotRetW = %v, want %v", gotRetW, tt.wantRetW)
			}
			if gotRetH != tt.wantRetH {
				t.Errorf("Editor.MeasureImageSize() gotRetH = %v, want %v", gotRetH, tt.wantRetH)
			}
		})
	}
}

func TestEditor_PrintImage_Published(t *testing.T) {
	testImagePath := "../../image/testdata/color.png"

	mustImageBoxFunc := func(theP *pubdata.Paragraph) *pubdata.Box_ImageData {
		t.Helper()
		if got, expect := len(theP.Lines), 1; got != expect {
			t.Fatalf("different line count, expect: %v, got: %v", expect, got)
		}
		theL := theP.Lines[0]
		if got, expect := len(theL.Boxes), 2; got < expect {
			t.Fatalf("different box count, expect: >=%v, got: %v", expect, got)
		}
		theB := theL.Boxes[1]
		imgB, ok := theB.Data.(*pubdata.Box_ImageData)
		if !ok {
			t.Fatalf("unexpected box type: expect: ImageBox, but %T", theB)
		}
		return imgB
	}

	for _, testcase := range []struct {
		name            string
		img             string
		opt             publisher.EditorOptions
		newMockCallback func(ctrl *gomock.Controller, img string) publisher.Callback
	}{
		{
			"img_fetch_none",
			testImagePath,
			publisher.EditorOptions{ImageFetchType: publisher.ImageFetchNone},
			func(ctrl *gomock.Controller, imgPath string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(theP *pubdata.Paragraph) error {
					imgB := mustImageBoxFunc(theP)
					if imgB.ImageData == nil {
						t.Error("image data should not be nil but nil")
					}
					if got, expect := imgB.ImageData.Source, imgPath; got != expect {
						t.Errorf("image source path is different, expect: %v, got: %v", expect, got)
					}
					if got, expect := imgB.ImageData.DataFetchType, publisher.ImageFetchNone; got != expect {
						t.Errorf("image fetch type should be ImageFetchNone(%v) but got %v", expect, got)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
		{
			"img_fetch_raw_rgba",
			testImagePath,
			publisher.EditorOptions{ImageFetchType: publisher.ImageFetchRawRGBA},
			func(ctrl *gomock.Controller, imgPath string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(theP *pubdata.Paragraph) error {
					imgB := mustImageBoxFunc(theP)
					if imgB.ImageData == nil {
						t.Error("image data should not be nil but nil")
					}
					if got, expect := imgB.ImageData.Source, imgPath; got != expect {
						t.Errorf("image source path is different, expect: %v, got: %v", expect, got)
					}
					if got, expect := imgB.ImageData.DataFetchType, publisher.ImageFetchRawRGBA; got != expect {
						t.Errorf("image fetch type should be ImageFetchRawRGBA(%v) but got %v", expect, got)
					}
					var sum uint64 = 0
					for _, b := range imgB.ImageData.Data {
						sum += uint64(b)
					}
					if got, expect := sum, 0; got == uint64(expect) {
						t.Errorf("image content is all zero, expect: >%v, got data sum: %v", expect, got)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
		{
			"img_fetch_png_encoded",
			testImagePath,
			publisher.EditorOptions{ImageFetchType: publisher.ImageFetchEncodedPNG},
			func(ctrl *gomock.Controller, imgPath string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(theP *pubdata.Paragraph) error {
					imgB := mustImageBoxFunc(theP)
					if imgB.ImageData == nil {
						t.Error("image data should not be nil but nil")
					}
					if got, expect := imgB.ImageData.Source, imgPath; got != expect {
						t.Errorf("image source path is different, expect: %v, got: %v", expect, got)
					}
					if got, expect := imgB.ImageData.DataFetchType, publisher.ImageFetchEncodedPNG; got != expect {
						t.Errorf("image fetch type should be ImageFetchEncodedPNG(%v) but got %v", expect, got)
					}
					pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
					if got, expect := len(imgB.ImageData.Data), len(pngMagic); got < expect {
						t.Fatalf("image data length shorter than png magic bytes, expect: >=%v, got: %v", expect, got)
					}
					var magicMatched = true
					for i, m := range pngMagic {
						magicMatched = magicMatched && (imgB.ImageData.Data[i] == m)
					}
					if got, expect := magicMatched, true; got != expect {
						t.Errorf("image magic not matched, expect: %v, got: %v", pngMagic, imgB.ImageData.Data[:len(pngMagic)])
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
		{
			"img_fetch_raw_rgba_but_not_found",
			"path/to/not/exist",
			publisher.EditorOptions{ImageFetchType: publisher.ImageFetchRawRGBA},
			func(ctrl *gomock.Controller, imgPath string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(1).DoAndReturn(func(theP *pubdata.Paragraph) error {
					imgB := mustImageBoxFunc(theP)
					if imgB.ImageData == nil {
						t.Error("image data should not be nil but nil")
					}
					if got, expect := imgB.ImageData.Source, imgPath; got != expect {
						t.Errorf("image source path is different, expect: %v, got: %v", expect, got)
					}
					if got, expect := imgB.ImageData.DataFetchType, publisher.ImageFetchRawRGBA; got != expect {
						t.Errorf("image fetch type should be ImageFetchRawRGBA(%v) but got %v", expect, got)
					}
					var sum uint64 = 0
					for _, b := range imgB.ImageData.Data {
						sum += uint64(b)
					}
					if got, expect := sum, 0; got > uint64(expect) {
						t.Errorf("image content should be all zero, but got data sum: %v", got)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
				cb.EXPECT().OnSync().Times(1)
				return cb
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			gs := setupGlobals(t, testcase.opt)
			editor := gs.editor
			ctrl := gs.ctrl
			defer gs.cancel()

			imgPath := testcase.img
			cb := testcase.newMockCallback(ctrl, imgPath)
			if err := editor.SetCallback(cb); err != nil {
				t.Fatalf("Can not set callbakc: %v", err)
			}
			if err := editor.PrintImage(imgPath, 10, 0); err != nil {
				t.Fatalf("Can not print image: %v", err)
			}
			// trigger publish
			if err := editor.Print("\n"); err != nil {
				t.Fatalf("Can not publish line with image: %v", err)
			}
			// Wait for completion
			if err := editor.Sync(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditor_PrintSpace(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	defer gs.cancel()

	const expectW = 12

	type args struct {
		widthInRW int
	}
	tests := []struct {
		name    string
		e       *publisher.Editor
		args    args
		wantErr bool
	}{
		{"print space", editor, args{expectW}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.e.PrintSpace(tt.args.widthInRW)
			if (err != nil) != tt.wantErr {
				t.Errorf("Editor.MeasureImageSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func BenchmarkPrintOnlyText(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	editor := publisher.NewEditor(ctx)
	defer editor.Close()
	if err := editor.SetViewSize(10, 100); err != nil {
		b.Fatal(err)
	}
	if err := editor.SetTextUnitPx(fixed.Point26_6{X: fixed.I(8), Y: fixed.I(14)}); err != nil {
		b.Fatal(err)
	}
	if err := editor.Sync(); err != nil {
		b.Fatal(err)
	}

	someText := `abcdefghijklmnopqrstuvwxyz`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		editor.Print(someText)
		editor.Print("\n")
		editor.Sync()
	}
}

func benchmarkHelperPrintOnlyImage(b *testing.B, opt publisher.EditorOptions) {
	b.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	editor := publisher.NewEditor(ctx, opt)
	defer editor.Close()
	if err := editor.SetViewSize(10, 100); err != nil {
		b.Fatal(err)
	}
	if err := editor.SetTextUnitPx(fixed.Point26_6{X: fixed.I(8), Y: fixed.I(14)}); err != nil {
		b.Fatal(err)
	}
	if err := editor.Sync(); err != nil {
		b.Fatal(err)
	}

	testImagePath := "../../image/testdata/color.png"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		editor.PrintImage(testImagePath, 10, 0)
		editor.Print("\n")
		editor.Sync()
	}
}

func BenchmarkPrintOnlyImageNone(b *testing.B) {
	benchmarkHelperPrintOnlyImage(b, publisher.EditorOptions{
		ImageFetchType: publisher.ImageFetchNone,
		ImageCacheSize: 1,
	})
}

func BenchmarkPrintOnlyImageRawRGBA(b *testing.B) {
	benchmarkHelperPrintOnlyImage(b, publisher.EditorOptions{
		ImageFetchType: publisher.ImageFetchRawRGBA,
		ImageCacheSize: 1,
	})
}

func BenchmarkPrintOnlyImageEncodedPNG(b *testing.B) {
	benchmarkHelperPrintOnlyImage(b, publisher.EditorOptions{
		ImageFetchType: publisher.ImageFetchEncodedPNG,
		ImageCacheSize: 1,
	})
}
