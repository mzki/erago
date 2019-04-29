package mobile

import (
	"fmt"
	"image"
	"runtime"
	"sync"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"github.com/mzki/erago/app"
	"github.com/mzki/erago/util/deque"
	"github.com/mzki/erago/util/log"
	customTheme "github.com/mzki/erago/view/exp/theme"
	"github.com/mzki/erago/view/exp/ui"
)

// AppContext is interface for
// handling mobile.App's event.
// Any functions of this are called asynchronously.
type AppContext interface {
	// it is called when mobile.app is quited.
	// the native framework must be quited by this signal.
	// the argument erorr indicates why app is quited.
	// nil error means quit correctly.
	NotifyQuit(error)

	// it is called when mobile.app readies paint buffer.
	// the native framework can draw its paint buffer by
	// LockBuffer() and UnlockBuffer() pattern.
	NotifyPaint()

	// it is called when mobile.app requires inputting
	// user's command. the native framework should lists
	// command candidates.
	NotifyCommandRequest(*CmdSlice)

	// it is called when mobile.app no longer require
	// inputting user's command.
	// the native framework should stop sending command
	// to mobile.app.
	NotifyCommandRequestClose()
}

var appInstance = newTheApp()

type theApp struct {
	eventQ deque.EventDeque

	context AppContext

	size image.Point

	bufMu *sync.Mutex
	buf   *image.RGBA
}

func newTheApp() *theApp {
	return &theApp{
		eventQ: deque.NewEventDeque(),
		bufMu:  new(sync.Mutex),
	}
}

func (a *theApp) send(e interface{}) {
	a.eventQ.Send(e)
}

// get underlying buffer with mutex lock.
// as soon as call UnlockBuffer() after process of
// using the buffer.
func (a *theApp) LockBuffer() []byte {
	a.bufMu.Lock()
	if a.buf == nil {
		return []byte{}
	}
	return a.buf.Pix
}

// release mutex lock of underlying buffer.
func (a *theApp) UnlockBuffer([]byte) {
	a.bufMu.Unlock()
}

// entry point of main application. appconf nil is OK,
// use default if it is.
// its internal errors are handled by itself.
func (a *theApp) main(mobileDir string, config *Config) {
	if a.context == nil {
		panic("mobile.app.main(): run main with no application context")
	}
	if config == nil {
		panic("mobile.app.main(): nil config")
	}

	var retErr error
	defer func() {
		a.context.NotifyQuit(retErr)
	}()

	// set up application config.
	appConf, err := mobileConfig(mobileDir, config)
	if err != nil {
		retErr = err
		return
	}
	log.Info("Loading Config ... OK")

	// construct theme.
	t, err := app.BuildTheme(appConf)
	if err != nil {
		log.Info("Error: BuildTheme FAIL: ", err)
		retErr = err
		return
	}
	log.Info("Buinding Theme ... OK")

	// capture panic as error in this thread
	defer func() {
		if rec := recover(); rec != nil {
			buf := make([]byte, 4096)
			buf_end := runtime.Stack(buf, false)
			retErr = fmt.Errorf("%v\n%v\n", rec, string(buf[:buf_end]))
			log.Info("PANIC: ", retErr)
		}
	}()

	// run UI handler.
	log.Info("Running Main ...")
	if err := a.runWindow(t, appConf); err != nil {
		log.Info("Error: app.runWindow(): ", err)
		retErr = err
	} else {
		log.Info("...quiting correctly")
	}
}

// references golang.org/x/exp/shiny/widget/widget.go
func (a *theApp) runWindow(t *theme.Theme, appConf *app.Config) error {
	eventQ := &a.eventQ
	presenter := ui.NewEragoPresenter(eventQ)
	root := newUI(presenter, a.context)

	Theme := t

	var (
		paintPending bool = false
	)

	// because mobile app never sends mouse down event,
	// gesture.Filter also never used.
	// gef := gesture.EventFilter{EventDeque: eventQ}
	fps := app.FpsLimitter{EventDeque: eventQ, FPS: 30} // since mobile is low spec, but overhead exists.
	mef := app.ModelErrorFilter{}
	for {
		e := eventQ.NextEvent()

		// if e = gef.Filter(e); e == nil {
		// 	continue
		// }
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
					log.Info("RunGameThread() ... start OK")
					defer presenter.Quit()
				}
			}

		case gesture.Event, key.Event, mouse.Event:
			root.OnInputEvent(e, image.Point{})

		case app.PaintScheduled:
			// paint event is comming after some times
			// unmark paint request now.
			root.Marks.UnmarkNeedsPaint()

		case externalCmdEvent:
			root.ExternalCommand(e)

		case paint.Event:
			if err := a.paintBaseRoot(Theme, root); err != nil {
				return err
			}
			a.publish()
			paintPending = false

		case size.Event:
			if dpi := float64(e.PixelsPerPt) * unit.PointsPerInch; dpi != Theme.GetDPI() {
				newT := new(theme.Theme)
				if Theme != nil {
					*newT = *Theme
				}
				newT.DPI = dpi
				// in mobile, appConf.FontSize is DIP, but require argument is Pt.
				// so convert it.
				fontSize := newT.Convert(unit.DIPs(appConf.FontSize), unit.Pt).F

				catalog := newT.FontFaceCatalog.(*customTheme.OneFontFaceCatalog)
				catalog.UpdateFontFaceOptions(&customTheme.FontFaceOptions{
					DPI:  dpi,
					Size: fontSize,
				})
				Theme = newT
			}

			size := e.Size()
			a.size = size
			root.Measure(Theme, size.X, size.Y)
			root.Wrappee().Rect = e.Bounds()
			root.Layout(Theme)
			root.Wrappee().Mark(node.MarkNeedsPaintBase)

		case ui.PresenterTask:
			e.Run()

		case error:
			log.Debug("FATAL: UI's Fatal Error")
			return e
		}

		if !paintPending && root.Wrappee().Marks.NeedsPaintBase() {
			paintPending = true
			a.eventQ.Send(paint.Event{})
		}
	}
}

func (a *theApp) publish() {
	a.context.NotifyPaint()
}

func (a *theApp) paintBaseRoot(t *theme.Theme, root node.Node) error {
	a.bufMu.Lock()
	defer a.bufMu.Unlock()
	if a.buf != nil && a.buf.Bounds().Size() != a.size {
		a.buf = nil
	}
	if a.buf == nil {
		a.buf = image.NewRGBA(image.Rectangle{Max: a.size})
	}

	return root.PaintBase(&node.PaintBaseContext{
		Theme: t,
		Dst:   a.buf,
	}, image.Point{})
}
