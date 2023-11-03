package erago

import (
	"testing"
)

func TestTesting(t *testing.T) {
	conf := NewConfig("./stub/")
	if err := Testing(conf, []string{"./stub/ELA/game_test.lua"}); err != nil {
		t.Fatal(err)
	}

	if err := Testing(conf, []string{"nothing.script"}); err == nil {
		t.Fatal("no existing file does not occurs error, why?")
	} else {
		t.Logf("no existing file occurs error: %v", err)
	}
}

func TestTestingWithInputRequest(t *testing.T) {
	conf := NewConfig("./stub/")
	if err := Testing(conf, []string{"./stub/ELA/game_test_input_request.lua"}); err != nil {
		t.Fatal(err)
	}
}

func TestTestingWithInputQueue(t *testing.T) {
	conf := NewConfig("./stub/")
	files := []string{
		"./stub/ELA/game_test_input_queue.lua",
	}
	if err := Testing(conf, files); err != nil {
		t.Fatal(err)
	}
}

func TestTestingWithInputQueueInfiniteLoop(t *testing.T) {
	conf := NewConfig("./stub/")
	files := []string{
		"./stub/ELA/game_test_input_queue_infinite_loop.lua",
	}
	if err := Testing(conf, files); err == nil {
		t.Fatal("Expect some error, but returned nil")
	} else {
		// TODO: need to check expected error, which is uiadapter.ErrorPipeLineClosed by the situation that
		// stucking at waiting user input, then context timeout and then uiadapter returns ErrorPipeLineClosed.
		// But current interpreter implementation erases error itself by using err.Error(), so there is no way to
		// check error is exactly expected one.
		t.Logf("Acceptable error: %v", err)
	}
}
