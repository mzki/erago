package app

import (
	"time"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/paint"
)

// fpsLimitter limits and delays paint events to not exceed over specified fps.
// To construct this:
//   fps := FpsLimitter{EventDeque: deque, FPS: fps}
// zero value is invalid.
type FpsLimitter struct {
	EventDeque screen.EventDeque

	FPS       int // zero FPS means no limit fps.
	lastPaint time.Time
	scheduled bool // have a schedule for sending paint event?
}

// paintSchedule notifies setting a new schedule which sends paint event after some time.
type PaintScheduled struct{}

type delayedPaintEvent struct{}

// Filter handles paint event to limits fps.
// if paint event is delayed, return type paintSchedule and
// consume paint event.
// other events are returned itself.
func (l *FpsLimitter) Filter(ev interface{}) interface{} {
	switch ev := ev.(type) {
	case delayedPaintEvent:
		l.scheduled = false
		l.lastPaint = time.Now()
		return paint.Event{}

	case paint.Event:
		if l.FPS <= 0 {
			return ev
		}
		if l.scheduled {
			// consume paint event since it is sent after some times.
			return nil
		}

		frame := time.Second / time.Duration(l.FPS)
		now := time.Now()
		if d := now.Sub(l.lastPaint); d > 0 && d < frame {
			// too early coming paint event, delays it.
			_ = time.AfterFunc(frame-d, func() {
				l.EventDeque.SendFirst(delayedPaintEvent{})
			})
			l.scheduled = true
			return PaintScheduled{}
		}
		// too slow coming paint event, return immediatly
		l.lastPaint = now
		return ev
	}
	return ev
}
