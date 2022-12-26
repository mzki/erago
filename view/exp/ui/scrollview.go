package ui

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/mouse"
)

const (
	// ScrollBarWidthDp is a default width of scroll bar in ScrollView in Dip
	ScrollBarWidthDp = 32
	// ScrollBarMinHeightDp is a default minimum height of scroll bar in Dip.
	// ScrollBar height is smaller by increasing Scroller#MaximumStep, the minimum height
	// prevents making too small to visible.
	ScrollBarMinHeightDp = 32
	// ScrollButtonHeightDp is a default height of scroll button for both of up and down.
	ScrollButtonHeightDp = 32

	// ScrollDefaultVisibleStep is a default number of steps of content which is visible on view window.
	ScrollDefaultVisibleStep = 30
)

var (
	ScrollBarFgColor      = colornames.Grey600
	ScrollBarFgHoverColor = colornames.Grey400
	ScrollBarFgFocusColor = colornames.Grey300
	ScrollBarBgColor      = colornames.Grey900
)

// DiscreteScroller is a interface for scrollable view whose scroll is discreted by step.
// It cooperates with ScrollView.
type DiscreteScroller interface {
	// Scroll scrolls view by step. Positive step means lower content is visible, otherwise upper content is visible.
	Scroll(step int)
	// CurrentStep indicates current position of scroll region.
	// step =0 indicates top of content and =(MaximumStep-VisibleStep) indicates bottom end of content.
	// That is, current step can take [0:(MaximumStep-Visiblestep)].
	CurrentStep() int
	// MaximumStep indicates maximum poition of scroll region.
	MaximumStep() int
	// VisibleStep indicates number of steps of the content which is visble on view window.
	// Negative value means can not detect it by implementer, in such case use default value instead.
	VisibleStep() int
	// OnScroll registers callback function which is called every call of Scroll.
	// currStep and maxStep which is same as return value of CurrentStep and MaximumStep,
	// are passed to the callback.
	OnScroll(fn func(currStep, maxStep, visibleStep int))
}

// DiscreteScrollerNode is a composit interface of Scroller and node.Node.
type DiscreteScrollerNode interface {
	DiscreteScroller
	node.Node
}

// ScrollInner Implements DiscreteScrollerNode interface
type ScrollInner struct {
	DiscreteScroller
	node.Node
}

// NewScrollInnerFromChildNode is a helper function which takes inner Node
// and find object implementing Scroller interface by preorder travasal in tree of Node children.
// This function will panic when given node has no object implementing Scroller
// interface in the finding path.
func NewScrollInnerFromChildNode(n node.Node) *ScrollInner {
	scroller := findDiscreteScroller(n)
	if scroller == nil {
		panic("")
	}
	return &ScrollInner{
		Node:             n,
		DiscreteScroller: scroller,
	}
}

func findDiscreteScroller(n node.Node) DiscreteScroller {
	if s, ok := n.Wrappee().Wrapper.(DiscreteScroller); ok {
		return s
	} else {
		for c := n.Wrappee().FirstChild; c != nil; c = c.NextSibling {
			if s := findDiscreteScroller(c.Wrapper); s != nil {
				return s
			}
		}
	}
	return nil // not found
}

type ScrollView struct {
	*widget.Flow
	innerView DiscreteScrollerNode
	scrollbar *ScrollBar
}

func NewScrollView(inner DiscreteScrollerNode) *ScrollView {
	sbar := NewScrollBar()
	flow := widget.NewFlow(
		widget.AxisHorizontal,
		withStretch(inner, 1),
		withStretch(sbar, 0),
	)
	sv := &ScrollView{
		Flow:      flow,
		innerView: inner,
		scrollbar: sbar,
	}
	sv.Flow.Wrapper = sv

	// synchronized innerView and scrollbar position.
	sv.innerView.OnScroll(func(currStep, maxStep, visibleStep int) {
		// NOTE calling scrollbar.Scroll occurs infinite loop. avoid it.
		sv.scrollbar.Update(currStep, maxStep, visibleStep)
	})
	sv.scrollbar.OnScroll(func(currStep, maxStep, visibleStep int) {
		sv.innerView.Scroll(currStep - sv.innerView.CurrentStep())
	})
	return sv
}

