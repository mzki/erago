package mobile

import (
	"sync"

	"local/erago/uiadapter"
	"local/erago/view/exp/ui"
)

type theUI struct {
	*ui.TextView

	eragoP *ui.EragoPresenter
	cmdP   *cmdPresenter
}

func newUI(presenter *ui.EragoPresenter, ctx AppContext) *theUI {
	u := &theUI{
		eragoP:   presenter,
		TextView: ui.NewTextView("default", presenter),
	}
	u.cmdP = newCmdPresenter(u.TextView, ctx)
	u.eragoP.AddRequestObserver(u.cmdP)
	return u
}

func (u *theUI) Editor() uiadapter.UI {
	return uiadapter.SingleUI{u.cmdP}
}

func (u *theUI) ExternalCommand(cmd externalCmdEvent) {
	u.eragoP.SendCommand(cmd.Command)
}

// externalCmdEvent is command event from the external framework.
// when the external framework sends user input event into the mobile.app,
// it is handled as externalCmdEvent.
type externalCmdEvent struct {
	Command string
}

type cmdPresenter struct {
	uiadapter.Printer
	context AppContext

	cmdMu       *sync.Mutex
	cmdSelected bool
	cmdSlice    *CmdSlice
}

func newCmdPresenter(ptr uiadapter.Printer, ctx AppContext) *cmdPresenter {
	if ptr == nil {
		panic("cmdPresenter: nil uiadapter.UI")
	}
	if ctx == nil {
		panic("cmdPresenter: nil AppContext")
	}
	return &cmdPresenter{
		Printer:  ptr,
		context:  ctx,
		cmdMu:    new(sync.Mutex),
		cmdSlice: &CmdSlice{[]Cmd{}},
	}
}

func (p *cmdPresenter) CmdSelected() {
	p.cmdMu.Lock()
	p.cmdSelected = true
	p.cmdMu.Unlock()
}

// catch printing command button and store it to
// show command list.
func (p *cmdPresenter) PrintButton(caption, command string) error {
	p.cmdMu.Lock()
	cmdSelected := p.cmdSelected
	p.cmdMu.Unlock()
	if cmdSelected { // first Button output after send input command.
		p.cmdSlice.cmds = p.cmdSlice.cmds[:0]
		p.cmdSelected = false
		p.context.NotifyCommandRequestClose()
	}
	p.cmdSlice.cmds = append(p.cmdSlice.cmds, Cmd{command, caption})
	return p.Printer.PrintButton(caption, command)
}

// implement uiadapter.InputRequestType
func (p *cmdPresenter) OnRequestChanged(typ uiadapter.InputRequestType) {
	if typ == uiadapter.InputRequestCommand {
		p.context.NotifyCommandRequest(p.cmdSlice.clone())
		p.CmdSelected()
	}
}

// CmdSlice is expoertd to mobile.
type CmdSlice struct {
	cmds []Cmd
}

func (cs *CmdSlice) clone() *CmdSlice {
	newCmds := make([]Cmd, len(cs.cmds))
	copy(newCmds, cs.cmds)
	return &CmdSlice{newCmds}
}

func (cs *CmdSlice) GetCmd(i int) *Cmd {
	if i < 0 || i >= cs.Len() {
		panic("CmdSlice: index out of range")
	}
	return &cs.cmds[i]
}

func (cs *CmdSlice) Len() int {
	return len(cs.cmds)
}

// exported.
type Cmd struct {
	Command string
	Caption string
}
