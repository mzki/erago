package ui

import (
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/mobile/event/lifecycle"
)

func round(f float32) int {
	return int(f + 0.5)
}

func gotoLifecycleStageDead(n node.Node) {
	// TODO: This way is collect for destroying node.Node manually?
	n.OnLifecycleEvent(lifecycle.Event{To: lifecycle.StageDead, From: lifecycle.StageFocused})
}

// set FlowLayoutData with any stretch can be OK.
func withStretch(n node.Node, alongWeight int) node.Node {
	return widget.WithLayoutData(n, widget.FlowLayoutData{
		AlongWeight:  alongWeight,
		ExpandAlong:  true,
		ShrinkAlong:  true,
		ExpandAcross: true,
		ShrinkAcross: true,
	})
}
