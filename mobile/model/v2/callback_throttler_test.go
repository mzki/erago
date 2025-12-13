package model

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mzki/erago/uiadapter"
	"github.com/mzki/erago/view/exp/text/pubdata"
)

type uiStub struct {
	OnPublishBytesFunc          func(bs []byte) error
	OnPublishBytesTemporaryFunc func(bs []byte) error
	OnRemoveFunc                func(nParagraph int) error
	OnRemoveAllFunc             func() error
	OnDebugTimestampFunc        func(publishId int64, timestamp string, epochTimeNano int64) error
	OnCommandRequestedFunc      func()
	OnInputRequestedFunc        func()
	OnInputRequestClosedFunc    func()
}

func (ui *uiStub) OnPublishBytes(bs []byte) error {
	if f := ui.OnPublishBytesFunc; f != nil {
		return f(bs)
	}
	return nil
}

func (ui *uiStub) OnPublishBytesTemporary(bs []byte) error {
	if f := ui.OnPublishBytesTemporaryFunc; f != nil {
		return f(bs)
	}
	return nil
}

func (ui *uiStub) OnRemove(nParagraph int) error {
	if f := ui.OnRemoveFunc; f != nil {
		return f(nParagraph)
	}
	return nil
}

func (ui *uiStub) OnRemoveAll() error {
	if f := ui.OnRemoveAllFunc; f != nil {
		return f()
	}
	return nil
}

func (ui *uiStub) OnDebugTimestamp(publishId int64, timestamp string, epochTimeNano int64) error {
	if f := ui.OnDebugTimestampFunc; f != nil {
		return f(publishId, timestamp, epochTimeNano)
	}
	return nil

}

func (ui *uiStub) OnCommandRequested() {
	if f := ui.OnCommandRequestedFunc; f != nil {
		f()
	}
}

func (ui *uiStub) OnInputRequested() {
	if f := ui.OnInputRequestedFunc; f != nil {
		f()
	}
}

func (ui *uiStub) OnInputRequestClosed() {
	if f := ui.OnInputRequestClosedFunc; f != nil {
		f()
	}
}

func newCallbackThrottlerForTest(ctx context.Context, d time.Duration, ui UI) *CallbackThrottler {
	return NewCallbackThrottler(ctx, d, ui, newParagraphListBinaryEncodeFunc(MessageByteEncodingProtobuf))
}

func TestNewCallbackThrottler(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx     context.Context
		d       time.Duration
		ui      UI
		encoder paragraphListBinaryEncoderFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{ctx: context.Background(), d: 1 * time.Second, ui: &stubUI{}, encoder: newParagraphListBinaryEncodeFunc(MessageByteEncodingProtobuf)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = NewCallbackThrottler(tt.args.ctx, tt.args.d, tt.args.ui, tt.args.encoder)
		})
	}
}

func TestCallbackThrottler_StartThrottleAndClose(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
	}{
		{
			name: "start and close",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				d := 100 * time.Millisecond
				ui := &uiStub{}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 3*time.Second)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			if thr.IsClosed() {
				t.Fatal("CallbackThrottler somehow closed")
			}
			err := thr.Close()
			if err != nil {
				t.Errorf("%v", err)
			}
			if !thr.IsClosed() {
				t.Fatal("CallbackThrottler closed but somehow is not closed")
			}
		})
	}
}

func TestCallbackThrottler_OnPublish(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct {
		p func() *pubdata.Paragraph
	}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
		want2ndErr   bool
	}{
		{
			name: "publish success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnPublishBytesFunc: func(bs []byte) error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{p: func() *pubdata.Paragraph {
				return &pubdata.Paragraph{Id: 1, Lines: []*pubdata.Line{}, Alignment: pubdata.Alignment_ALIGNMENT_LEFT, Fixed: true}
			}},
		},
		{
			name: "publish error",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnPublishBytesFunc: func(_ []byte) error {
						recvCh <- struct{}{}
						return fmt.Errorf("failed!!!")
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{p: func() *pubdata.Paragraph {
				return &pubdata.Paragraph{Id: 1, Lines: []*pubdata.Line{}, Alignment: pubdata.Alignment_ALIGNMENT_LEFT, Fixed: true}
			}},
			wantErr:    false,
			want2ndErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err := thr.OnPublish(tt.args.p()); (err != nil) != tt.wantErr {
				t.Errorf("CallbackThrottler.OnPublish() error = %v, wantErr %v", err, tt.wantErr)
			}
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve singal from UI, missing call UI.OnPublishBytes?")
			}

			if tt.want2ndErr {
				// ui side error will be get from 2nd call
				if err := thr.OnPublish(tt.args.p()); err == nil {
					t.Errorf("UI.OnPublish retured error but 2nd call of CallbackThrottler.OnPoublsh still no error")
				}
			}
		})
	}
}

func TestCallbackThrottler_OnPublishTemporary(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct {
		p func() *pubdata.Paragraph
	}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
	}{
		{
			name: "publish at PublishBytes instead PublishBytesTemporary",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					// OnPublishBytesTemprary is never used in CallbackThrottler. instead called OnPublishBytes
					OnPublishBytesFunc: func(bs []byte) error {
						recvCh <- struct{}{}
						return nil
					},
					OnPublishBytesTemporaryFunc: func(bs []byte) error {
						return fmt.Errorf("never called this!!!")
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{p: func() *pubdata.Paragraph {
				return &pubdata.Paragraph{Id: 1, Lines: []*pubdata.Line{}, Alignment: pubdata.Alignment_ALIGNMENT_LEFT, Fixed: true}
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err := thr.OnPublishTemporary(tt.args.p()); (err != nil) != tt.wantErr {
				t.Errorf("CallbackThrottler.OnPublishTemporary() error = %v, wantErr %v", err, tt.wantErr)
			}
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve signal from UI, missing call UI.OnPublishBytes?")
			}

			// check whether error is not happened at 2nd call
			if err := thr.OnPublishTemporary(tt.args.p()); err != nil {
				t.Errorf("CallbackThrottler.OnPublishTemporary() error = %v", err)
			}
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve signal from UI, missing call UI.OnPublishBytes?")
			}
		})
	}
}