// NewScrollViewFromNode creates ScrollView from node.Node.
// The node.Node should contains DiscreteScroller interface implementer in its child tree (including node itself).
func NewScrollViewFromNode(n node.Node) *ScrollView {
	return NewScrollView(NewScrollInnerFromChildNode(n))
}

// Implement node.Node interface.
func (sv *ScrollView) Layout(t *theme.Theme) {
	sv.Flow.Layout(t)
	// initialize scrollbar steps by innerView's one.
	sv.scrollbar.Update(sv.innerView.CurrentStep(), sv.innerView.MaximumStep(), sv.innerView.VisibleStep())
}

func (sv *ScrollView) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	sv.scrollbar.Update(sv.innerView.CurrentStep(), sv.innerView.MaximumStep(), sv.innerView.VisibleStep())
	return sv.Flow.PaintBase(ctx, origin)
}

func (sv *ScrollView) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	origin = origin.Add(sv.Rect.Min)
	var p image.Point
	switch e := e.(type) {
	case gesture.Event:
		p = image.Point{
			X: int(e.CurrentPos.X) - origin.X,
			Y: int(e.CurrentPos.Y) - origin.Y,
		}
	case mouse.Event:
		p = image.Point{
			X: int(e.X) - origin.X,
			Y: int(e.Y) - origin.Y,
		}
	}

	// Scrollbar has priority than innerView and not restricted in its Rectangle since holding
	// scrollbar takes over the innerView's region.
	if sv.scrollbar.OnInputEvent(e, origin) == node.Handled {
		return node.Handled
	}
	if p.In(sv.innerView.Wrappee().Rect) {
		return sv.innerView.Wrappee().Wrapper.OnInputEvent(e, origin)
	}
	return node.NotHandled
}

type scrollBarState int

const (
	scrollBarNone scrollBarState = iota
	scrollBarUnhover
	scrollBarHover
	scrollBarFocus
)

var scrollBarFgColorMap = map[scrollBarState]color.Color{
	scrollBarUnhover: ScrollBarFgColor,
	scrollBarHover:   ScrollBarFgHoverColor,
	scrollBarFocus:   ScrollBarFgFocusColor,
}

type ScrollBar struct {
	node.LeafEmbed

	currStep    int
	maxStep     int
	visibleStep int
	onScroll    func(int, int, int)

	scrollState    scrollBarState
	barRect        image.Rectangle
	prevGesturePos image.Point

	ScrollBarWidthDp     int
	ScrollBarMinHeightDp int
	ScrollButtonHeightDp int
	minBarHeightPxCache  int
}

func NewScrollBar() *ScrollBar {
	sbar := &ScrollBar{
		ScrollBarWidthDp:     ScrollBarWidthDp,
		ScrollBarMinHeightDp: ScrollBarMinHeightDp,
		ScrollButtonHeightDp: ScrollButtonHeightDp,
	}
	sbar.Wrapper = sbar
	return sbar
}

func (sbar *ScrollBar) fgColor() color.Color {
	if c, ok := scrollBarFgColorMap[sbar.scrollState]; ok {
		return c
	}
	return scrollBarFgColorMap[scrollBarUnhover] // fallback color
}

func (sbar *ScrollBar) availbleStep() int { return sbar.maxStep - sbar.visibleStep }
func (sbar *ScrollBar) availbleDy() int   { return sbar.Rect.Dy() - sbar.barRect.Dy() }

func (sbar *ScrollBar) clipCurrStep(currStep int) int {
	if currStep < 0 {
		currStep = 0
	} else if currStep > sbar.availbleStep() {
		currStep = sbar.availbleStep()
	}
	return currStep
}

