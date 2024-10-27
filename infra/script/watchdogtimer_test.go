package script

import (
	"context"
	"reflect"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func Test_newWatchDogTimer(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"normal case", args{5 * time.Second}, 5 * time.Second},
		{"zero duration", args{0}, wdtDefaultTimeout},
		{"negative duration", args{-100}, wdtDefaultTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newWatchDogTimer(tt.args.d); got.timeoutDuration != tt.want {
				t.Errorf("newWatchDogTimer(), timer duration = %v, want %v", got.timeoutDuration, tt.want)
			}
		})
	}
}

func newDefaultWatchDogTimer() *watchDogTimer {
	return newWatchDogTimer(wdtDefaultTimeout)
}

func Test_watchDogTimer_Stop(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		newwdt func() *watchDogTimer
		want   bool
	}{
		{"normal case, stop after run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			return wdt
		}, true},
		{"stop before Run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			return wdt
		}, false},
		{"stop after Quit", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			wdt.Quit()
			return wdt
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt()
			defer wdt.Quit()
			if got := wdt.Stop(); got != tt.want {
				t.Errorf("watchDogTimer.Stop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_watchDogTimer_Reset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		newwdt func() *watchDogTimer
		want   bool
	}{
		{"normal case, Reset after run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			return wdt
		}, true},
		{"Reset before Run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			return wdt
		}, false},
		{"Reset after Quit", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			wdt.Quit()
			return wdt
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt()
			defer wdt.Quit()
			if got := wdt.Reset(); got != tt.want {
				t.Errorf("watchDogTimer.Reset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_watchDogTimer_Quit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		newwdt func() *watchDogTimer
		want   bool
	}{
		{"normal case, Quit after run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			return wdt
		}, true},
		{"Quit before Run", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			return wdt
		}, false},
		{"Quit after Quit", func() *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(context.Background())
			wdt.Quit()
			return wdt
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt()
			defer wdt.Quit()
			if got := wdt.Quit(); got != tt.want {
				t.Errorf("watchDogTimer.Quit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_watchDogTimer_Expired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		newwdt        func() *watchDogTimer
		shouldExpired bool
	}{
		{"normal case, Expired after Timeout Elapsed", func() *watchDogTimer {
			const timeout = 100 * time.Millisecond
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			time.Sleep(timeout * 3)
			return wdt
		}, true},
		{"Expired before Timeout Elapsed", func() *watchDogTimer {
			const timeout = 10 * time.Second
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			return wdt
		}, false},
		{"Expired after Quit and before Timeout Elapsed", func() *watchDogTimer {
			const timeout = 10 * time.Second
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			wdt.Quit()
			return wdt
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt()
			defer wdt.Quit()
			var gotExpired bool
			select {
			case <-wdt.Expired():
				gotExpired = true
			default:
				gotExpired = false
			}
			if got := gotExpired; !reflect.DeepEqual(got, tt.shouldExpired) {
				t.Errorf("watchDogTimer.Expired() = %v, want %v", got, tt.shouldExpired)
			}
		})
	}
}

func Test_watchDogTimer_IsExpired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		newwdt func() *watchDogTimer
		want   bool
	}{
		{"normal case, Expired after Timeout Elapsed", func() *watchDogTimer {
			const timeout = 100 * time.Millisecond
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			time.Sleep(timeout * 3)
			return wdt
		}, true},
		{"Expired before Timeout Elapsed", func() *watchDogTimer {
			const timeout = 10 * time.Second
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			return wdt
		}, false},
		{"Expired after Quit and before Timeout Elapsed", func() *watchDogTimer {
			const timeout = 10 * time.Second
			wdt := newWatchDogTimer(timeout)
			wdt.Run(context.Background())
			wdt.Quit()
			return wdt
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt()
			defer wdt.Quit()
			if got := wdt.IsExpired(); got != tt.want {
				t.Errorf("watchDogTimer.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_watchDogTimer_Run(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
	tests := []struct {
		name    string
		newwdt  func(args args) *watchDogTimer
		newargs func() args
		want    bool
	}{
		{"normal case, Run after new", func(args args) *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			return wdt
		}, func() args {
			return args{context.Background(), func() {}}
		}, true},
		{"Run after Quit", func(args args) *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(args.ctx)
			wdt.Quit()
			// to ensure quit completely.
			timeout := time.After(1 * time.Second)
			ticker := time.Tick(100 * time.Millisecond)
		running:
			for !wdt.IsQuit() {
				select {
				case <-timeout:
					// This case indicates quit not ends, detect bug at comparision of got and want.
					break running
				case <-ticker:
					// loop next time
				}
			}
			return wdt
		}, func() args {
			return args{context.Background(), func() {}}
		}, false},
		{"Run canceled by context", func(args args) *watchDogTimer {
			wdt := newDefaultWatchDogTimer()
			wdt.Run(args.ctx)
			args.cancel()
			// to ensure cancel completely.
			timeout := time.After(1 * time.Second)
			ticker := time.Tick(100 * time.Millisecond)
		running:
			for !wdt.IsQuit() {
				select {
				case <-timeout:
					// This case indicates quit not ends, detect bug at comparision of got and want.
					break running
				case <-ticker:
					// loop next time
				}
			}
			return wdt
		}, func() args {
			ctx, cancel := context.WithCancel(context.Background())
			return args{ctx, cancel}
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			args := tt.newargs()
			defer args.cancel()
			wdt := tt.newwdt(args)
			defer wdt.Quit()
			if got := wdt.Run(args.ctx); got != tt.want {
				t.Errorf("watchDogTimer.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_watchDogTimer_KeepAliveFuncCall(t *testing.T) {
	t.Parallel()
	type args struct {
		L              *lua.LState
		fn             lua.LGFunction
		fnCallDuration time.Duration
		fnCallTimes    int
	}
	tests := []struct {
		name          string
		newwdt        func(args args) *watchDogTimer
		args          args
		shouldExpired bool
	}{
		{"normal case, Not expired watch dog timer", func(args args) *watchDogTimer {
			// setup duration to LState to use later on LGFunction.
			args.L.SetGlobal("fnCallDuration", lua.LNumber(args.fnCallDuration))

			wdt := newWatchDogTimer(args.fnCallDuration * time.Duration(args.fnCallTimes))
			wdt.Run(context.Background())
			return wdt
		}, args{lua.NewState(), func(L *lua.LState) int {
			fnCallDurationLN := lua.LVAsNumber(L.GetGlobal("fnCallDuration"))
			fnCallDuration := time.Duration(fnCallDurationLN)
			time.Sleep(fnCallDuration)
			return 0
		}, 100 * time.Millisecond, 10}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt // store local scope for parallel use
			t.Parallel()
			wdt := tt.newwdt(tt.args)
			defer wdt.Quit()
			fnCallTimesPlus1 := tt.args.fnCallTimes + 1 // +1 to exceed expired time
			for i := 0; i < fnCallTimesPlus1; i++ {
				wdt.KeepAliveFuncCall(tt.args.L, tt.args.fn)
			}
			if got := wdt.IsExpired(); got != tt.shouldExpired {
				t.Errorf("watchDogTimer.KeepAliveFuncCall(); Expired = %v, want %v", got, tt.shouldExpired)
			}
		})
	}
}

func Test_watchDogTimer_WrapKeepAliveLG(t *testing.T) {
	t.Skip("This case is just wrapper for KeepAlive Func call, not need to test")
	type args struct {
		fn lua.LGFunction
	}
	tests := []struct {
		name string
		wdt  *watchDogTimer
		args args
		want lua.LGFunction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wdt.WrapKeepAliveLG(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("watchDogTimer.WrapKeepAliveLG() = %v, want %v", got, tt.want)
			}
		})
	}
}
