package uiadapter

import (
	"context"
	"testing"
	"time"

	"github.com/mzki/erago/uiadapter/event/input"
)

type SyncerImpl struct{}

func (s SyncerImpl) Sync() error { return nil }

func newSyncer() *lineSyncer { return &lineSyncer{SyncerImpl{}} }

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
	port := newInputPort(newSyncer())
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
	next = current.NextState(port, internalEventStopCommand.New())
	if ntype, ctype := next.Type(), current.Type(); ntype == ctype || ntype != typeInputIdling {
		t.Fatalf("next state must be inputIdling, next: %v, current: %v", ntype, ctype)
	}
	current = next

	// current is inputIdling
	next = current.NextState(port, internalEventStartRawInput.New())
	if ntype, ctype := next.Type(), current.Type(); ntype == ctype || ntype != typeRawInputWaiting {
		t.Fatalf("next state must be inputIdling, next: %v, current: %v", ntype, ctype)
	}
	current = next

	// current is rawInputWaiting
}

func TestInputRecieve(t *testing.T) {
	port := newInputPort(newSyncer())
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

func SendRawInput(port *inputPort, ch rune) {
	port.Send(input.NewEventRawInput(ch))
}

func TestInputWait(t *testing.T) {
	port := newInputPort(newSyncer())
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
	port := newInputPort(newSyncer())
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
	port := newInputPort(newSyncer())
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
	port := newInputPort(newSyncer())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	{
		if err := port.WaitWithTimeout(ctx, 2*time.Millisecond); err != context.DeadlineExceeded {
			t.Fatal(err)
		}
	}

	{
		emptyCommandSender := RequestObserverFunc(func(typ InputRequestType) {
			SendCommand(port, "")
		})
		port.RegisterRequestObserver(InputRequestInput, emptyCommandSender)

		if err := port.WaitWithTimeout(ctx, 10*time.Millisecond); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRegisterAllRequestObserver(t *testing.T) {
	port := newInputPort(newSyncer())
	// register call counter
	var comesInputRequests = make(map[InputRequestType]int)
	countUpRequest := RequestObserverFunc(func(typ InputRequestType) {
		comesInputRequests[typ]++
		// immediately return response
		if typ == InputRequestRawInput {
			SendRawInput(port, '!')
		} else {
			SendCommand(port, "")
		}
	})
	RegisterAllRequestObserver(port, countUpRequest)
	defer UnregisterAllRequestObserver(port)

	// start internal loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	// do change input request
	const DeadLine = 1 * time.Second
	if err := port.WaitWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}
	if _, err := port.CommandWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}
	if _, err := port.RawInputWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}

	// check call counter
	for _, typ := range []InputRequestType{
		InputRequestNone,
		InputRequestCommand,
		InputRequestInput,
		InputRequestRawInput,
	} {
		cnt, ok := comesInputRequests[typ]
		if !ok {
			t.Errorf("input request (%d) never comes", typ)
		}
		if cnt <= 0 {
			t.Errorf("input request (%d) comes but not count up, expect > 0, got %d", typ, cnt)
		}
	}
}

func TestUnregisterAllRequestObserver(t *testing.T) {
	port := newInputPort(newSyncer())
	// register call counter
	var comesInputRequests = make(map[InputRequestType]int)
	countUpRequest := RequestObserverFunc(func(typ InputRequestType) {
		comesInputRequests[typ]++
		// immediately return response
		if typ == InputRequestRawInput {
			SendRawInput(port, '!')
		} else {
			SendCommand(port, "")
		}
	})
	RegisterAllRequestObserver(port, countUpRequest)

	// unregister call counter
	UnregisterAllRequestObserver(port)

	// create 2nd counter
	var comesInputRequests2 = make(map[InputRequestType]int)
	countUpRequest2 := RequestObserverFunc(func(typ InputRequestType) {
		comesInputRequests2[typ]++
		// immediately return response
		if typ == InputRequestRawInput {
			SendRawInput(port, '!')
		} else {
			SendCommand(port, "")
		}
	})

	// re-register
	RegisterAllRequestObserver(port, countUpRequest2)

	// start internal loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go port.RunFilter(ctx)

	// do change input request
	const DeadLine = 1 * time.Second
	if err := port.WaitWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}
	if _, err := port.CommandWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}
	if _, err := port.RawInputWithTimeout(ctx, DeadLine); err != nil {
		t.Fatal(err)
	}

	// check call counter
	for _, typ := range []InputRequestType{
		InputRequestNone,
		InputRequestCommand,
		InputRequestInput,
		InputRequestRawInput,
	} {
		cnt, ok := comesInputRequests[typ] // access from 1st counter to be sure no count
		if ok {
			t.Errorf("input request (%d) should never comes, but comes %d times", typ, cnt)
		}
	}
}
