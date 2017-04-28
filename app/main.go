package app

import (
	"fmt"
	"image"
	"io"
	"os"
	"runtime"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/math/f64"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"local/erago/util/log"
	customTheme "local/erago/view/exp/theme"
	"local/erago/view/exp/ui"
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
func setLogConfig(appConf *Config) (func(), error) {
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
	var writer io.WriteCloser
	switch logfile := appConf.LogFile; logfile {
	case LogFileStdOut, "":
		log.Info("Output log to Stdout")
		writer = os.Stdout
	case LogFileStdErr:
		log.Info("Output log to Stderr")
		writer = os.Stderr
	default:
		log.Info("Output log to ", logfile)
		fp, err := os.Create(logfile)
		if err != nil {
			return nil, err
		}
		writer = fp
	}
	log.SetOutput(writer)

	return func() {
		writer.Close()
	}, nil
}

// entry point of main application. appconf nil is OK,
// use default if it is.
// its internal errors are handled by itself.
func Main(appConf *Config) {
	if appConf == nil {
		appConf = NewConfig(DefaultBaseDir)
	}

	// returned value must be called once.
	reset, err := setLogConfig(appConf)
	if err != nil {
		log.Infoln("Error: Can't create log file:", err)
		return
	}
	defer reset()

	// construct theme.
	t, err := BuildTheme(appConf)
	if err != nil {
		log.Info("Error: BuildTheme FAIL: ", err)
		return
	}

	appMain(t, appConf)
}

func appMain(t *theme.Theme, appConf *Config) {
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
		if err := runWindow(s, t, appConf); err != nil {
			log.Info("Error: app.runWindow(): ", err)
		} else {
			log.Info("...quiting correctly")
		}
	})
}

// references golang.org/x/exp/shiny/widget/widget.go
func runWindow(s screen.Screen, t *theme.Theme, appConf *Config) error {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Width:  appConf.Width,
		Height: appConf.Height,
	})
	if err != nil {
		return fmt.Errorf("NewWindow FAIL: %v", err)
	}
	defer w.Release()

	presenter := ui.NewEragoPresenter(w)
	root := NewUI(presenter)

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
			if e.To == lifecycle.StageDead {
				return mef.Err()
			}
			if e.Crosses(lifecycle.StageVisible) == lifecycle.CrossOn {
				log.Debug("RunGameThread() ... ")
				if startOK := presenter.RunGameThread(root.Editor(), appConf.Game); startOK {
					log.Debug("RunGameThread() ... start OK")
					defer presenter.Quit()
				}
			}

		case gesture.Event, key.Event, mouse.Event:
			root.OnInputEvent(e, image.Point{})

		case PaintScheduled:
			// paint event is comming after some times
			// unmark paint request now.
			root.Marks.UnmarkNeedsPaint()

		case paint.Event:
			if err := paintRoot(s, w, Theme, root); err != nil {
				return err
			}
			w.Publish()
			paintPending = false

		case size.Event:
			if dpi := float64(e.PixelsPerPt) * unit.PointsPerInch; dpi != Theme.GetDPI() {
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

// TODO: This way is OK for quitting application manually?
var stageDeadEvent = lifecycle.Event{To: lifecycle.StageDead, From: lifecycle.StageFocused}

func (me *ModelErrorFilter) Filter(e interface{}) interface{} {
	if me.err == nil {
		// catch model error and store it. otherwise pass through.
		switch e := e.(type) {
		case ui.ModelError:
			log.Debug("catch ModelError")
			if e.HasCause() {
				me.err = e
				return nil
			}
			return stageDeadEvent

		default:
			return e
		}
	}

	// modelErr is catched, quit app immediately when specified events are arrived.
	switch e := e.(type) {
	case gesture.Event:
		return stageDeadEvent

	case key.Event:
		if e.Direction == key.DirPress {
			return stageDeadEvent
		}

	case mouse.Event:
		if e.Direction == mouse.DirPress {
			return stageDeadEvent
		}
	}
	return e
}
