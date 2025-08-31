package app

import (
	"fmt"
	"image"
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

	"github.com/mzki/erago/app/config"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/util/log"
	customTheme "github.com/mzki/erago/view/exp/theme"
	"github.com/mzki/erago/view/exp/ui"
)

// build new theme according to application configure.
func BuildTheme(appConf *config.Config) (*theme.Theme, error) {
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

// entry point of main application. appconf nil is OK,
// use default if it is.
// its internal errors are handled by itself.
func Main(title string, appConf *config.Config) {
	if appConf == nil {
		appConf = config.NewConfig(config.DefaultBaseDir)
	}

	// returned value must be called once.
	reset, err := config.SetupLogConfig(appConf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log configuration failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: try to change config logfile from %v\n", appConf.LogFile)
		// sometimes stdout/err are also not avialble, e.g. windows gui application.
		// try to create "fatal file" to record error.
		w, fsyserr := filesystem.Store("fatal")
		if fsyserr != nil {
			// TODO: this case, we do not have any ways to notify the error to users....
			fmt.Fprintf(os.Stderr, "failed to create fatal file by %v: previous fail %v\n", fsyserr, err)
			return
		}
		defer w.Close()
		fmt.Fprintf(w, "log configuration failed: %v\n", err)
		fmt.Fprintf(w, "Hint: try to change config logfile from %v\n", appConf.LogFile)
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
func runWindow(title string, s screen.Screen, t *theme.Theme, appConf *config.Config) error {
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
