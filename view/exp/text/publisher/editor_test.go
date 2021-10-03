package publisher_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text/pubdata"
	"github.com/mzki/erago/view/exp/text/publisher"
	mock_publisher "github.com/mzki/erago/view/exp/text/publisher/mock"
)

type globals struct {
	editor *publisher.Editor
	ctrl   *gomock.Controller
	cancel context.CancelFunc
}

func setupGlobals(t *testing.T) struct {
	editor *publisher.Editor
	ctrl   *gomock.Controller
	cancel context.CancelFunc
} {
	ctx, cancel := context.WithCancel(context.Background())
	editor := publisher.NewEditor(ctx)
	if err := editor.SetViewSize(10, 100); err != nil {
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

	for _, testcase := range []struct {
		text            string
		newMockCallback func(s string) publisher.Callback
	}{
		{
			"newline\nnewline\n",
			func(s string) publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublish(gomock.Any()).Times(2).DoAndReturn(func(p *pubdata.Paragraph) error {
					if v := p.Lines.Len(); v != 1 {
						t.Logf("%#v", p)
						t.Fatalf("Paragraph should have 1 line but not. want: %v, got: %v", 1, v)
					}
					line := p.Lines.Get(0)
					if v := line.Boxes.Len(); v != 1 {
						t.Logf("%#v", line)
						t.Fatalf("Line should have 1 box but not. want: %v, got: %v", 1, v)
					}
					box := line.Boxes.Get(0)
					if v := box.ContentType(); v != pubdata.ContentTypeText {
						t.Logf("%#v", box)
						t.Fatalf("Box should have ContentTypeText but not. want: %v, got: %v", pubdata.ContentTypeText, v)
					}
					data := box.TextData()
					if data.Text != "newline" {
						t.Errorf("Box.Text not match. want: %v, got: %v", "newline", data.Text)
					}
					return nil
				})
				// Sync expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Return(nil)
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
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					doneCh <- struct{}{}
					return nil
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

	doneCh := make(chan struct{})
	defer close(doneCh)

	for _, testcase := range []struct {
		newMockCallback func() publisher.Callback
	}{
		{
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					<-doneCh // infinite loop
					return nil
				})
				return cb
			},
		},
	} {
		cb := testcase.newMockCallback()
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v", err)
		}
		if err := editor.Sync(); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Sync() with not respond from Callback, wantErr: %v, got: %v", context.DeadlineExceeded, err)
		}
	}
}

func TestEditorSyncError(t *testing.T) {
	gs := setupGlobals(t)
	editor := gs.editor
	ctrl := gs.ctrl
	defer gs.cancel()

	doneCh := make(chan struct{}, 1)

	var errOnPublishTemporary = errors.New("errOnPublishTemporary")

	for _, testcase := range []struct {
		newMockCallback func() publisher.Callback
		err             error
	}{
		{
			func() publisher.Callback {
				cb := mock_publisher.NewMockCallback(ctrl)
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).DoAndReturn(func(*pubdata.Paragraph) error {
					doneCh <- struct{}{}
					return errOnPublishTemporary
				})
				return cb
			},
			errOnPublishTemporary,
		},
	} {
		cb := testcase.newMockCallback()
		if err := editor.SetCallback(cb); err != nil {
			t.Fatalf("Can not set callbakc: %v", err)
		}
		if err := editor.Sync(); !errors.Is(err, testcase.err) {
			t.Errorf("Unexpected Sync() error: want: %v, got:%v", testcase.err, err)
		}
		select {
		case <-doneCh:
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
					if v := p.Lines.Len(); v != 1 {
						t.Fatalf("Paragraph should have 1 line but not. want: %v, got: %v", 1, v)
					}
					line := p.Lines.Get(0)
					if v := line.Boxes.Len(); v != 1 {
						t.Logf("%#v", p)
						t.Fatalf("Line should have 1 box but not. want: %v, got: %v", 1, v)
					}
					box := line.Boxes.Get(0)
					expect := int(args.color)
					if v := box.TextData().FgColor; v != expect {
						t.Errorf("Published paragraph should have color by SetColor(). want: %v, got: %v", expect, v)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
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
					expect := pubdata.AlignmentString(args.align)
					if v := p.Alignment; v != expect {
						t.Errorf("Published paragraph should have alignment by SetAlignment(). want: %v, got: %v", expect, v)
					}
					return nil
				})
				// Sync() expectation
				cb.EXPECT().OnPublishTemporary(gomock.Any()).Times(1).Return(nil)
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
