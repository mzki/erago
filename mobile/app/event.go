package mobile

import (
	"time"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/geom"
)

var appRunning bool = false

// OnCreate creates erago application with config.
// nil config will occurs panic.
// the configuration is reflected at call OnStart().
// after OnStart(), calling OnCreate will be panic.
func PreConfigure(config *Config) {
	if appRunning {
		panic("mobile.OnCreate(): app is already running.")
	}
	if config == nil {
		panic("nil config")
	}
	configInstance = config
}

// OnStart() starts erago application.
// after this, any OnXXX() is valid.
// caller must call OnStop() at end to quit the
// application correctly.
func OnStart(mobileDir string, ctx AppContext) {
	if appRunning {
		panic("mobile.OnStart: already start")
	}
	appRunning = true

	appInstance = newTheApp()
	appInstance.context = ctx

	go appInstance.main(mobileDir, configInstance)

	appInstance.send(lifecycle.Event{
		From: lifecycle.StageDead,
		To:   lifecycle.StageVisible,
	})
}

// OnResume() notify to application
// screen is shown.
func OnResume() {
	if !appRunning {
		panic("mobile.OnResume(): no running app")
	}
	appInstance.send(lifecycle.Event{
		From: lifecycle.StageVisible,
		To:   lifecycle.StageFocused,
	})
}

// OnPause() notify to application
// screen is not shown.
func OnPause() {
	if !appRunning {
		panic("mobile.OnPause(): no running app")
	}
	appInstance.send(lifecycle.Event{
		From: lifecycle.StageFocused,
		To:   lifecycle.StageVisible,
	})
}

// OnStop() stops the application.
// this must be called at end.
// after that all OnXXX() is invalid.
func OnStop() {
	if appRunning {
		appRunning = false
		appInstance.send(lifecycle.Event{
			From: lifecycle.StageVisible,
			To:   lifecycle.StageDead,
		})
	}
}

// OnMeasure() notify the screen dimention.
func OnMeasure(width, height int, dpi float64) {
	if !appRunning {
		panic("mobile.OnMeasure(): no running app")
	}
	pxPerPt := float32(dpi / unit.PointsPerInch)
	appInstance.send(size.Event{
		Orientation: size.OrientationPortrait, // must be
		WidthPx:     width,
		HeightPx:    height,
		PixelsPerPt: pxPerPt,
		WidthPt:     geom.Pt(float32(width) / pxPerPt),
		HeightPt:    geom.Pt(float32(height) / pxPerPt),
	})
}

func OnSingleTapped(x, y int) {
	if !appRunning {
		panic("mobile.OnSingleTapped(): no running app")
	}
	gp := gesture.Point{float32(x), float32(y)}
	appInstance.send(gesture.Event{
		Type:       gesture.TypeTap,
		InitialPos: gp,
		CurrentPos: gp,
		Time:       time.Now(),
	})
}

func OnDoubleTapped(x, y int) {
	if !appRunning {
		panic("mobile.OnDoubleTapped(): no running app")
	}
	gp := gesture.Point{float32(x), float32(y)}
	appInstance.send(gesture.Event{
		Type:        gesture.TypeIsDoublePress,
		DoublePress: true,
		InitialPos:  gp,
		CurrentPos:  gp,
		Time:        time.Now(),
	})
}

func OnCommandSelected(cmd string) {
	if !appRunning {
		panic("mobile.OnCommandSelected(): no running app")
	}
	appInstance.send(externalCmdEvent{
		Command: cmd,
	})
}

func OnScroll(dx, dy int) {
	if !appRunning {
		panic("mobile.OnScroll(): no running app")
	}
	// TODO
}

func OnScrollLines(nlines int) {
	if !appRunning {
		panic("mobile.OnScrollLines(): no running app")
	}
	if nlines == 0 {
		return
	}

	// in mobile, negative value means viewport is down and,
	// positive value means up.
	var wheel = mouse.ButtonWheelUp
	if nlines < 0 {
		nlines = -nlines
		wheel = mouse.ButtonWheelDown
	}

	for i := 0; i < nlines; i++ {
		appInstance.send(mouse.Event{Button: wheel})
	}
}

func LockBuffer() []byte {
	return appInstance.LockBuffer()
}

func UnlockBuffer(b []byte) {
	appInstance.UnlockBuffer(b)
}
