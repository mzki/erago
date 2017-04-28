package ui

import (
	"fmt"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"

	customT "local/erago/view/exp/theme"
)

// these are used to focused frame or unfocused frame.
var (
	focusFrameColor   = theme.Foreground
	unfocusFrameColor = customT.NeutralDivider
)

func unfocusFrame(inner node.Node) *Frame {
	return NewFrame(unit.DIPs(2), unfocusFrameColor, inner)
}

type textVNode struct {
	View *TextView

	// Frame is parent of TextView to paint frame.
	// to change frame color, it is referenced.
	Frame *Frame

	// Node may not same as View's Wrapper Node
	// because Node is often stacked to decorate View.
	// In other words, Node's last child is View.
	Node node.Node
}

// focus as current view.
func (tv *textVNode) Focus() {
	tv.View.Focus()
	tv.Frame.ThemeColor = focusFrameColor
}

// unfocus as not current view.
func (tv *textVNode) Unfocus() {
	tv.View.Unfocus()
	tv.Frame.ThemeColor = unfocusFrameColor
}

// viewManager manages multiple views manually.
type viewManager struct {
	// current is currently used TextView to print string.
	// All of TextView are managed by string-map using view's name.
	current   *textVNode
	textViews map[string]*textVNode

	sender *EragoPresenter
}

func newViewManager(vname string, sender *EragoPresenter) *viewManager {
	vm := &viewManager{
		textViews: make(map[string]*textVNode, 2),
		sender:    sender,
	}
	vm.current = vm.appendTextView(vname)
	return vm
}

// vanish all views.
func (vm *viewManager) removeAll() {
	for _, name := range vm.getViewNames() {
		vm.remove(name)
	}
	vm.current = nil
}

func (vm *viewManager) remove(vname string) {
	tv, err := vm.findViewNode(vname)
	if err != nil {
		// not found, do nothing
		return
	}
	tv.View.Close()
	gotoLifecycleStageDead(tv.Node)
	vm.purge(vname)
	delete(vm.textViews, vname)
}

// purge breaks off tree connection between v's node and
// its parent.
func (vm *viewManager) purge(vname string) {
	n := vm.mustViewNode(vname)
	if p := n.Wrappee().Parent; p != nil {
		p.Wrapper.Remove(n)
	}
}

func (vm *viewManager) setCurrentView(vname string) error {
	v, err := vm.findViewNode(vname)
	if err != nil {
		return err
	}
	vm.current.Unfocus()
	vm.current = v
	vm.current.Focus()
	return nil
}

func (vm viewManager) currentView() *TextView {
	return vm.current.View
}

func (vm viewManager) currentViewNode() node.Node {
	return vm.current.Node
}

func errorViewNotFound(vname string) error {
	return fmt.Errorf("viewManager: view `%s` is not found.")
}

func (vm viewManager) findViewNode(name string) (*textVNode, error) {
	if tv, ok := vm.textViews[name]; ok {
		return tv, nil
	}
	return nil, errorViewNotFound(name)
}

func (vm viewManager) mustView(name string) *TextView {
	v, err := vm.findViewNode(name)
	if err != nil {
		panic(err)
	}
	return v.View
}

func (vm viewManager) mustViewNode(name string) node.Node {
	v, err := vm.findViewNode(name)
	if err != nil {
		panic(err)
	}
	return v.Node
}

func (vm viewManager) getViewNames() []string {
	names := make([]string, 0, len(vm.textViews))
	for _, tv := range vm.textViews {
		names = append(names, tv.View.name)
	}
	return names
}

func (vm *viewManager) appendTextView(vname string) *textVNode {
	newV := NewTextView(vname, vm.sender)
	f := unfocusFrame(newV)
	tv := &textVNode{
		Node:  f,
		Frame: f,
		View:  newV,
	}
	tv.Unfocus()
	vm.textViews[vname] = tv
	return tv
}

// it does not use vm internally, but to create formated new image view, use this method.
func (vm *viewManager) newImageView(src string) node.Node {
	img := NewImageView(src)
	return unfocusFrame(img)
}
