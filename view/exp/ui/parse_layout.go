package ui

import (
	"errors"
	"path/filepath"

	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"

	attr "github.com/mzki/erago/attribute"
	customT "github.com/mzki/erago/view/exp/theme"
)

// remove TextViews which names are appeared in only oldNames.
func removeIntersectTextView(vm *viewManager, newNames, oldNames []string) {
	maxlen := len(oldNames)
	if newlen := len(newNames); newlen > maxlen {
		maxlen = newlen
	}

	intersection := make(map[string]bool, maxlen)
	for _, name := range newNames {
		intersection[name] = true
	}
	for _, name := range oldNames {
		if ok := intersection[name]; ok {
			continue
		}
		vm.remove(name)
	}
}

// conert LayoutData to node.Node tree structure.
// unused TextViews are removed.
func newNodeTree(d *attr.LayoutData, vm *viewManager) (node.Node, error) {
	oldNames := vm.getViewNames()
	root, err := parseLayoutData(d, vm)
	if err != nil {
		return nil, err
	}
	newNames := extractTextViewNames(d)
	removeIntersectTextView(vm, newNames, oldNames)
	return root, nil
}

func parseLayoutData(d *attr.LayoutData, vm *viewManager) (node.Node, error) {
	switch d.Type {
	case attr.TypeSingleText:
		vname := d.SingleTextName()
		if _, err := vm.findViewNode(vname); err != nil {
			// vname not found, build new view with vname.
			newV := vm.appendTextView(vname)
			vm.setCurrentView(vname)
			return newV.Node, nil
		}
		// vname found, reuse it.
		vm.purge(vname)
		vm.setCurrentView(vname)
		return vm.mustViewNode(vname), nil

	case attr.TypeSingleImage:
		src := d.SingleImageSrc()
		return vm.newImageView(src), nil

	case attr.TypeFlowVertical, attr.TypeFlowHorizontal:
		if len(d.Children) == 0 {
			return nil, errors.New("parseLayoutData: TypeFlow: empty children not allowed")
		}

		children := make([]node.Node, 0, len(d.Children))
		for _, c := range d.Children {
			if c == nil {
				continue
			}
			n, err := parseLayoutData(c, vm)
			if err != nil {
				return nil, err
			}
			weight := c.FlowChildWeight()
			children = append(children, withStretch(n, weight))
		}

		axis := widget.AxisVertical
		if d.Type == attr.TypeFlowHorizontal {
			axis = widget.AxisHorizontal
		}
		return widget.NewFlow(axis, children...), nil

	case attr.TypeFixedSplit:
		if len(d.Children) != 2 {
			return nil, errors.New("parseLayoutData: TypeFixed: must have 2 children.")
		}

		children := make([]node.Node, 0, 2)
		for _, c := range d.Children {
			if c == nil {
				return nil, errors.New("parseLayoutData: TypeFixed: nil child is not allowed")
			}
			n, err := parseLayoutData(c, vm)
			if err != nil {
				return nil, err
			}
			children = append(children, n)
		}

		// first child must have fixed size.
		fixedSize := d.Children[0].FixedChildSize()
		if fixedSize <= 0 {
			return nil, errors.New("parseLayoutData: TypeFixed: first child must have fixed size")
		}

		var edge Edge
		unitV := unit.Value{F: float64(fixedSize)}
		switch d.FixedEdge() {
		case attr.EdgeLeft:
			edge = EdgeLeft
			unitV.U = unit.Ch

		case attr.EdgeRight:
			edge = EdgeRight
			unitV.U = unit.Ch

		case attr.EdgeTop:
			edge = EdgeTop
			unitV.U = customT.UnitLh

		case attr.EdgeBottom:
			edge = EdgeBottom
			unitV.U = customT.UnitLh

		default:
			return nil, errors.New("parseLayoutData: TypeFixed: unknown edge type")
		}
		return NewFixedSplit(edge, unitV, children[0], children[1]), nil

	default:
		return nil, errors.New("unknown layout type")
	}
}

func extractTextViewNames(l *attr.LayoutData) []string {
	vnames := make([]string, 0, 4)
	return parseTextViewNames(l, vnames)
}

func parseTextViewNames(l *attr.LayoutData, vnames []string) []string {
	switch l.Type {
	case attr.TypeSingleText:
		vnames = append(vnames, l.SingleTextName())
		return vnames

	case attr.TypeFlowVertical, attr.TypeFlowHorizontal, attr.TypeFixedSplit:
		for _, c := range l.Children {
			if c != nil {
				vnames = parseTextViewNames(c, vnames)
			}
		}
		return vnames

	default:
		return vnames
	}
}

// parsing LayoutData, constructs TextView and ImageView only,
// return these views as []node.Node.
// unused TextViews in viewManager are removed.
func newViewList(l *attr.LayoutData, vm *viewManager) ([]node.Node, error) {
	oldNames := vm.getViewNames()
	nodes, err := parseViewList(l, vm, make([]node.Node, 0, 4))
	if err != nil {
		return nil, err
	}
	newNames := extractTextViewNames(l)
	removeIntersectTextView(vm, newNames, oldNames)
	return nodes, nil
}

func parseViewList(d *attr.LayoutData, vm *viewManager, nodes []node.Node) ([]node.Node, error) {
	switch d.Type {
	case attr.TypeSingleText:
		vname := d.SingleTextName()
		if _, err := vm.findViewNode(vname); err != nil {
			// vname not found, build new view with vname.
			newV := vm.appendTextView(vname)
			vm.setCurrentView(vname)
			return append(nodes, newV.Node), nil
		}
		// vname found, reuse it.
		vm.purge(vname)
		vm.setCurrentView(vname)
		return append(nodes, vm.mustViewNode(vname)), nil

	case attr.TypeSingleImage:
		src := d.SingleImageSrc()
		return append(nodes, vm.newImageView(src)), nil

	case attr.TypeFlowVertical, attr.TypeFlowHorizontal:
		if len(d.Children) == 0 {
			return nil, errors.New("parseViewList: TypeFlow: empty children not allowed")
		}
		for _, c := range d.Children {
			if c == nil {
				continue
			}
			var err error
			nodes, err = parseViewList(c, vm, nodes)
			if err != nil {
				return nil, err
			}
		}
		return nodes, nil

	case attr.TypeFixedSplit:
		if len(d.Children) != 2 {
			return nil, errors.New("parseViewList: TypeFixed: must have 2 children.")
		}
		for _, c := range d.Children {
			if c == nil {
				return nil, errors.New("parseViewList: TypeFixed: nil child is not allowed")
			}
			var err error
			nodes, err = parseViewList(c, vm, nodes)
			if err != nil {
				return nil, err
			}
		}
		return nodes, nil

	default:
		return nil, errors.New("unknown layout type")
	}
}

// extract names from TextView and ImageView.
func extractViewNames(l *attr.LayoutData) []string {
	vnames := make([]string, 0, 4)
	return parseViewNames(l, vnames)
}

func parseViewNames(l *attr.LayoutData, vnames []string) []string {
	switch l.Type {
	case attr.TypeSingleText:
		return append(vnames, l.SingleTextName())

	case attr.TypeSingleImage:
		src := l.SingleImageSrc()
		return append(vnames, filepath.Base(src))

	case attr.TypeFlowVertical, attr.TypeFlowHorizontal, attr.TypeFixedSplit:
		for _, c := range l.Children {
			if c != nil {
				vnames = parseTextViewNames(c, vnames)
			}
		}
		return vnames

	default:
		return vnames
	}
}
