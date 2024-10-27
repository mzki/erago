package script

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/mzki/erago/util/log"
	lua "github.com/yuin/gopher-lua"
)

type watchDogTimerNotification = chan struct{}
type watchDogTimerNotificationRx = <-chan struct{}

// watchDogTimer is a watcher for the timer expiring with
// feature resetting count down to keep alive.
// It is not support concurrency, need to call APIs on a single thread.
type watchDogTimer struct {
	timer           *time.Timer
	timeoutDuration time.Duration

	timerExpired   watchDogTimerNotification
	commandChannel chan wdtCommand
	running        atomic.Bool
	quit           chan struct{}
}

var wdtDefaultTimeout = InfiniteLoopTimeoutSecond

func newWatchDogTimer(d time.Duration) *watchDogTimer {
	if d <= 0 {
		d = wdtDefaultTimeout
	}
	return &watchDogTimer{
		// timer is initialzed later to not start at initialization phase.
		timeoutDuration: d,
		timerExpired:    make(watchDogTimerNotification),
		commandChannel:  make(chan wdtCommand),
		running:         atomic.Bool{},
		quit:            make(chan struct{}),
	}
}

func (wdt *watchDogTimer) initTimer() {
	if wdt.timer == nil {
		wdt.timer = time.NewTimer(wdt.timeoutDuration)
	}
}

type wdtCommand int

const (
	wdtRequestResetTimer wdtCommand = iota + 1
	wdtRequestStopTimer
	wdtRequestQuitTimer
)

// Stop stops internal timer count down. It returns true if timer stopped,
// returns false if timer is not stopped due to internal timer is not running yet.
func (wdt *watchDogTimer) Stop() bool {
	log.Debugln("WatchDogTimer: Stop")
	if !wdt.running.Load() {
		return false
	}
	wdt.commandChannel <- wdtRequestStopTimer
	return true
}

// Reset resets internal timer count down to start from entire time limit. It returns
// true if timer reset, returns false if timer is not reset due to internal timer
// is not running yet.
func (wdt *watchDogTimer) Reset() bool {
	log.Debugln("WatchDogTimer: Reset")
	if !wdt.running.Load() {
		return false
	}
	wdt.commandChannel <- wdtRequestResetTimer
	return true
}

// Quit quits internal timer running. It returns true if internal timer is quited,
// returns false if internal timer is not quited due to internal timer is not running yet.
// Once returning true, calling Quit() may cause panic. User need to re-create watchDogTimer
// newly after calling Run() and Quit().
func (wdt *watchDogTimer) Quit() bool {
	if !wdt.running.Load() {
		return false
	}
	select {
	case wdt.commandChannel <- wdtRequestQuitTimer:
		// send quit command correctly
	case <-time.After(5 * time.Second):
		log.Infoln("WatchDogTimer: Never Quit. please report to developer")
		panic("WatchDogTimer.Quit never quits even 5 second elapsed.")
	case <-wdt.quit:
		//quit correctly
	}
	return true
}

func (wdt *watchDogTimer) reset() {
	// Need to ensure timer stopped before Reset()
	wdt.timer.Reset(wdt.timeoutDuration)
}

func (wdt *watchDogTimer) stop() {
	if !wdt.timer.Stop() {
		select {
		case <-wdt.timer.C:
		default:
		}
	}
}

func (wdt *watchDogTimer) drainCommand() {
	for {
		select {
		case <-wdt.commandChannel:
		default:
			return
		}
	}
}

// Expired returns timer expired notification channel.
// Once notification channel receives timer expired, the channel closed.
func (wdt *watchDogTimer) Expired() watchDogTimerNotificationRx {
	return wdt.timerExpired
}

func (wdt *watchDogTimer) IsExpired() bool {
	select {
	case <-wdt.timerExpired:
		// treat same as whether timerExpired closed or not
		return true
	default:
		return false
	}
}

func (wdt *watchDogTimer) IsRunning() bool { return wdt.running.Load() }

func (wdt *watchDogTimer) IsQuit() bool {
	select {
	case <-wdt.quit:
		return true
	default:
		return false
	}
}

// Run runs watch dog timer and notify timer expired via Expired().
// The returned value indicates running timer is succeed or not, true for first call
// of Run(), false for timer is already running or WatchDogTimer is already quit.
func (wdt *watchDogTimer) Run(ctx context.Context) bool {
	if wdt.running.Load() {
		return false
	}
	if wdt.IsExpired() {
		return false
	}
	if wdt.IsQuit() {
		return false
	}

	wdt.initTimer()
	wdt.running.Store(true)
	go func() {
		defer close(wdt.quit)
		defer wdt.drainCommand()
		defer wdt.running.Store(false)
	loop:
		for {
			select {
			case cmd := <-wdt.commandChannel:
				switch cmd {
				case wdtRequestResetTimer:
					wdt.stop() // to ensure timer stopped before reset.
					wdt.reset()
				case wdtRequestStopTimer:
					wdt.stop()
				case wdtRequestQuitTimer:
					break loop
				}
			case <-wdt.timer.C:
				wdt.timerExpired <- struct{}{}
				close(wdt.timerExpired)
				break loop
			case <-ctx.Done():
				break loop
			}
		}
	}()
	return true
}

func (wdt *watchDogTimer) KeepAliveFuncCall(L *lua.LState, fn lua.LGFunction) int {
	wdt.Stop()
	defer wdt.Reset() // need defer since fn may occur panic even if on proper flow.
	ret := fn(L)
	return ret
}

func (wdt *watchDogTimer) WrapKeepAliveLG(fn lua.LGFunction) lua.LGFunction {
	return lua.LGFunction(func(L *lua.LState) int {
		return wdt.KeepAliveFuncCall(L, fn)
	})
}
