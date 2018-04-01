package ui

import (
	"errors"
	"image"
	"image/draw"
	"sync"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	attr "local/erago/attribute"
)

// TabView is a container widget which paints only current
// view and has tabs for selecting view to paint.
// It implements uiadapter.UI interface, the functions accessing current view are
// called through currentView.Printer.
type TabView struct {
	*widget.Flow

	layoutLocker *sync.Mutex
	labels       *tabViewLabels
	content      *tabViewContent
}

func NewTabView(sender *EragoPresenter) *TabView {
	content := newTabViewContent(sender)
	labels := newTabViewLabels([]string{content.currentNodeName()}, func(tab *tabLabel) {
		if ok := content.setCurrentNode(tab.text); !ok {
			panic("onSelect: select no exist tab: " + tab.text)
		}
	})

	v := &TabView{
		Flow: widget.NewFlow(widget.AxisVertical,
			withStretch(labels, 0),
			withStretch(content, 1),
		),
		layoutLocker: new(sync.Mutex),
		labels:       labels,
		content:      content,
	}
	v.Wrapper = v
	return v
}

func (v *TabView) Measure(t *theme.Theme, widthHint, heightHint int) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.MeasuredSize = image.Point{widthHint, heightHint}
	v.measure(t)
}

func (v *TabView) measure(t *theme.Theme) {
	v.Flow.Measure(t, v.MeasuredSize.X, v.MeasuredSize.Y)
}

func (v *TabView) Layout(t *theme.Theme) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.layout(t)
}

// change view layout according to layoutData. it doesnt use concurrently, use layoutLocker before call this.
func (v *TabView) layout(t *theme.Theme) {
	v.Flow.Layout(t)
}

func (v *TabView) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.Flow.OnInputEvent(ev, origin)
}

func (v *TabView) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.Flow.PaintBase(ctx, origin)
}

func (v *TabView) Paint(ctx *node.PaintContext, origin image.Point) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.Flow.Paint(ctx, origin)
}

// // These methods are implementation of uiadapter.UI.
//

func (v *TabView) Print(s string) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().Print(s)
}

func (v *TabView) PrintLabel(s string) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().PrintLabel(s)
}

func (v *TabView) PrintButton(caption string, command string) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().PrintButton(caption, command)
}

func (v *TabView) PrintLine(sym string) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().PrintLine(sym)
}

func (v *TabView) SetColor(color uint32) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().SetColor(color)
}

func (v *TabView) GetColor() (color uint32, err error) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().GetColor()
}

func (v *TabView) ResetColor() error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().ResetColor()
}

func (v *TabView) SetAlignment(a attr.Alignment) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().SetAlignment(a)
}

func (v *TabView) GetAlignment() (attr.Alignment, error) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().GetAlignment()
}

func (v *TabView) NewPage() {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().NewPage()
}

func (v *TabView) ClearLine(nline int) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().ClearLine(nline)
}

func (v *TabView) ClearLineAll() {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	v.content.viewManager.currentView().ClearLineAll()
}

func (v *TabView) MaxRuneWidth() (int, error) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().MaxRuneWidth()
}

func (v *TabView) CurrentRuneWidth() (int, error) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().CurrentRuneWidth()
}

func (v *TabView) LineCount() (int, error) {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().LineCount()
}

func (v *TabView) SetLayout(layout *attr.LayoutData) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()

	list, err := newViewList(layout, v.content.viewManager)
	if err != nil {
		return err
	}

	names := extractViewNames(layout)
	// check whether node list and node names have identical length?
	if len(names) != len(list) {
		return errors.New("different length between nodes and names")
	}

	v.labels.resetTabs(names)
	v.content.removeAllNodes()
	for i, n := range list {
		v.content.setNode(n, names[i])
	}

	// make current tab same as the current view.
	cname := v.labels.currentTabName()
	if !v.content.setCurrentNode(cname) {
		return errors.New("TabView.SetLayout(): invalid state")
	}
	return nil
}

func (v *TabView) SetSingleLayout(name string) error {
	// TODO: move to uiadapter.OutputPort?
	panic("not implemented")
}

func (v *TabView) SetHorizontalLayout(vname1 string, vname2 string, rate float64) error {
	panic("not implemented")
}

func (v *TabView) SetVerticalLayout(vname1 string, vname2 string, rate float64) error {
	panic("not implemented")
}

func (v *TabView) SetCurrentView(vname string) error {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.setCurrentView(vname)
}