func TestCallbackThrottler_OnRemove(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct {
		nParagraph int
	}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
		want2ndErr   bool
	}{
		{
			name: "remove success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveFunc: func(nCount int) error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{nParagraph: 1},
		},
		{
			name: "remove error",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveFunc: func(_ int) error {
						recvCh <- struct{}{}
						return fmt.Errorf("failed!!!")
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args:       args{nParagraph: 2},
			wantErr:    false,
			want2ndErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err := thr.OnRemove(tt.args.nParagraph); (err != nil) != tt.wantErr {
				t.Errorf("CallbackThrottler.OnRemove() error = %v, wantErr %v", err, tt.wantErr)
			}
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve singal from UI, missing call UI.OnRemove?")
			}

			if tt.want2ndErr {
				// ui side error will be get from 2nd call
				if err := thr.OnRemove(tt.args.nParagraph); err == nil {
					t.Errorf("UI returned error but 2nd call of CallbackThrottler.OnRemove still no error")
				}
			}
		})
	}
}

func TestCallbackThrottler_OnRemoveAll(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct{}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
		want2ndErr   bool
	}{
		{
			name: "remove all success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{},
		},
		{
			name: "remove all error",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return fmt.Errorf("failed!!!")
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args:       args{},
			wantErr:    false,
			want2ndErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err := thr.OnRemoveAll(); (err != nil) != tt.wantErr {
				t.Errorf("CallbackThrottler.OnRemoveAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve singal from UI, missing call UI.OnRemoveAll?")
			}

			if tt.want2ndErr {
				// ui side error will be get from 2nd call
				if err := thr.OnRemoveAll(); err == nil {
					t.Errorf("UI returned error but 2nd call of CallbackThrottler.OnRemoveAll still no error")
				}
			}
		})
	}
}

func TestCallbackThrottler_OnSync(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct{}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
		want2ndErr   bool
	}{
		{
			name: "sync success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{},
		},
		{
			name: "sync error",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				d := throttlePeriod
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return fmt.Errorf("failed!!!")
					},
				}
				return newCallbackThrottlerForTest(ctx, d, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args:       args{},
			wantErr:    false,
			want2ndErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err := thr.OnRemoveAll(); (err != nil) != tt.wantErr {
				t.Errorf("CallbackThrottler.OnRemoveAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := thr.OnSync(); (err != nil) != tt.want2ndErr {
				t.Errorf("CallbackThrottler.OnSync() error = %v, wantErr %v", err, tt.wantErr)
			}
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve singal from UI, missing call UI.OnRemoveAll?")
			}
		})
	}
}

func TestCallbackThrottler_OnSync_Period(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct{}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
	}{
		{
			name: "measure sync period",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			syncStep := func() {
				t.Helper()
				if err := thr.OnRemoveAll(); err != nil {
					t.Fatalf("CallbackThrottler.OnRemoveAll() error = %v", err)
				}
				if err := thr.OnSync(); err != nil {
					t.Fatalf("CallbackThrottler.OnSync() error = %v", err)
				}
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				select {
				case <-recvCh:
					// success
				case <-ctx.Done():
					t.Error("Timeout to recieve singal from UI, missing call UI.OnRemoveAll?")
				}
			}

			// dry run 1st time
			syncStep()

			const nSampleSize = 10 // 100ms * 10 = 1sec, make sure this time is less than context timeout.
			deltaList := make([]time.Duration, 0, nSampleSize)
			for i := 0; i < 10; i++ {
				now := time.Now()
				syncStep()
				delta := time.Since(now)
				deltaList = append(deltaList, delta)
			}
			sum := time.Duration(0)
			for _, d := range deltaList {
				sum += d
			}
			ave := float64(sum) / float64(len(deltaList))
			if dExp := throttlePeriod / 2; ave >= float64(dExp) {
				dAve := time.Duration(ave)
				t.Errorf("Sync period time is not matched with expected period, expect %v < %v", dAve, dExp)
			}
		})
	}
}

func TestCallbackThrottler_Publish_Period(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct{}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
	}{
		{
			name: "measure publish period",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnRemoveAllFunc: func() error {
						recvCh <- struct{}{}
						return nil
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			syncStep := func() {
				t.Helper()
				if err := thr.OnRemoveAll(); err != nil {
					t.Fatalf("CallbackThrottler.OnRemoveAll() error = %v", err)
				}
				// do not call Sync to wait to publish event at throttle period.
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				select {
				case <-recvCh:
					// success
				case <-ctx.Done():
					t.Error("Timeout to recieve singal from UI, missing call UI.OnRemoveAll?")
				}
			}

			// dry run 1st time
			syncStep()

			const nSampleSize = 10 // 100ms * 10 = 1sec, make sure this time is less than context timeout.
			deltaList := make([]time.Duration, 0, nSampleSize)
			for i := 0; i < 10; i++ {
				now := time.Now()
				syncStep()
				delta := time.Since(now)
				deltaList = append(deltaList, delta)
			}
			sum := time.Duration(0)
			for _, d := range deltaList {
				sum += d
			}
			ave := float64(sum) / float64(len(deltaList))
			if ave < float64(throttlePeriod)*0.9 || float64(throttlePeriod)*1.1 < ave {
				dAve := time.Duration(ave)
				t.Errorf("Sync period time is not matched with expected period, expect %v == %v", dAve, throttlePeriod)
			}
		})
	}
}

