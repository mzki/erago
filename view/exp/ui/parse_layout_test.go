package ui

import (
	"errors"
	"testing"

	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"

	attr "local/erago/attribute"
)

var testLayoutData = attr.NewFlowVertical(
	attr.NewSingleText("1"),
	attr.NewSingleText("2"),
	attr.NewFlowHorizontal(
		attr.NewSingleText("3"),
		attr.NewSingleImage("nothing"),
		attr.NewFixedSplit(attr.EdgeTop,
			attr.WithParentValue(attr.NewSingleText("4"), 10),
			attr.NewSingleText("5"),
		),
	),
)

func buildTestViewManager() *viewManager {
	return newViewManager("default", NewEragoPresenter(eventQueueStub{}))
}

var testLayoutDataViewManager = buildTestViewManager()

func removeTestLayoutData(vm *viewManager) {
	for _, s := range []string{"1", "2", "3", "4", "5"} {
		testLayoutDataViewManager.remove(s)
	}
}

func TestParseLayout(t *testing.T) {
	_, err := parseLayoutData(testLayoutData, testLayoutDataViewManager)
	if err != nil {
		t.Fatal(err)
	}

	vm := testLayoutDataViewManager
	for _, s := range []string{"1", "2", "3", "4", "5"} {
		if _, err := vm.findViewNode(s); err != nil {
			t.Error(err)
		}
	}
}

func TestNewNodeTree(t *testing.T) {
	vm := testLayoutDataViewManager
	if vm.current.View.closed {
		t.Fatal("current view is closed, why?")
	}

	_, err := newNodeTree(attr.NewSingleText(vm.currentView().name), testLayoutDataViewManager)
	if err != nil {
		t.Fatal(err)
	}

	if vm.current.View.closed {
		t.Fatal("after layout itself, current view is closed, why?")
	}
}

func TestNewViewList(t *testing.T) {
	vm := buildTestViewManager()
	_, err := newViewList(testLayoutData, vm)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"1", "2", "3", "4", "5"} {
		if _, err := vm.findViewNode(s); err != nil {
			t.Error(err)
		}
	}
	if _, err := vm.findViewNode("default"); err == nil {
		t.Error("old view is remained")
	}
}

var benchLayoutData = attr.NewFlowVertical(
	attr.NewSingleText("1"),
	attr.NewSingleText("2"),
	attr.NewFlowHorizontal(
		attr.NewSingleText("3"),
		attr.NewFixedSplit(attr.EdgeTop,
			attr.WithParentValue(attr.NewSingleText("4"), 10),
			attr.NewSingleText("5"),
		),
	),
)

func BenchmarkParseLayoutData1(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseLayoutData(benchLayoutData, testLayoutDataViewManager)
		if err != nil {
			b.Fatal(err)
		}
		_ = extractTextViewNames(benchLayoutData)
	}
}

func BenchmarkParseLayoutData2(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names := make([]string, 0, 4)
		_, _, err := parseLayoutData2(benchLayoutData, names, testLayoutDataViewManager)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func parseLayoutData2(d *attr.LayoutData, vnames []string, vm *viewManager) (node.Node, []string, error) {
	switch d.Type {
	case attr.TypeSingleText:
		vname := d.SingleTextName()
		vnames = append(vnames, vname)
		if _, err := vm.findViewNode(vname); err != nil {
			// vname not found, build new view with vname.
			newV := vm.appendTextView(vname)
			vm.setCurrentView(vname)
			return newV.Node, vnames, nil
		}
		// vname found, reuse it.
		vm.purge(vname)
		vm.setCurrentView(vname)
		return vm.mustViewNode(vname), vnames, nil

	case attr.TypeSingleImage:
		return nil, vnames, errors.New("not implement")

	case attr.TypeFlowVertical, attr.TypeFlowHorizontal:
		if len(d.Children) == 0 {
			return nil, nil, errors.New("parseLayoutData: TypeFlow: empty children not allowed")
		}

		children := make([]node.Node, 0, len(d.Children))
		for _, c := range d.Children {
			if c == nil {
				continue
			}
			var n node.Node
			var err error
			n, vnames, err = parseLayoutData2(c, vnames, vm)
			if err != nil {
				return nil, nil, err
			}
			weight := c.FlowChildWeight()
			children = append(children, withStretch(n, weight))
		}

		axis := widget.AxisVertical
		if d.Type == attr.TypeFlowHorizontal {
			axis = widget.AxisHorizontal
		}
		return widget.NewFlow(axis, children...), vnames, nil

	case attr.TypeFixedSplit:
		if len(d.Children) == 0 {
			return nil, nil, errors.New("parseLayoutData: TypeFixed: empty children not allowed")
		}

		children := make([]node.Node, 0, len(d.Children))
		for _, c := range d.Children {
			if c == nil {
				continue
			}
			var n node.Node
			var err error
			n, vnames, err = parseLayoutData2(c, vnames, vm)
			if err != nil {
				return nil, nil, err
			}
			size := c.FixedChildSize()
			children = append(children, withStretch(n, size)) // TODO: fixed strech
		}

		axis := widget.AxisVertical
		if d.Type == attr.TypeFlowHorizontal {
			axis = widget.AxisHorizontal
		}
		// TODO: implement widget.Fixed
		return widget.NewFlow(axis, children...), vnames, nil

	default:
		return nil, nil, errors.New("unknown layout type")
	}
}