// implement node.Node interface
func (sbar *ScrollBar) Measure(t *theme.Theme, widthHint int, heightHint int) {
	sbar.LeafEmbed.Measure(t, widthHint, heightHint)
	sbar.MeasuredSize.X = dpToPx(t, sbar.ScrollBarWidthDp)
	if twoBtnHeight := dpToPx(t, sbar.ScrollButtonHeightDp) * 2; twoBtnHeight > heightHint {
		sbar.MeasuredSize.Y = twoBtnHeight
	} else if minBarHeight := dpToPx(t, sbar.ScrollBarMinHeightDp); twoBtnHeight+minBarHeight > heightHint {
		sbar.MeasuredSize.Y = twoBtnHeight + minBarHeight
	} else {
		sbar.MeasuredSize.Y = heightHint
	}
	sbar.minBarHeightPxCache = dpToPx(t, ScrollBarMinHeightDp)
}

// Implement node.Node interface.
func (sbar *ScrollBar) Layout(t *theme.Theme) {
	sbar.LeafEmbed.Layout(t)
	sbar.scrollState = scrollBarNone // reset
	sbar.updateBarRect()
}

func (sbar *ScrollBar) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	draw.Draw(ctx.Dst, sbar.Rect, image.NewUniform(ScrollBarBgColor), image.Point{}, draw.Over)
	// TODO: draw scroll button is needed? it is less used nowaday?
	// rtop, rbottom := sbar.Rect, sbar.Rect
	// rtop.Max.Y = dpToPx(ctx.Theme, sbar.ScrollButtonHeightDp)
	// rbottom.Min.Y = rbottom.Max.Y - dpToPx(ctx.Theme, sbar.ScrollBarMinHeightDp)

	// draw scroll bar
	draw.Draw(ctx.Dst, sbar.barRect, image.NewUniform(sbar.fgColor()), image.Point{}, draw.Over)
	return sbar.LeafEmbed.PaintBase(ctx, origin)
}

// Implement node.Node interface.
func (sbar *ScrollBar) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	switch ev := e.(type) {
	case gesture.Event:
		p := image.Point{round(ev.CurrentPos.X), round(ev.CurrentPos.Y)}
		switch ev.Type {
		case gesture.TypeDrag:
			if sbar.scrollState == scrollBarFocus {
				if dy := p.Y - sbar.prevGesturePos.Y; dy != 0 {
					if minY := sbar.barRect.Min.Y + dy; minY < 0 {
						dy = dy - (minY - 0)
					} else if maxY := sbar.barRect.Max.Y + dy; maxY > sbar.Rect.Max.Y {
						dy = dy - (maxY - sbar.Rect.Max.Y)
					}
					sbar.barRect.Min.Y += dy
					sbar.barRect.Max.Y += dy
					sbar.updateCurrStepByBarRect()
					sbar.prevGesturePos = p
					return node.Handled
				}
			}
		case gesture.TypeTap:
			if !p.In(sbar.barRect) && p.In(sbar.Rect) && sbar.scrollState != scrollBarFocus {
				nextStep := sbar.currStepAt(p.Y)
				nextStep -= sbar.visibleStep / 2 // to move middle of scroll bar
				nextStep = sbar.clipCurrStep(nextStep)
				sbar.Scroll(nextStep - sbar.currStep)
				return node.Handled
			}
		}
	case mouse.Event:
		p := image.Point{round(ev.X), round(ev.Y)}
		switch {
		case ev.Button == mouse.ButtonLeft && ev.Direction == mouse.DirPress:
			if p.In(sbar.barRect) {
				sbar.scrollState = scrollBarFocus
				sbar.prevGesturePos = p
				sbar.Mark(node.MarkNeedsPaintBase)
				return node.Handled
			}
		case ev.Button == mouse.ButtonLeft && ev.Direction == mouse.DirRelease:
			if sbar.scrollState == scrollBarFocus {
				if p.In(sbar.barRect) {
					sbar.scrollState = scrollBarHover
				} else {
					sbar.scrollState = scrollBarUnhover
				}
				sbar.prevGesturePos = image.Point{}
				sbar.Mark(node.MarkNeedsPaintBase)
				return node.Handled
			}
		default:
			// handle mouse moving without drag
			// This event is also used by other view, so NOT returning node.Handled.
			if sbar.scrollState != scrollBarFocus {
				if p.In(sbar.barRect) {
					sbar.scrollState = scrollBarHover
					sbar.Mark(node.MarkNeedsPaintBase)
				} else if sbar.scrollState == scrollBarHover {
					sbar.scrollState = scrollBarUnhover
					sbar.Mark(node.MarkNeedsPaintBase)
				}
			}
		}
	}
	return node.NotHandled // consume to not propagate event for other views
}