func TestCallbackThrottler_OnRequestChanged(t *testing.T) {
	t.Parallel()
	type keyRecvChType string
	const keyRecvCh = keyRecvChType("recvCh")
	const throttlePeriod = 100 * time.Millisecond
	type args struct {
		req uiadapter.InputRequestType
	}
	tests := []struct {
		name         string
		newThrottler func(ctx context.Context) *CallbackThrottler
		newCtx       func() (context.Context, context.CancelFunc)
		args         args
		wantErr      bool
	}{
		{
			name: "command request changed success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnCommandRequestedFunc: func() {
						recvCh <- struct{}{}
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{req: uiadapter.InputRequestCommand},
		},
		{
			name: "raw ipnut request changed success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnCommandRequestedFunc: func() {
						recvCh <- struct{}{}
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{req: uiadapter.InputRequestRawInput},
		},
		{
			name: "input request changed success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnInputRequestedFunc: func() {
						recvCh <- struct{}{}
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{req: uiadapter.InputRequestInput},
		},
		{
			name: "none request changed success",
			newThrottler: func(ctx context.Context) *CallbackThrottler {
				recvCh := ctx.Value(keyRecvCh).(chan struct{})
				ui := &uiStub{
					OnInputRequestClosedFunc: func() {
						recvCh <- struct{}{}
					},
				}
				return newCallbackThrottlerForTest(ctx, throttlePeriod, ui)
			},
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx := context.WithValue(context.Background(), keyRecvCh, make(chan struct{}, 1)) // 1 buffer to avoid blocking
				return context.WithTimeout(ctx, 3*time.Second)
			},
			args: args{req: uiadapter.InputRequestNone},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()
			thr := tt.newThrottler(ctx)
			thr.StartThrottle()
			defer func() {
				if err := thr.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			thr.OnRequestChanged(tt.args.req)
			recvCh := ctx.Value(keyRecvCh).(chan struct{})
			select {
			case <-recvCh:
				// success
			case <-ctx.Done():
				t.Error("Timeout to recieve singal from UI, missing call UI.OnXXXRequested?")
			}
		})
	}
}

func TestCallbackThrottler_OnXXX_Closed(t *testing.T) {
	t.Parallel()

	const throttlePeriod = 100 * time.Millisecond

	t.Run("OnPublish", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		err = thr.OnPublish(newParagraphForTest(1, true))
		if err == nil {
			t.Errorf("Called OnXXX at Closed CallbackThrottler, but returns no error.")
		}
	})

	t.Run("OnPublishTemporary", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		err = thr.OnPublishTemporary(newParagraphForTest(1, false))
		if err == nil {
			t.Errorf("Called OnXXX at Closed CallbackThrottler, but returns no error.")
		}
	})

	t.Run("OnRemove", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		err = thr.OnRemove(3)
		if err == nil {
			t.Errorf("Called OnXXX at Closed CallbackThrottler, but returns no error.")
		}
	})

	t.Run("OnRemoveAll", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		err = thr.OnRemoveAll()
		if err == nil {
			t.Errorf("Called OnXXX at Closed CallbackThrottler, but returns no error.")
		}
	})

	t.Run("OnSync", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		err = thr.OnSync()
		if err == nil {
			t.Errorf("Called OnXXX at Closed CallbackThrottler, but returns no error.")
		}
	})

	t.Run("OnRequestChanged", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		thr := newCallbackThrottlerForTest(ctx, throttlePeriod, &stubUI{})
		thr.StartThrottle()
		err := thr.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		thr.OnRequestChanged(uiadapter.InputRequestCommand)
		// no returns for this API. just to confirm no panic.
	})
}

func Test_newPendingEvents(t *testing.T) {
	tests := []struct {
		name string
		want *pendingEvents
	}{
		// TODO: Add test cases. there is no
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPendingEvents(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPendingEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newParagraphForTest(id int64, fixed bool) *pubdata.Paragraph {
	return &pubdata.Paragraph{
		Id: id,
		Lines: []*pubdata.Line{
			{
				Boxes: []*pubdata.Box{
					{
						RuneWidth:     0,
						LineCountHint: 1,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_SPACE,
						Data:          &pubdata.Box_SpaceData{},
					},
				},
				RuneWidth: 0,
			},
		},
		Alignment: pubdata.Alignment_ALIGNMENT_LEFT,
		Fixed:     fixed,
	}
}

func Test_pendingEvents_appendAddParagraph(t *testing.T) {
	type args struct {
		p func() *pubdata.Paragraph
	}
	tests := []struct {
		name             string
		newPendingEvents func() *pendingEvents
		args             args
		verify           func(t *testing.T, pe *pendingEvents, args args)
	}{
		{
			name: "append at empty",
			newPendingEvents: func() *pendingEvents {
				return newPendingEvents()
			},
			args: args{
				p: func() *pubdata.Paragraph { return newParagraphForTest(1, true) },
			},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new paragraph but not in events. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				internalP := paragraphList[0]
				argP := args.p()
				if !reflect.DeepEqual(internalP, argP) {
					t.Errorf("internal paragraph and appended one is different. got %v, expect %v", internalP, argP)
				}
			},
		},
		{
			name: "append at paragraph",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendAddParagraph(newParagraphForTest(1, true))
				return pe
			},
			args: args{
				p: func() *pubdata.Paragraph { return newParagraphForTest(2, true) },
			},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new paragraph at paragraph, should be merged into one event. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				if len(paragraphList) != 2 {
					t.Fatalf("appended new paragraph at paragraphList, should have 2 paragraph in the list. got %v, expect %v", len(paragraphList), 2)
				}
				internalP := paragraphList[1]
				argP := args.p()
				if !reflect.DeepEqual(internalP, argP) {
					t.Errorf("internal paragraph and appended one is different. got %v, expect %v", internalP, argP)
				}
			},
		},
		{
			name: "append at remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraph(1)
				return pe
			},
			args: args{
				p: func() *pubdata.Paragraph { return newParagraphForTest(1, true) },
			},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new paragraph at remove, should have different events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				if len(paragraphList) != 1 {
					t.Fatalf("appended new paragraph at paragraphList, should have 1 paragraph in the list. got %v, expect %v", len(paragraphList), 1)
				}
				internalP := paragraphList[0]
				argP := args.p()
				if !reflect.DeepEqual(internalP, argP) {
					t.Errorf("internal paragraph and appended one is different. got %v, expect %v", internalP, argP)
				}
			},
		},
		{
			name: "append at removeAll",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll()
				return pe
			},
			args: args{
				p: func() *pubdata.Paragraph { return newParagraphForTest(1, true) },
			},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new paragraph at remove all, should have different events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				if len(paragraphList) != 1 {
					t.Fatalf("appended new paragraph at paragraphList, should have 1 paragraph in the list. got %v, expect %v", len(paragraphList), 1)
				}
				internalP := paragraphList[0]
				argP := args.p()
				if !reflect.DeepEqual(internalP, argP) {
					t.Errorf("internal paragraph and appended one is different. got %v, expect %v", internalP, argP)
				}
			},
		},
		{
			name: "append at input request",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendInputRequest(uiadapter.InputRequestCommand)
				return pe
			},
			args: args{
				p: func() *pubdata.Paragraph { return newParagraphForTest(1, true) },
			},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new paragraph at input request, should have different events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				if len(paragraphList) != 1 {
					t.Fatalf("appended new paragraph at paragraphList, should have 1 paragraph in the list. got %v, expect %v", len(paragraphList), 1)
				}
				internalP := paragraphList[0]
				argP := args.p()
				if !reflect.DeepEqual(internalP, argP) {
					t.Errorf("internal paragraph and appended one is different. got %v, expect %v", internalP, argP)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := tt.newPendingEvents()
			pe.appendAddParagraph(tt.args.p())
			tt.verify(t, pe, tt.args)
		})
	}
}

