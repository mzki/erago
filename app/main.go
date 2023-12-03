package app

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"runtime"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/util/log"
	customTheme "github.com/mzki/erago/view/exp/theme"
	"github.com/mzki/erago/view/exp/ui"
)

// build new theme according to application configure.
func BuildTheme(appConf *Config) (*theme.Theme, error) {
	Theme := customTheme.Default
	fontFile := appConf.Font
	if fontFile == "default" || fontFile == "" {
		fontFile = customTheme.DefaultFontName
	}
	fontOpt := &customTheme.FontFaceOptions{
		Size: appConf.FontSize,
	}

	catalog, err := customTheme.NewOneFontFaceCatalog(fontFile, fontOpt)
	if err != nil {
		return nil, err
	}
	Theme.FontFaceCatalog = catalog
	return &Theme, nil
}

// set log configuration and return finalize function with internal error.
// when returned error, the finalize function is nil and need not be called.
func SetLogConfig(appConf *Config) (func(), error) {
	// set log level.
	switch level := appConf.LogLevel; level {
	case LogLevelInfo:
		log.SetLevel(log.InfoLevel)
	case LogLevelDebug:
		log.SetLevel(log.DebugLevel)
	default:
		log.Infof("unknown log level(%s). use 'info' level insteadly.", level)
		log.SetLevel(log.InfoLevel)
	}

	// set log distination
	var (
		dstString string
		writer    io.WriteCloser
		closeFunc func()
	)
	switch logfile := appConf.LogFile; logfile {
	case LogFileStdOut, "":
		dstString = "Stdout"
		writer = os.Stdout
		closeFunc = func() {}
	case LogFileStdErr:
		dstString = "Stdout"
		writer = os.Stderr
		closeFunc = func() {}
	default:
		dstString = logfile
		fp, err := filesystem.Store(logfile)
		if err != nil {
			return nil, err
		}
		writer = fp
		closeFunc = func() { fp.Close() }
	}
	logLimit := appConf.LogLimitMegaByte * 1000 * 1000
	if logLimit < 0 {
		logLimit = 0
	}
	log.SetOutput(log.LimitWriter(writer, logLimit))
	if err := testingLogOutput("log output sanity check..."); err != nil {
		closeFunc()
		return nil, err
	}
	log.Infof("Output log to %s", dstString)

	return closeFunc, nil
}

func testingLogOutput(msg string) error {
	log.Debug(msg)
	err := log.Err()
	switch {
	case errors.Is(err, log.ErrOutputDiscardedByLevel):
	case errors.Is(err, io.EOF):
	case err == nil:
	default:
		return fmt.Errorf("log output error: %w", err)
	}
	return nil // normal operation
}

// entry point of main application. appconf nil is OK,
// use default if it is.
// its internal errors are handled by itself.
func Main(title string, appConf *Config) {
	if appConf == nil {
		appConf = NewConfig(DefaultBaseDir)
	}

	// returned value must be called once.
	reset, err := SetLogConfig(appConf)
	if err != nil {
		// TODO: what is better way to handle fatal error in this case?
		fmt.Fprintf(os.Stderr, "log configuration failed: %v\n", err)
		return
	}
	defer reset()

	// construct theme.
	t, err := BuildTheme(appConf)
	if err != nil {
		log.Info("Error: BuildTheme FAIL: ", err)
		return
	}

	log.Infof("-- %s --\n", title)

	// main loop
	driver.Main(func(s screen.Screen) {
		// capture panic as error in this thread
		defer func() {
			if rec := recover(); rec != nil {
				buf := make([]byte, 4096)
				buf_end := runtime.Stack(buf, false)
				log.Info("PANIC: ", fmt.Errorf("%v\n%v\n", rec, string(buf[:buf_end])))
			}
		}()

		// run UI handler.
		if err := runWindow(title, s, t, appConf); err != nil {
			log.Info("Error: app.runWindow(): ", err)
		} else {
			log.Info("...quiting correctly")
		}
	})
}