const scrollbarFallbackDStep = 80

func (sbar *ScrollBar) Update(currStep, maxStep, visibleStep int) {
	if sbar.currStep != currStep || sbar.maxStep != maxStep || sbar.visibleStep != visibleStep {
		sbar.currStep = currStep
		sbar.maxStep = maxStep
		sbar.visibleStep = visibleStep
		sbar.updateBarRect()
	}
}

func (sbar *ScrollBar) updateBarRect() {
	if sbar.maxStep == 0 {
		sbar.maxStep = scrollbarFallbackDStep + sbar.visibleStep
	}
	sy := float64(sbar.Rect.Dy()) * float64(sbar.currStep) / float64(sbar.maxStep)
	dy := float64(sbar.Rect.Dy()) * float64(sbar.visibleStep) / float64(sbar.maxStep)
	if dyi := int(math.Round(dy)); dyi < sbar.minBarHeightPxCache {
		dy = float64(sbar.minBarHeightPxCache)
	}
	barR := sbar.Rect
	barR.Min.Y = int(math.Round(sy))
	barR.Max.Y = int(math.Round(sy + dy))
	if barR.Min.Y < sbar.Rect.Min.Y {
		barR.Min.Y = sbar.Rect.Min.Y
	}
	if barR.Max.Y > sbar.Rect.Max.Y {
		barR.Max.Y = sbar.Rect.Max.Y
	}
	sbar.barRect = barR // update
	sbar.Mark(node.MarkNeedsPaintBase)
}

func (sbar *ScrollBar) updateCurrStepByBarRect() {
	syi := sbar.currStepAt(sbar.barRect.Min.Y)
	syi = sbar.clipCurrStep(syi)
	if sbar.currStep != syi {
		sbar.currStep = syi
		// updating currStep means scrolling, need to notify updated currStep
		if fn := sbar.onScroll; fn != nil {
			fn(sbar.currStep, sbar.maxStep, sbar.visibleStep)
		}
		sbar.Mark(node.MarkNeedsPaintBase)
	}
}

func (sbar *ScrollBar) currStepAt(atPx int) int {
	currStepF := float64(atPx-sbar.Rect.Min.Y) / float64(sbar.Rect.Dy()) * float64(sbar.maxStep)
	currStep := int(math.Round(currStepF))
	return currStep
}

// --- Implement DiscreteScroller interface ---

func (sbar *ScrollBar) Scroll(step int) {
	sbar.currStep += step
	if sbar.currStep > sbar.availbleDy() {
		sbar.currStep = sbar.availbleDy()
	} else if sbar.currStep < 0 {
		sbar.currStep = 0
	}
	sbar.updateBarRect()
	if fn := sbar.onScroll; fn != nil {
		fn(sbar.currStep, sbar.maxStep, sbar.visibleStep)
	}
}

func (sbar *ScrollBar) CurrentStep() int                                     { return sbar.currStep }
func (sbar *ScrollBar) MaximumStep() int                                     { return sbar.maxStep }
func (sbar *ScrollBar) VisibleStep() int                                     { return sbar.visibleStep }
func (sbar *ScrollBar) OnScroll(fn func(currStep, maxStep, visibleStep int)) { sbar.onScroll = fn }