func Test_pendingEvents_appendRemoveParagraph(t *testing.T) {
	type args struct {
		nCount int
	}
	tests := []struct {
		name             string
		newPendingEvents func() *pendingEvents
		args             args
		verify           func(t *testing.T, pe *pendingEvents, args args)
	}{
		{
			name: "append at empty",
			newPendingEvents: func() *pendingEvents {
				return newPendingEvents()
			},
			args: args{nCount: 2},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event but not in events. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeRemoveCount {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveCount)
				}
				nCount := ev.(*peEventRemoveCount).getRawData()
				if !reflect.DeepEqual(nCount, args.nCount) {
					t.Errorf("internal value and appended one is different. got %v, expect %v", nCount, args.nCount)
				}
			},
		},
		{
			name: "append at first paragraph(any), will be remained",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendAddParagraph(newParagraphForTest(3, true))
				return pe
			},
			args: args{nCount: 3},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event at first paragraph, should have different events. got %v, expect %v", len(pe.events), 2)
				}
				if ev := pe.events[0]; ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeRemoveCount {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveCount)
				}
				nCount := ev.(*peEventRemoveCount).getRawData()
				if nCount != args.nCount {
					t.Errorf("remove count is different. got %v, expect %v", nCount, args.nCount)
				}
			},
		},
		{
			name: "append at paragraph, paragraph > remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll() // as first event
				pe.appendAddParagraph(newParagraphForTest(1, true))
				pe.appendAddParagraph(newParagraphForTest(2, true))
				return pe
			},
			args: args{nCount: 1},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event but not in events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				paragraphList := ev.(*peEventParagraphList).getRawData()
				if len(paragraphList) != 1 {
					t.Errorf("merged paragraphList length is different. got %v, expect %v", len(paragraphList), 1)
				}
			},
		},
		{
			name: "append at paragraph, paragraph == remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll() // as first event
				pe.appendAddParagraph(newParagraphForTest(1, true))
				return pe
			},
			args: args{nCount: 1},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event but is merged. should be empty events. got %v, expect %v", len(pe.events), 1)
				}
			},
		},
		{
			name: "append at paragraph, paragraph < remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll() // as first event
				pe.appendAddParagraph(newParagraphForTest(1, true))
				return pe
			},
			args: args{nCount: 2},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event but not in events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeRemoveCount {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveCount)
				}
				nCount := ev.(*peEventRemoveCount).getRawData()
				if nCount != 1 {
					t.Errorf("merged remove request, the rest of count is different. got %v, expect %v", nCount, 1)
				}
			},
		},
		{
			name: "append at remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll() // as first event
				pe.appendRemoveParagraph(1)
				return pe
			},
			args: args{nCount: 3},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event at remove, should merge into one. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeRemoveCount {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveCount)
				}
				nCount := ev.(*peEventRemoveCount).getRawData()
				if nCount != args.nCount+1 {
					t.Errorf("merged remove count is different. got %v, expect %v", nCount, args.nCount+1)
				}
			},
		},
		{
			name: "append at removeAll",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendInputRequest(uiadapter.InputRequestCommand) // as first event
				pe.appendRemoveParagraphAll()
				return pe
			},
			args: args{nCount: 3},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event at remove all, should merge into one. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			name: "append at input request",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll() // as first event
				pe.appendInputRequest(uiadapter.InputRequestCommand)
				return pe
			},
			args: args{nCount: 3},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 3 {
					t.Fatalf("appended new event at input request, should have different events. got %v, expect %v", len(pe.events), 3)
				}
				ev := pe.events[2]
				if ev.Type() != peEventTypeRemoveCount {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveCount)
				}
				nCount := ev.(*peEventRemoveCount).getRawData()
				if nCount != args.nCount {
					t.Errorf("remove count is different. got %v, expect %v", nCount, args.nCount)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := tt.newPendingEvents()
			pe.appendRemoveParagraph(tt.args.nCount)
			tt.verify(t, pe, tt.args)
		})
	}
}

