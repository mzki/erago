package ui

import (
	"errors"
	"image"
	"sync"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/lifecycle"

	attr "local/erago/attribute"
	"local/erago/util/log"
)

// MultipleView is abstruct of multiple views which has current view state.
// It implements uiadapter.UI interface, the functions accessing
// current view are called through currentView.Printer.
type MultipleView struct {
	// ShellEmbed has MultipleView's root as a child.
	// a child may be changed dynamically from other goroutine.
	node.ShellEmbed
	sender *EragoPresenter

	// mutex for changing layout and drawing content.
	layoutLocker *sync.Mutex
	viewManager  *viewManager
	root         node.Node
	// theme is cached to measure and layout by itself.
	// theme's fields are must not be modified.
	theme *theme.Theme
}

const firstViewName = "default"

func NewMultipleView(sender *EragoPresenter) *MultipleView {
	if sender == nil {
		panic("nil sender is not allowed")
	}
	mv := &MultipleView{
		viewManager:  newViewManager(firstViewName, sender),
		sender:       sender,
		layoutLocker: new(sync.Mutex),
	}
	mv.Wrapper = mv

	n := mv.viewManager.currentViewNode()
	mv.root = n
	mv.ShellEmbed.Insert(n, nil)
	return mv
}

// finalize MultipleView. Multiple calling It is OK but do nothing.
// It is called automatically when lifecycle stage go to StageDead.
// but for unexpected panic, it is exported.
func (mv *MultipleView) Close() {
	mv.viewManager.removeAll()
	gotoLifecycleStageDead(mv.root)
}

func (mv *MultipleView) OnLifeCycleEvent(e lifecycle.Event) {
	mv.root.OnLifecycleEvent(e)
	if e.To == lifecycle.StageDead {
		mv.Close()
	}
}

func (mv *MultipleView) Measure(t *theme.Theme, widthHint, heightHint int) {
	mv.MeasuredSize = image.Point{widthHint, heightHint}
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	mv.theme = t
	mv.measure(t)
}

func (mv *MultipleView) measure(t *theme.Theme) {
	mv.root.Measure(t, mv.MeasuredSize.X, mv.MeasuredSize.Y)
}

func (mv *MultipleView) Layout(t *theme.Theme) {
	mv.layoutLocker.Lock()
	mv.theme = t
	mv.layout(t)
	mv.layoutLocker.Unlock()
}

// change view layout according to layoutData. it doesnt use concurrently, use layoutLocker before call this.
func (mv *MultipleView) layout(t *theme.Theme) {
	mv.root.Wrappee().Rect = mv.Rect
	mv.root.Layout(t)
}

func (mv *MultipleView) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	mv.layoutLocker.Lock()
	n := mv.root
	handled := n.OnInputEvent(ev, origin.Add(mv.Rect.Min))
	// if n.Wrappee().Marks.NeedsPaintBase() { // TODO: propagates root's onMarkChanged to mv
	// 	mv.Mark(node.MarkNeedsPaintBase)
	// }
	mv.layoutLocker.Unlock()
	if handled == node.Handled {
		return node.Handled
	}

	// the followings are to react for the input.
	switch ev := ev.(type) {
	case gesture.Event:
		switch ev.Type {
		case gesture.TypeTap:
			if ev.DoublePress { // exactly tap at once
				return node.NotHandled
			}
			mv.sender.SendCommand("")
			return node.Handled
		case gesture.TypeIsDoublePress:
			mv.sender.SendControlSkippingWait(true)
			return node.Handled
		}
	}
	return node.NotHandled
}

// implements node.Node interface.
func (mv *MultipleView) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	mv.Marks.UnmarkNeedsPaintBase()

	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()

	// // TODO: propagates each View's MarkPaintBase to MultipleView.
	// if root := mv.root; root.Wrappee().Marks.NeedsPaintBase() {
	// 	return root.PaintBase(ctx, origin.Add(mv.Rect.Min))
	// }
	mv.root.PaintBase(ctx, origin.Add(mv.Rect.Min))
	return nil
}

// implements node.Node interface.
func (mv *MultipleView) Paint(ctx *node.PaintContext, origin image.Point) error {
	mv.Marks.UnmarkNeedsPaint()

	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()

	return mv.root.Paint(ctx, origin.Add(mv.Rect.Min))
}

// implements fmt.Stringer interface.
func (mv *MultipleView) String() string {
	return mv.viewManager.currentView().String()
}

// implement uiadapter.UI.
func (mv *MultipleView) SetLayout(l *attr.LayoutData) error {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.setLayout(l)
}

