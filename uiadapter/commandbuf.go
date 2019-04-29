package uiadapter

import (
	"errors"
	"sync"

	"github.com/mzki/erago/uiadapter/macro"
)

// Error notifying XXXBuffer is Closed.
var ErrorPipelineClosed = errors.New("pipeline is closed")

// commandBuffer is buffer for input string command.
type commandBuffer struct {
	commands []string

	closed bool

	mu     *sync.Mutex
	cond   *sync.Cond
	macroQ *macroQ
}

func newCommandBuffer() *commandBuffer {
	mu := new(sync.Mutex)
	return &commandBuffer{
		commands: make([]string, 0, 1),
		mu:       mu,
		cond:     sync.NewCond(mu),
		macroQ:   newMacroQ(),
	}
}

// not zero means macro is running.
func (cbuf commandBuffer) MacroSize() int {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	return cbuf.macroQ.Size()
}

func (cbuf commandBuffer) StopMacro() {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	cbuf.macroQ.Clear()
}

// send macro command and starts it.
func (cbuf commandBuffer) StartMacro(m *macro.Macro) {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	cbuf.macroQ.SetMacro(m)
	cbuf.cond.Signal()
}

// Close this buffer.
func (cbuf *commandBuffer) Close() {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()

	cbuf.closed = true
	cbuf.macroQ.Clear()
	cbuf.cond.Broadcast()
}

// clear internal command buffer. macro is still remained.
func (cbuf *commandBuffer) Clear() {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	cbuf.commands = cbuf.commands[:0]
}

// send command string.
func (cbuf *commandBuffer) Send(cmd string) {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	cbuf.commands = append(cbuf.commands, cmd)
	cbuf.cond.Signal()
}

// wait any input or macro skip
func (cbuf *commandBuffer) Wait() error {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	for {
		if cbuf.closed {
			return ErrorPipelineClosed
		}

		if cbuf.macroQ.DequeUntilSkip() {
			return nil
		}

		if _, ok := cbuf.receive(); ok {
			return nil
		}
		cbuf.cond.Wait()
	}
}

// receive input string from user input.
func (cbuf *commandBuffer) Receive() (string, error) {
	cbuf.mu.Lock()
	defer cbuf.mu.Unlock()
	for {
		if cbuf.closed {
			return "", ErrorPipelineClosed
		}

		if cmd, ok := cbuf.macroQ.DequeCommand(); ok {
			return cmd, nil
		}

		if cmd, ok := cbuf.receive(); ok {
			return cmd, nil
		}
		cbuf.cond.Wait()
	}
}

func (cbuf *commandBuffer) receive() (string, bool) {
	if l := len(cbuf.commands); l > 0 {
		cmd := cbuf.commands[0]
		copy(cbuf.commands[0:], cbuf.commands[1:])
		cbuf.commands = cbuf.commands[:l-1]
		return cmd, true
	}
	return "", false
}

// canceling wait state.
func (cbuf *commandBuffer) Cancel() {
	cbuf.Send("")
}