func Test_pendingEvents_appendRemoveParagraphAll(t *testing.T) {
	type args struct{}
	tests := []struct {
		name             string
		newPendingEvents func() *pendingEvents
		args             args
		verify           func(t *testing.T, pe *pendingEvents, args args)
	}{
		{
			name: "append at empty",
			newPendingEvents: func() *pendingEvents {
				return newPendingEvents()
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event but not in events. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			name: "append at paragraph, remove all always stay itself only",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendAddParagraph(newParagraphForTest(1, true))
				pe.appendAddParagraph(newParagraphForTest(2, true))
				return pe
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event but not in events. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			name: "append at remove",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraph(1)
				return pe
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event at remove, should merge into one. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			name: "append at removeAll",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendRemoveParagraphAll()
				return pe
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 1 {
					t.Fatalf("appended new event at remove all, should merge into one. got %v, expect %v", len(pe.events), 1)
				}
				ev := pe.events[0]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			name: "append at input request",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendInputRequest(uiadapter.InputRequestCommand)
				return pe
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 2 {
					t.Fatalf("appended new event at input request, should have different events. got %v, expect %v", len(pe.events), 2)
				}
				ev := pe.events[1]
				if ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
		{
			// This case, remove all is stopped at input command type, so paragraph will be remained.
			// This is due to input command type is different from paragraph modifucation.
			name: "append at input request, prepend paragrah",
			newPendingEvents: func() *pendingEvents {
				pe := newPendingEvents()
				pe.appendAddParagraph(newParagraphForTest(12, true))
				pe.appendInputRequest(uiadapter.InputRequestCommand)
				return pe
			},
			args: args{},
			verify: func(t *testing.T, pe *pendingEvents, args args) {
				if len(pe.events) != 3 {
					t.Fatalf("appended new event at input request, should have different events. got %v, expect %v", len(pe.events), 3)
				}
				if ev := pe.events[0]; ev.Type() != peEventTypeParagraphList {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeParagraphList)
				}
				if ev := pe.events[1]; ev.Type() != peEventTypeInputRequest {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeInputRequest)
				}
				if ev := pe.events[2]; ev.Type() != peEventTypeRemoveAll {
					t.Fatalf("events have unexpected type. got %v, expect %v", ev.Type(), peEventTypeRemoveAll)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := tt.newPendingEvents()
			pe.appendRemoveParagraphAll()
			tt.verify(t, pe, tt.args)
		})
	}
}

func Test_pendingEvents_appendInputRequest(t *testing.T) {
	t.Parallel()

	t.Run("empty -> append input request", func(t *testing.T) {
		pe := newPendingEvents()
		pe.appendInputRequest(uiadapter.InputRequestInput)

		if len(pe.events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(pe.events))
		}
		ev := pe.events[0]
		if ev.Type() != peEventTypeInputRequest {
			t.Fatalf("expected event type InputRequest, got %v", ev.Type())
		}
		got := ev.(*peEventInputRequest).getRawData()
		if got != uiadapter.InputRequestInput {
			t.Errorf("unexpected input request value: got %v want %v", got, uiadapter.InputRequestInput)
		}
	})

	t.Run("last is paragraph -> append command request", func(t *testing.T) {
		pe := newPendingEvents()
		paragraphs := make([]*pubdata.Paragraph, 0, 1)
		paragraphs = append(paragraphs, newParagraphForTest(1, true))
		pe.events = append(pe.events, (*peEventParagraphList)(&paragraphs))

		pe.appendInputRequest(uiadapter.InputRequestCommand)

		if len(pe.events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(pe.events))
		}
		if pe.events[0].Type() != peEventTypeParagraphList {
			t.Fatalf("first event should be ParagraphList, got %v", pe.events[0].Type())
		}
		if pe.events[1].Type() != peEventTypeInputRequest {
			t.Fatalf("second event should be InputRequest, got %v", pe.events[1].Type())
		}
		got := pe.events[1].(*peEventInputRequest).getRawData()
		if got != uiadapter.InputRequestCommand {
			t.Errorf("unexpected input request value: got %v want %v", got, uiadapter.InputRequestCommand)
		}
	})

	t.Run("last is RemoveCount -> append InputRequestInput", func(t *testing.T) {
		pe := newPendingEvents()
		nCount := 10
		pe.events = append(pe.events, (*peEventRemoveCount)(&nCount))

		pe.appendInputRequest(uiadapter.InputRequestInput)

		if len(pe.events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(pe.events))
		}
		first := pe.events[0]
		if first.Type() != peEventTypeRemoveCount {
			t.Fatalf("first event should be RemoveCount, got %v", first.Type())
		}
		if first.(*peEventRemoveCount).getRawData() != nCount {
			t.Errorf("first Remove Count expected to %v, got %v", nCount, first.(*peEventRemoveCount).getRawData())
		}
		second := pe.events[1]
		if second.Type() != peEventTypeInputRequest {
			t.Fatalf("second event should be InputRequest, got %v", second.Type())
		}
		if second.(*peEventInputRequest).getRawData() != uiadapter.InputRequestInput {
			t.Errorf("second input request expected Input, got %v", second.(*peEventInputRequest).getRawData())
		}
	})

	t.Run("last is RemoveAll -> append InputRequestInput", func(t *testing.T) {
		pe := newPendingEvents()
		pe.appendRemoveParagraphAll()

		pe.appendInputRequest(uiadapter.InputRequestInput)

		if len(pe.events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(pe.events))
		}
		first := pe.events[0]
		if first.Type() != peEventTypeRemoveAll {
			t.Fatalf("first event should be RemoveAll, got %v", first.Type())
		}
		second := pe.events[1]
		if second.Type() != peEventTypeInputRequest {
			t.Fatalf("second event should be InputRequest, got %v", second.Type())
		}
		if second.(*peEventInputRequest).getRawData() != uiadapter.InputRequestInput {
			t.Errorf("second input request expected Input, got %v", second.(*peEventInputRequest).getRawData())
		}
	})

	t.Run("last is InputRequestNone -> append InputRequestInput (no merge)", func(t *testing.T) {
		pe := newPendingEvents()
		rNone := uiadapter.InputRequestNone
		pe.events = append(pe.events, (*peEventInputRequest)(&rNone))

		pe.appendInputRequest(uiadapter.InputRequestInput)

		if len(pe.events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(pe.events))
		}
		first := pe.events[0]
		if first.Type() != peEventTypeInputRequest {
			t.Fatalf("first event should be InputRequest, got %v", first.Type())
		}
		if first.(*peEventInputRequest).getRawData() != uiadapter.InputRequestNone {
			t.Errorf("first input request expected None, got %v", first.(*peEventInputRequest).getRawData())
		}
		second := pe.events[1]
		if second.Type() != peEventTypeInputRequest {
			t.Fatalf("second event should be InputRequest, got %v", second.Type())
		}
		if second.(*peEventInputRequest).getRawData() != uiadapter.InputRequestInput {
			t.Errorf("second input request expected Input, got %v", second.(*peEventInputRequest).getRawData())
		}
	})

	t.Run("last is InputRequestCommand -> append InputRequestCommand (merge removes last)", func(t *testing.T) {
		pe := newPendingEvents()
		rCmd := uiadapter.InputRequestCommand
		pe.events = append(pe.events, (*peEventInputRequest)(&rCmd))

		pe.appendInputRequest(uiadapter.InputRequestCommand)

		if len(pe.events) != 1 {
			t.Fatalf("expected 1 events after merge-removal, got %d", len(pe.events))
		}
	})

	t.Run("earlier InputRequestCommand and last Paragraph -> append InputRequestCommand, remains both input events", func(t *testing.T) {
		pe := newPendingEvents()
		rCmd := uiadapter.InputRequestCommand
		paragraphs := make([]*pubdata.Paragraph, 0, 1)
		paragraphs = append(paragraphs, newParagraphForTest(2, true))

		pe.events = append(pe.events, (*peEventInputRequest)(&rCmd))
		pe.events = append(pe.events, (*peEventParagraphList)(&paragraphs))

		pe.appendInputRequest(uiadapter.InputRequestNone)

		if len(pe.events) != 3 {
			t.Fatalf("expected 3 events after appending input request, got %d", len(pe.events))
		}
		if pe.events[0].Type() != peEventTypeInputRequest {
			t.Fatalf("expected remaining event to be InputRequest, got %v", pe.events[0].Type())
		}
		if pe.events[1].Type() != peEventTypeParagraphList {
			t.Fatalf("expected remaining event to be ParagraphList, got %v", pe.events[1].Type())
		}
		if pe.events[2].Type() != peEventTypeInputRequest {
			t.Fatalf("expected remaining event to be InputRequest, got %v", pe.events[2].Type())
		}
		// ensure paragraph content unchanged
		gotParagraphs := pe.events[1].(*peEventParagraphList).getRawData()
		if !reflect.DeepEqual(gotParagraphs[0], paragraphs[0]) {
			t.Errorf("paragraph changed unexpectedly")
		}
	})
}

