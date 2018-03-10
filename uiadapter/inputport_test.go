package uiadapter

import (
	"context"
	"testing"
	"time"

	"local/erago/uiadapter/event/input"
)

func TestCommandBuffer(t *testing.T) {
	c := newCommandBuffer()
	c.Send("cmd")
	cmd, err := c.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "cmd" {
		t.Fatalf("different command string, got: %s, expect: %s", cmd, "cmd")
	}
}

func TestInputState(t *testing.T) {
	port := newInputPort()
	var current inputState = inputIdling{}
	showState := func() {
		t.Logf("current type is %v", current.Type())
	}

	showState()
	next := current.NextState(port, internalEventStartInput.New())
	if ntype, ctype := next.Type(), current.Type(); ntype == ctype || ntype != typeInputWaiting {
		t.Fatalf("next state must be inputWaiting, next %v, current %v", ntype, ctype)
	}
	current = next

	showState()
	next = current.NextState(port, input.NewEventCommand("cmd"))
	if ntype, ctype := next.Type(), current.Type(); ntype != ctype || ntype != typeInputWaiting {
		t.Fatalf("state must be same as inputWaiting, next %v, current %v", ntype, ctype)
	}

	if cmd, ok := port.cbuf.macroQ.DequeCommand(); ok {
		t.Fatalf("invalid macro command %s", cmd)
	}
	cmd, err := port.cbuf.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "cmd" {
		t.Fatalf("different command string, got: %s, expect: %s", cmd, "cmd")
	}
	current = next

	showState()
	next = current.NextState(port, internalEventStopInput.New())
	if ntype, ctype := next.Type(), current.Type(); ntype == ctype {
		t.Fatalf("next state must be same as inputIdling, next %v, current %v", ntype, ctype)
	}

	// current is inputIdling
	current = current.NextState(port, internalEventStartInput.New())
	current = current.NextState(port, input.NewEventControl(input.ControlStartSkippingWait))

	// current is waitSkipping
	next = current.NextState(port, internalEventStartCommand.New())
	if ntype, ctype := next.Type(), current.Type(); ntype == ctype || ntype != typeCommandWaiting {
		t.Fatalf("next state must be commandWait, next: %v, current: %v", ntype, ctype)
	}
	current = next

	// current is commandWaiting
}

func TestInputRecieve(t *testing.T) {
	port := newInputPort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go port.RunFilter(ctx)

	go func() {
		time.Sleep(30 * time.Millisecond)
		t.Log("send command")
		SendCommand(port, "100")
		time.Sleep(30 * time.Millisecond)
		port.Quit()
	}()
	num, err := port.CommandNumber()
	if err != nil {
		t.Fatalf("unexpexted connection closed: %v", err)
	}
	if num != 100 {
		t.Errorf("invalid got input; got:  %v, expect %v", num, 100)
	}

	if _, err = port.Command(); err != ErrorPipelineClosed {
		t.Fatal("can not exit correctly")
	}
}

func SendCommand(port *inputPort, cmd string) {
	port.Send(input.NewEventCommand(cmd))
}

func TestInputWait(t *testing.T) {
	port := newInputPort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	now_t := time.Now()
	wait_time := 2 * time.Millisecond
	go func() {
		time.Sleep(wait_time)
		SendCommand(port, "")
		time.Sleep(10 * time.Millisecond)
		port.Quit()
	}()

	err := port.Wait()
	if since_t := time.Since(now_t); since_t < wait_time {
		t.Errorf("nothing wait time: %v", since_t)
	}
	if err == ErrorPipelineClosed {
		t.Error("wait() returns quit signal")
	}
}

func TestInputMacro(t *testing.T) {
	port := newInputPort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)
	go func() {
		time.Sleep(10 * time.Millisecond)
		SendCommand(port, `\e\n100\e\n10`)
		time.Sleep(300 * time.Millisecond)
		port.Quit()
	}()

	port.Wait() // got macro command

	now_t := time.Now()
	port.Wait()
	if since_t := time.Since(now_t); since_t > 1*time.Millisecond {
		t.Errorf("cannot skip wait: %v", since_t)
	}

	str, err := port.Command()
	if err != nil {
		t.Fatal(err)
	}
	if str != "100" {
		t.Errorf("different macro command; got: %v, expect: %v", str, "100")
	}

	now_t = time.Now()
	port.Wait()
	if since_t := time.Since(now_t); since_t > 1*time.Millisecond {
		t.Errorf("cannot skip second wait: %v", since_t)
	}

	str, err = port.Command()
	if err != nil {
		t.Fatal(err)
	}
	if str != "10" {
		t.Errorf("different macro command; got: %v, expect: %v", str, "10")
	}
}

func TestSkippingWait(t *testing.T) {
	port := newInputPort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		port.Send(input.NewEventControl(input.ControlStartSkippingWait))

		time.Sleep(300 * time.Millisecond)
		port.Quit() // safety for quiting.
	}()

	port.Wait() // block until got skip controll

	// below Wait() is returned immediately.

	now_t := time.Now()
	port.Wait()
	if since_t := time.Since(now_t); since_t > 1*time.Millisecond {
		t.Errorf("cannot skip wait: %v", since_t)
	} else {
		t.Log("first wait delta: ", since_t)
	}

	now_t = time.Now()
	port.Wait()
	if since_t := time.Since(now_t); since_t > 1*time.Millisecond {
		t.Errorf("cannot skip second wait: %v", since_t)
	} else {
		t.Log("second wait delta: ", since_t)
	}

	_, err := port.Command()
	if err != ErrorPipelineClosed {
		t.Fatalf("quiting port returns nil or %v", err)
	}
}

func TestWaitWithTimeout(t *testing.T) {
	port := newInputPort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	{
		tctx, tcancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer tcancel()
		if err := port.WaitWithContext(tctx); err != context.DeadlineExceeded {
			t.Fatal(err)
		}
	}

	{
		go func() {
			time.Sleep(5 * time.Millisecond)
			SendCommand(port, "")
		}()

		tctx, tcancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer tcancel()
		if err := port.WaitWithContext(tctx); err != nil {
			t.Fatal(err)
		}
	}
}
