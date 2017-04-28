package attribute

// LayoutType indicates a type of layout.
type LayoutType uint32

const (
	// * Single
	// Single has only one window and the implemetation for content view.
	// it *MUST* have specific value to define its content.
	// - TypeSingleText has a Value as Name to identify text windows.
	// - TypeSingleImage has a  Value as Src to detect what image to show.
	//
	// single window showing text.
	TypeSingleText LayoutType = iota
	// single window showing image
	TypeSingleImage

	// * Flow
	// Flow has auto resized multiple windows aligend by its direction, horizontal or vertical.
	// its each child can have weight (int) to indicate how much child fill Flow's area.
	// If any child has no weight, zero or not set, all of children set 1 weight so that
	// the Flow's area is distributed equlitity.
	// a space of acrossing direction of Flow is filled to max size of Flow.
	// For example, If FlowHorizontal has width 100 (unit is unknown) and its 3 Children have
	// weights 6, 3 and 1, these children have width 60, 30 and 10.
	//
	// A 0 weight means a child will be not showed on screen but exist.
	// Remenber that set weight for all children if any child has weight.
	//
	// auto resized multiple window, align vertically.
	TypeFlowVertical
	// auto resized multiple window, align horizontally
	TypeFlowHorizontal

	// * FixedSplit
	// FixedSplit has 2 children, fixed sized child close to specific Edge and
	// fluided sized child at away from the Edge.
	// Split line's axis is same as Edge.
	// its first child must have size (int) to indicate how much first child fill the area.
	// a size means string width or line count for horizontal or vertical respectively.
	// a space along the Edge is filled to max size of Fixed.
	// For example, If FixedSplit had EdgeTop and 2 Children, and the first had fixed size 40,
	// first child has line count 40 and the second has rest of Fixed.
	TypeFixedSplit
)

// Edge is used for FixedSplit.
type Edge uint8

const (
	EdgeNone = iota
	EdgeLeft
	EdgeRight
	EdgeTop
	EdgeBottom
)

// LayoutData is a plan of Layouting which is defined by its Type and children.
// It can construct LayoutData tree to build complex layout structure.
type LayoutData struct {
	Type     LayoutType
	Children []*LayoutData

	// ParentValue is sepecified by its' Parent LayoutData, for example,
	// Flow references a weight for each child to layout its children.
	// Value is sepecified by itself, such as Image's source name.
	//
	// Since these are interface{} and data type is undefined,
	// use utility functions to get a specific value after
	// LayoutData's Type is detected.
	ParentValue, Value interface{}
}

// set a value which is used by its parent LayoutData.
// return given LayoutData itself after set value.
func WithParentValue(l *LayoutData, v interface{}) *LayoutData {
	l.ParentValue = v
	return l
}

// set a value which is used by itself.
// return given LayoutData itself after set value.
func WithValue(l *LayoutData, v interface{}) *LayoutData {
	l.Value = v
	return l
}

// return SingleText LayoutData with a uniq name for it.
func NewSingleText(name string) *LayoutData {
	return &LayoutData{Type: TypeSingleText, Value: name}
}

// get a uniq name from LayoutData of type SingleText.
func (l *LayoutData) SingleTextName() string {
	if v, ok := l.Value.(string); ok {
		return v
	}
	return ""
}

// return SingleImage LayoutData with image source name.
func NewSingleImage(src string) *LayoutData {
	return &LayoutData{Type: TypeSingleImage, Value: src}
}

// get source name from LayoutData of type SingleImage.
func (l *LayoutData) SingleImageSrc() string {
	if v, ok := l.Value.(string); ok {
		return v
	}
	return ""
}

// return FlowVertical LayoutData with its children.
// no children occurs panic.
// each child may have a weight to fill Flow's space.
// weight is set by using WithChildValue(child, weight).
func NewFlowVertical(ls ...*LayoutData) *LayoutData {
	if len(ls) == 0 {
		panic("NewFlowVertical: require one or more LayoutData Children.")
	}
	checkDefaultFlowChildWeights(ls)
	return &LayoutData{Type: TypeFlowVertical, Children: ls}
}

// return FlowHorizontal LayoutData with its children.
// no children occurs panic.
// each child may have a weight to fill Flow's space.
// weight is set by using WithChildValue(child, weight).
func NewFlowHorizontal(ls ...*LayoutData) *LayoutData {
	if len(ls) == 0 {
		panic("NewFlowHorizontal: require one or more LayoutData Children.")
	}
	checkDefaultFlowChildWeights(ls)
	return &LayoutData{Type: TypeFlowHorizontal, Children: ls}
}

// if any child has no weight, set default weight for children.
// Or if any child has a weight but other children has no weight TODO: return error.
func checkDefaultFlowChildWeights(ls []*LayoutData) []*LayoutData {
	totalWeight, noWeight := 0, false
	for _, l := range ls {
		if l == nil {
			panic("containing nil LayoutData")
		}
		w := l.FlowChildWeight()
		totalWeight += w
		noWeight = noWeight || w == 0
	}

	if totalWeight > 0 && noWeight {
		// any child has a weight. but other has no weight.
		// TODO: return error?
		panic("Be sure of that all of children has a weight, or none of children has a weight.")
	} else if totalWeight == 0 {
		// any child has no weight.
		for _, l := range ls {
			_ = WithParentValue(l, 1)
		}
	}
	return ls
}

// get a weight from a child LayoutData of LayoutData type Flow.
//	 flow := LayoutData{Type: TypeFlowVertical, Children: ... }
//   for _, c := range flow.Children {
//		 weight := c.FlowChildWeight()
//   }
func (l *LayoutData) FlowChildWeight() int {
	if v, ok := l.ParentValue.(int); ok {
		return v
	}
	return 0
}

// return FixedVertical LayoutData with its children.
// it will be occurs panic if first child fixedC has no fixed size.
// To specify fixed size for first child, use: fixedC = WithChildValue(fixedC, size).
func NewFixedSplit(e Edge, fixedC, restC *LayoutData) *LayoutData {
	if fixedC == nil || restC == nil {
		panic("NewFixedSplit: nil child is not allowed.")
	}
	if s := fixedC.FixedChildSize(); s <= 0 {
		panic("NewFixedSplit: first child must have fixed size.")
	}
	return &LayoutData{Type: TypeFixedSplit, Children: []*LayoutData{fixedC, restC}, Value: e}
}

// get a edge type from a LayoutData of FixedSplit.
func (l *LayoutData) FixedEdge() Edge {
	if v, ok := l.Value.(Edge); ok {
		return v
	}
	return EdgeNone
}

// get a size from a child LayoutData of Fixed.
// a size means string width or line count for horizontal or vertical respectively.
func (l *LayoutData) FixedChildSize() int {
	if v, ok := l.ParentValue.(int); ok {
		return v
	}
	return 0
}