// references golang.org/x/exp/shiny/widget/widget.go
func runWindow(title string, s screen.Screen, t *theme.Theme, appConf *Config) error {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Width:  appConf.Width,
		Height: appConf.Height,
		Title:  title,
	})
	if err != nil {
		return fmt.Errorf("NewWindow FAIL: %v", err)
	}
	defer w.Release()

	presenter := ui.NewEragoPresenter(w)
	root := NewUI(presenter, appConf)

	Theme := t

	var (
		paintPending bool = false
	)

	gef := gesture.EventFilter{EventDeque: w}
	fps := FpsLimitter{EventDeque: w, FPS: 60}
	mef := ModelErrorFilter{}
	for {
		e := w.NextEvent()

		if e = gef.Filter(e); e == nil {
			continue
		}
		if e = fps.Filter(e); e == nil {
			continue
		}
		if e = mef.Filter(e); e == nil {
			continue
		}

		switch e := e.(type) {
		case lifecycle.Event:
			root.OnLifecycleEvent(e)

			// game thread control
			switch {
			case e == stageDeadEvent:
				// Model Error
				return mef.Err()
			case e == stageRestartEvent:
				if err := presenter.RestartGameThread(root.Editor(), appConf.Game); err == nil {
					log.Debug("RestartGameThread() ... OK")
					mef.Reset() // resets model error state too since game thread does.
				} else {
					log.Debug("RestartGameThread() ... NG")
					return fmt.Errorf("Failed to restart game: %w", err)
				}
			case e.Crosses(lifecycle.StageVisible) == lifecycle.CrossOn:
				log.Debug("RunGameThread() ... ")
				if err := presenter.RunGameThread(root.Editor(), appConf.Game); err == nil {
					log.Debug("RunGameThread() ... OK")
					defer presenter.Quit()
				} else {
					log.Debug("RunGameThread() ... NG")
					return fmt.Errorf("Failed to start game: %w", err)
				}
			}

		case gesture.Event, key.Event, mouse.Event:
			root.OnInputEvent(e, image.Point{})

			if e, ok := e.(key.Event); ok {
				if e.Code == key.CodeF5 && e.Direction == key.DirPress {
					// restart command
					w.Send(stageRestartEvent)
				}
			}

		case PaintScheduled:
			// paint event is comming after some times
			// unmark paint request now.
			root.Marks.UnmarkNeedsPaint()

		case paint.Event:
			if err := paintRoot(s, w, Theme, root); err != nil {
				return fmt.Errorf("paint failed: %w", err)
			}
			w.Publish()
			paintPending = false

		case size.Event:
			if dpi := DPI(e.PixelsPerPt); dpi != Theme.GetDPI() {
				newT := new(theme.Theme)
				if Theme != nil {
					*newT = *Theme
				}
				newT.DPI = dpi
				catalog := newT.FontFaceCatalog.(*customTheme.OneFontFaceCatalog)
				catalog.UpdateFontFaceOptions(&customTheme.FontFaceOptions{DPI: dpi})
				Theme = newT
			}

			size := e.Size()
			root.Measure(Theme, size.X, size.Y)
			root.Wrappee().Rect = e.Bounds()
			root.Layout(Theme)

		case ui.PresenterTask:
			e.Run()

		case error:
			log.Debug("FATAL: UI's Fatal Error")
			return e
		}

		if !paintPending && root.Wrappee().Marks.NeedsPaint() {
			paintPending = true
			w.Send(paint.Event{})
		}
	}
}

func paintRoot(s screen.Screen, w screen.Window, t *theme.Theme, root node.Node) error {
	ctx := &node.PaintContext{
		Screen: s,
		Drawer: w,
		Theme:  t,
		Src2Dst: f64.Aff3{
			1, 0, 0,
			0, 1, 0,
		},
	}
	return root.Paint(ctx, image.ZP)
}

// ModelErrorFilter filters ui.EragoPresenter's ModelError.
type ModelErrorFilter struct {
	err error
}

// return internal error. it is valid when model error is arrived.
func (me ModelErrorFilter) Err() error {
	return me.err
}

var (
	// TODO: This way is OK for quitting application manually?
	stageDeadEvent = lifecycle.Event{To: lifecycle.StageDead, From: lifecycle.StageFocused}

	stageRestartEvent = lifecycle.Event{To: lifecycle.StageVisible, From: lifecycle.StageVisible}
)

func (me *ModelErrorFilter) Filter(e interface{}) interface{} {
	if me.err == nil {
		// catch model error and store it. otherwise pass through.
		switch e := e.(type) {
		case ui.ModelError:
			log.Debug("catch ModelError")
			switch err := e.Cause(); err {
			case nil:
				// game quiting correctly. end application immediately.
				return stageDeadEvent
			case ui.ErrorGameQuitByRestartRequest:
				// caused by restart request, ignore this error.
				return nil
			default:
				// game thread error. end application by next event.
				me.err = e
				return nil
			}

		default:
			return e
		}
	}

	// modelErr is catched, quit app immediately when specified events are arrived.
	switch e := e.(type) {
	case key.Event:
		if e.Code == key.CodeReturnEnter && e.Direction == key.DirPress {
			return stageDeadEvent
		}
	}
	return e
}

// Reset resets internal error state.
func (me *ModelErrorFilter) Reset() {
	me.err = nil
}
