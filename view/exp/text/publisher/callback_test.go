package publisher

import (
	"errors"
	"testing"

	"github.com/mzki/erago/view/exp/text/pubdata"
)

var errTestCallbackFuncErrorValue = errors.New("callback func should return this")

func TestCallbackDefault_CallByUserCustom(t *testing.T) {
	type args struct {
		p *pubdata.Paragraph
		n int
	}
	tests := []struct {
		name     string
		cb       *CallbackDefault
		args     args
		testFunc func(cb *CallbackDefault, args args) error
		wantErr  bool
	}{
		{
			name: "OnPublish",
			cb: &CallbackDefault{
				OnPublishFunc: func(*pubdata.Paragraph) error { return errTestCallbackFuncErrorValue },
			},
			args:     args{p: nil},
			testFunc: func(cb *CallbackDefault, args args) error { return cb.OnPublish(args.p) },
			wantErr:  true,
		},
		{
			name: "OnPublishTemporary",
			cb: &CallbackDefault{
				OnPublishTemporaryFunc: func(*pubdata.Paragraph) error { return errTestCallbackFuncErrorValue },
			},
			args:     args{p: nil},
			testFunc: func(cb *CallbackDefault, args args) error { return cb.OnPublishTemporary(args.p) },
			wantErr:  true,
		},
		{
			name: "OnRemove",
			cb: &CallbackDefault{
				OnRemoveFunc: func(int) error { return errTestCallbackFuncErrorValue },
			},
			args:     args{n: 10},
			testFunc: func(cb *CallbackDefault, args args) error { return cb.OnRemove(args.n) },
			wantErr:  true,
		},
		{
			name: "OnRemoveAll",
			cb: &CallbackDefault{
				OnRemoveAllFunc: func() error { return errTestCallbackFuncErrorValue },
			},
			args:     args{},
			testFunc: func(cb *CallbackDefault, args args) error { return cb.OnRemoveAll() },
			wantErr:  true,
		},
		{
			name: "OnSync",
			cb: &CallbackDefault{
				OnSyncFunc: func() error { return errTestCallbackFuncErrorValue },
			},
			args:     args{},
			testFunc: func(cb *CallbackDefault, args args) error { return cb.OnSync() },
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.testFunc(tt.cb, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("CallbackDefault.%s() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