func (mv *MultipleView) setLayout(l *attr.LayoutData) error {
	log.Debug("MultipleView.setLayout(): changing layout dynamically")
	newRoot, err := newNodeTree(l, mv.viewManager)
	if err != nil {
		return err
	}
	if len(mv.viewManager.textViews) == 0 {
		return errors.New("MultipleView.setLayout(): new layout must have at least 1 TextView")
	}

	if mv.ShellEmbed.FirstChild != nil {
		gotoLifecycleStageDead(mv.root)
		mv.ShellEmbed.Remove(mv.root)
	}
	mv.ShellEmbed.Insert(newRoot, nil)
	mv.root = newRoot
	mv.measure(mv.theme)
	mv.layout(mv.theme)
	mv.sender.Mark(mv, node.MarkNeedsPaintBase)
	return nil
}

var errorEmptyNameNotAllowed = errors.New("MultipleView: emtpy view name is not allowed.")

// implement uiadapter.UI.
func (mv *MultipleView) SetSingleLayout(vname string) error {
	if vname == "" {
		return errorEmptyNameNotAllowed
	}
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.setLayout(attr.NewSingleText(vname))
}

// implement uiadapter.UI.
func (mv *MultipleView) SetHorizontalLayout(vname1, vname2 string, rate float64) error {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.setDoubleLayout(vname1, vname2, rate, widget.AxisHorizontal)
}

// implement uiadapter.UI.
func (mv *MultipleView) SetVerticalLayout(vname1, vname2 string, rate float64) error {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.setDoubleLayout(vname1, vname2, rate, widget.AxisVertical)
}

var errorInvalidBorderRate = errors.New("MultipleView: invalid rate of layouting border position")

func (mv *MultipleView) setDoubleLayout(vname1, vname2 string, rate float64, axis widget.Axis) error {
	if vname1 == "" || vname2 == "" {
		return errorEmptyNameNotAllowed
	}
	if rate <= 0.0 || rate >= 1.0 {
		return errorInvalidBorderRate
	}

	weight1 := int(10.0 * rate)
	weight2 := 10 - weight1

	text1 := attr.WithParentValue(attr.NewSingleText(vname1), weight1)
	text2 := attr.WithParentValue(attr.NewSingleText(vname2), weight2)

	if axis == widget.AxisVertical {
		return mv.setLayout(attr.NewFlowVertical(text1, text2))
	} else {
		return mv.setLayout(attr.NewFlowHorizontal(text1, text2))
	}
}

// implement uiadapter.UI.
func (mv *MultipleView) SetCurrentView(vname string) error {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	mv.sender.Mark(mv, node.MarkNeedsPaint) // TODO: this line move to each text View?
	return mv.viewManager.setCurrentView(vname)
}

// implement uiadapter.UI.
func (mv *MultipleView) GetCurrentViewName() string {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.viewManager.currentView().name
}

// implement uiadapter.UI.
func (mv *MultipleView) GetViewNames() []string {
	mv.layoutLocker.Lock()
	defer mv.layoutLocker.Unlock()
	return mv.viewManager.getViewNames()
}

//
// these functions implements uiadapter.Printer interface and,
// do not need mutex locking since these do not change internal layout.
//

func (mv *MultipleView) Print(s string) {
	mv.viewManager.currentView().Print(s)
}

func (mv *MultipleView) PrintLabel(s string) {
	mv.viewManager.currentView().PrintLabel(s)
}

func (mv *MultipleView) PrintButton(caption string, command string) {
	mv.viewManager.currentView().PrintButton(caption, command)
}

func (mv *MultipleView) PrintLine(sym string) {
	mv.viewManager.currentView().PrintLine(sym)
}

func (mv *MultipleView) SetColor(color uint32) {
	mv.viewManager.currentView().SetColor(color)
}

func (mv *MultipleView) GetColor() (color uint32) {
	return mv.viewManager.currentView().GetColor()
}

func (mv *MultipleView) ResetColor() {
	mv.viewManager.currentView().ResetColor()
}

func (mv *MultipleView) SetAlignment(a attr.Alignment) {
	mv.viewManager.currentView().SetAlignment(a)
}

func (mv *MultipleView) GetAlignment() attr.Alignment {
	return mv.viewManager.currentView().GetAlignment()
}

func (mv *MultipleView) NewPage() {
	mv.viewManager.currentView().NewPage()
}

func (mv *MultipleView) ClearLine(nline int) {
	mv.viewManager.currentView().ClearLine(nline)
}

func (mv *MultipleView) ClearLineAll() {
	mv.viewManager.currentView().ClearLineAll()
}

func (mv *MultipleView) MaxRuneWidth() int {
	return mv.viewManager.currentView().MaxRuneWidth()
}

func (mv *MultipleView) CurrentRuneWidth() int {
	return mv.viewManager.currentView().CurrentRuneWidth()
}

func (mv *MultipleView) LineCount() int {
	return mv.viewManager.currentView().LineCount()
}