func (v *TabView) GetCurrentViewName() string {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.currentView().name
}

func (v *TabView) GetViewNames() []string {
	v.layoutLocker.Lock()
	defer v.layoutLocker.Unlock()
	return v.content.viewManager.getViewNames()
}

// // TabView's parts

// A content part of TabView.
type tabViewContent struct {
	// ContainerEmbed has TabView's nodes as a children.
	// childen may be changed dynamically from other goroutine.
	node.ContainerEmbed
	sender *EragoPresenter

	viewManager *viewManager

	currentNode node.Node
	nodes       map[string]node.Node
	// theme is cached to measure and layout by itself.
	// theme's fields are must not be modified.
	theme *theme.Theme
}

func newTabViewContent(sender *EragoPresenter) *tabViewContent {
	if sender == nil {
		panic("nil sender is not allowed")
	}
	v := &tabViewContent{
		sender:      sender,
		viewManager: newViewManager(firstViewName, sender),
		nodes:       make(map[string]node.Node, 4),
	}
	v.Wrapper = v

	cn := v.viewManager.currentViewNode()
	cname := v.viewManager.currentView().name
	v.setNode(cn, cname)
	v.setCurrentNode(cname)
	return v
}

// implements node.Node interface.
func (v *tabViewContent) Measure(t *theme.Theme, widthHint, heightHint int) {
	for c := v.FirstChild; c != nil; c = c.NextSibling {
		c.Wrapper.Measure(t, widthHint, heightHint)
	}
	v.MeasuredSize = v.FirstChild.MeasuredSize
}

// implements node.Node interface.
func (v *tabViewContent) Layout(t *theme.Theme) {
	for c := v.FirstChild; c != nil; c = c.NextSibling {
		c.Rect = v.Rect
		c.Wrapper.Layout(t)
	}
	v.theme = t
}

// implements node.Node interface.
func (v *tabViewContent) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	v.Marks.UnmarkNeedsPaintBase()
	// because currentNode has same Rect as tabViewContent,
	// the arguments are passed directly.
	return v.currentNode.PaintBase(ctx, origin)
}

// implements node.Node interface.
func (v *tabViewContent) Paint(ctx *node.PaintContext, origin image.Point) error {
	v.Marks.UnmarkNeedsPaint()
	// because currentNode has same Rect as tabViewContent,
	// the arguments are passed directly.
	return v.currentNode.Paint(ctx, origin)
}

// implements node.Node interface.
func (v *tabViewContent) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	if v.currentNode.OnInputEvent(ev, origin) == node.Handled {
		return node.Handled
	}
	// TODO: change currentNode by gesture horizontal fling?
	return node.NotHandled
}

// return node's name is found?
func (v *tabViewContent) setCurrentNode(name string) bool {
	if n, ok := v.nodes[name]; ok {
		v.currentNode = n
		return true
	}
	return false
}

func (v *tabViewContent) currentNodeName() string {
	for name, n := range v.nodes {
		if n == v.currentNode {
			return name
		}
	}
	return ""
}

// return set new node? if already exist, return false.
func (v *tabViewContent) setNode(n node.Node, name string) bool {
	if _, ok := v.nodes[name]; ok {
		return false
	}
	v.nodes[name] = n
	v.Insert(n, nil)
	return true
}

func (v *tabViewContent) removeAllNodes() {
	for k, n := range v.nodes {
		// TODO: n may be purged from parent, in which Remove occurs panic.
		if n.Wrappee().Parent != nil {
			v.Remove(n)
		}
		delete(v.nodes, k)
	}
	v.currentNode = nil
}

// A label part of TabView.
type tabViewLabels struct {
	*widget.Flow

	currentTab *tabLabel
	tabs       []*tabLabel
	onSelect   func(*tabLabel)
}

func newTabViewLabels(names []string, onSelect func(*tabLabel)) *tabViewLabels {
	v := &tabViewLabels{
		Flow:     widget.NewFlow(widget.AxisHorizontal),
		tabs:     make([]*tabLabel, 0, len(names)),
		onSelect: onSelect,
	}
	v.Wrapper = v

	v.resetTabs(names)
	return v
}

func (v *tabViewLabels) resetTabs(names []string) {
	if len(names) == 0 {
		panic("TabViewLables: require more than one names")
	}
	for _, tab := range v.tabs {
		v.Flow.Remove(tab)
	}
	v.tabs = make([]*tabLabel, 0, len(names))
	for _, name := range names {
		v.addTab(name)
	}
	v.currentTab = v.tabs[0]
	v.currentTab.focus()
}