func Test_pendingEvents_publish(t *testing.T) {
	type fields struct {
		events []peEvent
	}
	type args struct {
		ui      UI
		encoder paragraphListBinaryEncoderFunc
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &pendingEvents{
				events: tt.fields.events,
			}
			if err := pe.publish(tt.args.ui, tt.args.encoder); (err != nil) != tt.wantErr {
				t.Errorf("pendingEvents.publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_peEventParagraphList_MergeParagraph(t *testing.T) {
	type args struct {
		p *pubdata.Paragraph
	}
	tests := []struct {
		name       string
		plist      *peEventParagraphList
		args       args
		wantMerged bool
	}{
		{
			name:       "fixed paragraph + fixed paragraph (sameID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, true)},
			args:       args{newParagraphForTest(2, true)},
			wantMerged: true,
		},
		{
			name:       "fixed paragraph + fixed paragraph (differentID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, true)},
			args:       args{newParagraphForTest(3, true)},
			wantMerged: true,
		},
		{
			name:       "fixed paragraph + un-fixed paragraph (sameID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, true)},
			args:       args{newParagraphForTest(2, false)},
			wantMerged: true,
		},
		{
			name:       "fixed paragraph + un-fixed paragraph (differentID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, true)},
			args:       args{newParagraphForTest(3, false)},
			wantMerged: true,
		},
		{
			name:       "un-fixed paragraph + fixed paragraph (sameID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, false)},
			args:       args{newParagraphForTest(2, true)},
			wantMerged: true,
		},
		{
			name:       "un-fixed paragraph + fixed paragraph (differentID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, false)},
			args:       args{newParagraphForTest(3, true)},
			wantMerged: true,
		},
		{
			name:       "un-fixed paragraph + un-fixed paragraph (sameID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, false)},
			args:       args{newParagraphForTest(2, false)},
			wantMerged: true,
		},
		{
			name:       "un-fixed paragraph + un-fixed paragraph (differentID)",
			plist:      &peEventParagraphList{newParagraphForTest(2, false)},
			args:       args{newParagraphForTest(3, false)},
			wantMerged: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMerged := tt.plist.MergeParagraph(tt.args.p); gotMerged != tt.wantMerged {
				t.Errorf("peEventParagraphList.MergeParagraph() = %v, want %v", gotMerged, tt.wantMerged)
			}
		})
	}
}

func Test_peEventParagraphList_MergeRemoveCount(t *testing.T) {
	type args struct {
		nCount int
	}
	tests := []struct {
		name           string
		plist          *peEventParagraphList
		args           args
		wantIsEmpty    bool
		wantMerged     bool
		wantRestNCount int
	}{
		{
			name:           "fixed paragraph > remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true)},
			args:           args{nCount: 1},
			wantIsEmpty:    false,
			wantMerged:     true,
			wantRestNCount: 0,
		},
		{
			name:           "fixed paragraph = remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true)},
			args:           args{nCount: 2},
			wantIsEmpty:    true,
			wantMerged:     true,
			wantRestNCount: 0,
		},
		{
			name:           "fixed paragraph < remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true)},
			args:           args{nCount: 3},
			wantIsEmpty:    true,
			wantMerged:     true,
			wantRestNCount: 1,
		},
		{
			name:           "un-fixed paragraph > remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true), newParagraphForTest(4, false)},
			args:           args{nCount: 1},
			wantIsEmpty:    false,
			wantMerged:     true,
			wantRestNCount: 0,
		},
		{
			name:           "un-fixed paragraph = remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true), newParagraphForTest(4, false)},
			args:           args{nCount: 2},
			wantIsEmpty:    false,
			wantMerged:     true,
			wantRestNCount: 0,
		},
		{
			name:           "un-fixed paragraph < remove count",
			plist:          &peEventParagraphList{newParagraphForTest(2, true), newParagraphForTest(3, true), newParagraphForTest(4, false)},
			args:           args{nCount: 3},
			wantIsEmpty:    false,
			wantMerged:     false,
			wantRestNCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMerged, gotRestNCount := tt.plist.MergeRemoveCount(tt.args.nCount)
			if gotMerged != tt.wantMerged {
				t.Errorf("peEventParagraphList.MergeRemoveCount() gotMerged = %v, want %v", gotMerged, tt.wantMerged)
			}
			if gotRestNCount != tt.wantRestNCount {
				t.Errorf("peEventParagraphList.MergeRemoveCount() gotRestNCount = %v, want %v", gotRestNCount, tt.wantRestNCount)
			}
			if isEmpty := tt.plist.IsEmpty(); isEmpty != tt.wantIsEmpty {
				t.Errorf("peEventParagraphList.MergeRemoveCount() gotIsEmpty = %v, want %v", isEmpty, tt.wantIsEmpty)
			}
		})
	}
}

func Test_peEventParagraphList_MergeInputRequest(t *testing.T) {
	type args struct {
		r uiadapter.InputRequestType
	}
	tests := []struct {
		name string
		p    peEventParagraphList
		args args
		want bool
	}{
		{
			name: "paragraph and input event",
			p:    peEventParagraphList{newParagraphForTest(3, true)},
			args: args{r: uiadapter.InputRequestCommand},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.MergeInputRequest(tt.args.r); got != tt.want {
				t.Errorf("peEventParagraphList.MergeInputRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peEventRemoveCount_MergeParagraph(t *testing.T) {
	type args struct {
		p *pubdata.Paragraph
	}
	tests := []struct {
		name       string
		p          peEventRemoveCount
		args       args
		wantMerged bool
	}{
		{
			name:       "remove count + paragraph",
			p:          peEventRemoveCount(3),
			args:       args{newParagraphForTest(5, true)},
			wantMerged: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMerged := tt.p.MergeParagraph(tt.args.p); gotMerged != tt.wantMerged {
				t.Errorf("peEventRemoveCount.MergeParagraph() = %v, want %v", gotMerged, tt.wantMerged)
			}
		})
	}
}

func Test_peEventRemoveCount_MergeRemoveCount(t *testing.T) {
	type args struct {
		nCount int
	}
	newEventRemoveCount := func(n int) *peEventRemoveCount {
		ev := peEventRemoveCount(n)
		return &ev
	}
	tests := []struct {
		name           string
		rmCount        *peEventRemoveCount
		args           args
		wantMerged     bool
		wantRestNCount int
		wantIsEmpty    bool
		wantMergedSelf int
	}{
		{
			name:           "remove count + remove count",
			rmCount:        newEventRemoveCount(5),
			args:           args{nCount: 3},
			wantMerged:     true,
			wantRestNCount: 0,
			wantIsEmpty:    false,
			wantMergedSelf: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMerged, gotRestNCount := tt.rmCount.MergeRemoveCount(tt.args.nCount)
			if gotMerged != tt.wantMerged {
				t.Errorf("peEventRemoveCount.MergeRemoveCount() gotMerged = %v, want %v", gotMerged, tt.wantMerged)
			}
			if gotRestNCount != tt.wantRestNCount {
				t.Errorf("peEventRemoveCount.MergeRemoveCount() gotRestNCount = %v, want %v", gotRestNCount, tt.wantRestNCount)
			}
			if isEmpty := tt.rmCount.IsEmpty(); isEmpty != tt.wantIsEmpty {
				t.Errorf("peEventRemoveCount.MergeRemoveCount() gotIsEmpty = %v, want %v", isEmpty, tt.wantIsEmpty)
			}
			if self := tt.rmCount.getRawData(); self != tt.wantMergedSelf {
				t.Errorf("peEventRemoveCount.MergeRemoveCount() gotMergedSelf = %v, want %v", self, tt.wantMergedSelf)
			}
		})
	}
}

func Test_peEventRemoveCount_MergeInputRequest(t *testing.T) {
	type args struct {
		r uiadapter.InputRequestType
	}
	tests := []struct {
		name string
		p    peEventRemoveCount
		args args
		want bool
	}{
		{
			name: "remove count + request",
			p:    peEventRemoveCount(7),
			args: args{r: uiadapter.InputRequestCommand},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.MergeInputRequest(tt.args.r); got != tt.want {
				t.Errorf("peEventRemoveCount.MergeInputRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peEventRemoveAll_MergeParagraph(t *testing.T) {
	type args struct {
		p *pubdata.Paragraph
	}
	tests := []struct {
		name       string
		p          peEventRemoveAll
		args       args
		wantMerged bool
	}{
		{
			name:       "remove all + paragraph",
			p:          peEventRemoveAll{},
			args:       args{newParagraphForTest(9, true)},
			wantMerged: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := peEventRemoveAll{}
			if gotMerged := p.MergeParagraph(tt.args.p); gotMerged != tt.wantMerged {
				t.Errorf("peEventRemoveAll.MergeParagraph() = %v, want %v", gotMerged, tt.wantMerged)
			}
		})
	}
}

func Test_peEventRemoveAll_MergeRemoveCount(t *testing.T) {
	type args struct {
		nCount int
	}
	tests := []struct {
		name           string
		p              peEventRemoveAll
		args           args
		wantMerged     bool
		wantRestNCount int
	}{
		{
			name:           "remove all + remove count",
			p:              peEventRemoveAll{},
			args:           args{3},
			wantMerged:     true,
			wantRestNCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := peEventRemoveAll{}
			gotMerged, gotRestNCount := p.MergeRemoveCount(tt.args.nCount)
			if gotMerged != tt.wantMerged {
				t.Errorf("peEventRemoveAll.MergeRemoveCount() gotMerged = %v, want %v", gotMerged, tt.wantMerged)
			}
			if gotRestNCount != tt.wantRestNCount {
				t.Errorf("peEventRemoveAll.MergeRemoveCount() gotRestNCount = %v, want %v", gotRestNCount, tt.wantRestNCount)
			}
		})
	}
}

func Test_peEventRemoveAll_MergeInputRequest(t *testing.T) {
	type args struct {
		r uiadapter.InputRequestType
	}
	tests := []struct {
		name string
		p    peEventRemoveAll
		args args
		want bool
	}{
		{
			name: "remove all + input request",
			p:    peEventRemoveAll{},
			args: args{r: uiadapter.InputRequestCommand},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := peEventRemoveAll{}
			if got := p.MergeInputRequest(tt.args.r); got != tt.want {
				t.Errorf("peEventRemoveAll.MergeInputRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peEventInputRequest_MergeParagraph(t *testing.T) {
	type args struct {
		p *pubdata.Paragraph
	}
	tests := []struct {
		name        string
		p           peEventInputRequest
		args        args
		wantMerged  bool
		wantIsEmpty bool
	}{
		{
			name:        "input request + paragraph",
			p:           peEventInputRequest(uiadapter.InputRequestCommand),
			args:        args{p: newParagraphForTest(4, true)},
			wantMerged:  false,
			wantIsEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMerged := tt.p.MergeParagraph(tt.args.p); gotMerged != tt.wantMerged {
				t.Errorf("peEventInputRequest.MergeParagraph() = %v, want %v", gotMerged, tt.wantMerged)
			}
			if isEmpty := tt.p.IsEmpty(); isEmpty != tt.wantIsEmpty {
				t.Errorf("peEventInputRequest.MergeParagraph(), IsEmpty %v, want %v", isEmpty, tt.wantIsEmpty)
			}
		})
	}
}

func Test_peEventInputRequest_MergeRemoveCount(t *testing.T) {
	type args struct {
		nCount int
	}
	tests := []struct {
		name           string
		p              peEventInputRequest
		args           args
		wantMerged     bool
		wantRestNCount int
		wantIsEmpty    bool
	}{
		{
			name:           "input request + remove count",
			p:              peEventInputRequest(uiadapter.InputRequestInput),
			args:           args{11},
			wantMerged:     false,
			wantRestNCount: 11,
			wantIsEmpty:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMerged, gotRestNCount := tt.p.MergeRemoveCount(tt.args.nCount)
			if gotMerged != tt.wantMerged {
				t.Errorf("peEventInputRequest.MergeRemoveCount() gotMerged = %v, want %v", gotMerged, tt.wantMerged)
			}
			if gotRestNCount != tt.wantRestNCount {
				t.Errorf("peEventInputRequest.MergeRemoveCount() gotRestNCount = %v, want %v", gotRestNCount, tt.wantRestNCount)
			}
			if isEmpty := tt.p.IsEmpty(); isEmpty != tt.wantIsEmpty {
				t.Errorf("peEventInputRequest.MergeRemoveCount(), IsEmpty %v, want %v", isEmpty, tt.wantIsEmpty)
			}
		})
	}
}

func Test_peEventInputRequest_MergeInputRequest(t *testing.T) {
	type args struct {
		r uiadapter.InputRequestType
	}
	newEventInputRequest := func(r uiadapter.InputRequestType) *peEventInputRequest {
		rr := peEventInputRequest(r)
		return &rr
	}
	tests := []struct {
		name        string
		ev          *peEventInputRequest
		args        args
		wantMerged  bool
		wantIsEmpty bool
	}{
		{
			name:        "command + command",
			ev:          newEventInputRequest(uiadapter.InputRequestCommand),
			args:        args{uiadapter.InputRequestCommand},
			wantMerged:  true,
			wantIsEmpty: false,
		},
		{
			name:        "command + input",
			ev:          newEventInputRequest(uiadapter.InputRequestCommand),
			args:        args{uiadapter.InputRequestInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "command + raw input",
			ev:          newEventInputRequest(uiadapter.InputRequestCommand),
			args:        args{uiadapter.InputRequestRawInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "command + none",
			ev:          newEventInputRequest(uiadapter.InputRequestCommand),
			args:        args{uiadapter.InputRequestNone},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "input + command",
			ev:          newEventInputRequest(uiadapter.InputRequestInput),
			args:        args{uiadapter.InputRequestCommand},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "input + input",
			ev:          newEventInputRequest(uiadapter.InputRequestInput),
			args:        args{uiadapter.InputRequestInput},
			wantMerged:  true,
			wantIsEmpty: false,
		},
		{
			name:        "input + raw input",
			ev:          newEventInputRequest(uiadapter.InputRequestInput),
			args:        args{uiadapter.InputRequestRawInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "input + none",
			ev:          newEventInputRequest(uiadapter.InputRequestInput),
			args:        args{uiadapter.InputRequestNone},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "raw input + command",
			ev:          newEventInputRequest(uiadapter.InputRequestRawInput),
			args:        args{uiadapter.InputRequestCommand},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "raw input + input",
			ev:          newEventInputRequest(uiadapter.InputRequestRawInput),
			args:        args{uiadapter.InputRequestInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "raw input + raw input",
			ev:          newEventInputRequest(uiadapter.InputRequestRawInput),
			args:        args{uiadapter.InputRequestRawInput},
			wantMerged:  true,
			wantIsEmpty: false,
		},
		{
			name:        "raw input + none",
			ev:          newEventInputRequest(uiadapter.InputRequestRawInput),
			args:        args{uiadapter.InputRequestNone},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "none + command",
			ev:          newEventInputRequest(uiadapter.InputRequestNone),
			args:        args{uiadapter.InputRequestCommand},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "none + input",
			ev:          newEventInputRequest(uiadapter.InputRequestNone),
			args:        args{uiadapter.InputRequestInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "none + raw input",
			ev:          newEventInputRequest(uiadapter.InputRequestNone),
			args:        args{uiadapter.InputRequestRawInput},
			wantMerged:  false,
			wantIsEmpty: false,
		},
		{
			name:        "none + none",
			ev:          newEventInputRequest(uiadapter.InputRequestNone),
			args:        args{uiadapter.InputRequestNone},
			wantMerged:  true,
			wantIsEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMerged := tt.ev.MergeInputRequest(tt.args.r); gotMerged != tt.wantMerged {
				t.Errorf("peEventInputRequest.MergeInputRequest() = %v, want %v", gotMerged, tt.wantMerged)
			}
			if isEmpty := tt.ev.IsEmpty(); isEmpty != tt.wantIsEmpty {
				t.Errorf("peEventInputRequest.MergeInputRequest(), IsEmpty %v, want %v", isEmpty, tt.wantIsEmpty)
			}

		})
	}
}