func (v *tabViewLabels) addTab(label string) {
	tab := widget.WithLayoutData(newTabLabel(label), widget.FlowLayoutData{
		AlongWeight: 1,
		ShrinkAlong: true,
	}).(*tabLabel)
	v.Insert(tab, nil)
	v.tabs = append(v.tabs, tab)
}

func (v *tabViewLabels) currentTabName() string { return v.currentTab.text }

func (v *tabViewLabels) OnInputEvent(ev interface{}, origin image.Point) node.EventHandled {
	switch ev := ev.(type) {
	case gesture.Event:
		switch ev.Type {
		case gesture.TypeTap:
			p := image.Point{round(ev.CurrentPos.X), round(ev.CurrentPos.Y)}
			if !ev.DoublePress && v.selectTab(p, origin) {
				v.Mark(node.MarkNeedsPaintBase)
				return node.Handled
			}
		}
	}
	return node.NotHandled
}

func (v *tabViewLabels) selectTab(at, origin image.Point) bool {
	p := at.Sub(origin).Sub(v.Rect.Min) // in v.Rect coordinate.
	for _, t := range v.tabs {
		if p.In(t.Rect) {
			v.currentTab.unfocus()
			v.currentTab = t
			t.focus()
			v.onSelect(t)
			return true
		}
	}
	return false
}

func (v *tabViewLabels) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	dstRect := v.Rect.Add(origin)
	draw.Draw(ctx.Dst, dstRect, theme.Background.Uniform(ctx.Theme), image.Point{}, draw.Src)
	return v.Flow.PaintBase(ctx, origin)
}

// tabLabel is a each label of tabs.
type tabLabel struct {
	node.LeafEmbed

	text    string
	focused bool

	heightCache        int
	widthPaddingCache  int
	bottomPaddingCache int
}

func newTabLabel(label string) *tabLabel {
	t := &tabLabel{text: label}
	t.Wrapper = t
	return t
}

func (t *tabLabel) focus()   { t.focused = true }
func (t *tabLabel) unfocus() { t.focused = false }

const (
	// tab size specs in dip.
	tabHeightDIP        = 48
	tabWidthPaddingDIP  = 12
	tabBottomPaddingDIP = 20
	tabIndicatorDIP     = 2
)

func (tab *tabLabel) Measure(t *theme.Theme, widthHint, heightHint int) {
	height := t.Pixels(unit.DIPs(tabHeightDIP)).Ceil()
	tab.heightCache = height

	wPadding := t.Pixels(unit.DIPs(tabWidthPaddingDIP)).Round()
	tab.widthPaddingCache = wPadding

	wText := t.Pixels(unit.Chs(float64(len(tab.text)))).Round()
	tab.MeasuredSize = image.Point{wPadding*2 + wText, height}

	tab.bottomPaddingCache = t.Pixels(unit.DIPs(tabBottomPaddingDIP)).Round()
}

func (t *tabLabel) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	t.Marks.UnmarkNeedsPaintBase()

	Theme := ctx.Theme
	drawRect := t.Rect.Add(origin)
	if drawRect.Empty() {
		return nil
	}
	// draw Background
	draw.Draw(ctx.Dst, drawRect, theme.Background.Uniform(Theme), image.Point{}, draw.Src)

	// draw Foreground.
	// TODO: 12 dp font size by Theme.Convert(unit.DIPs(12), unit.Pt)?
	face := Theme.AcquireFontFace(theme.FontFaceOptions{})
	descent := face.Metrics().Descent
	drawer := &font.Drawer{
		Dst:  ctx.Dst,
		Src:  theme.Foreground.Uniform(Theme),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(drawRect.Min.X + t.widthPaddingCache),
			Y: fixed.I(drawRect.Max.Y-t.bottomPaddingCache) - descent,
		},
	}
	drawer.DrawString(t.text)
	Theme.ReleaseFontFace(theme.FontFaceOptions{}, face)

	if t.focused {
		// set drawRect as bar with indicator's height
		ind_h := Theme.Pixels(unit.DIPs(tabIndicatorDIP)).Round()
		drawRect.Min.Y = drawRect.Max.Y - ind_h
		if drawRect.Min.Y < 0 {
			drawRect.Min.Y = 0
		}
		draw.Draw(ctx.Dst, drawRect, theme.Accent.Uniform(Theme), image.Point{}, draw.Src)
	}
	return nil
}
