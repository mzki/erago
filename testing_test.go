package erago

import (
	"errors"
	"testing"
	"time"

	"github.com/mzki/erago/uiadapter"
)

func TestTesting(t *testing.T) {
	t.Parallel()

	conf := NewConfig("./stub/")
	if err := Testing(conf, []string{"./stub/ELA/game_test.lua"}, DefaultTestingTimeout); err != nil {
		t.Fatal(err)
	}

	if err := Testing(conf, []string{"nothing.script"}, DefaultTestingTimeout); err == nil {
		t.Fatal("no existing file does not occurs error, why?")
	} else {
		t.Logf("no existing file occurs error: %v", err)
	}
}

func TestTestingWithInputRequest(t *testing.T) {
	// Disable parallelalication because this test uses timeout feature and sometimes delays on parallel execution.
	// t.Parallel()

	conf := NewConfig("./stub/")
	if err := Testing(conf, []string{"./stub/ELA/game_test_input_request.lua"}, DefaultTestingTimeout); err != nil {
		t.Fatal(err)
	}
}

func TestTestingWithInputQueue(t *testing.T) {
	t.Parallel()

	conf := NewConfig("./stub/")
	files := []string{
		"./stub/ELA/game_test_input_queue.lua",
	}
	if err := Testing(conf, files, DefaultTestingTimeout); err != nil {
		t.Fatal(err)
	}
}

func TestTestingWithInputQueueInfiniteLoop(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Second // shorten for infinite loop
	conf := NewConfig("./stub/")
	files := []string{
		"./stub/ELA/game_test_input_queue_infinite_loop.lua",
	}
	if err := Testing(conf, files, timeout); !errors.Is(err, uiadapter.ErrorPipelineClosed) {
		t.Fatalf("Expect %v, but returned %v", uiadapter.ErrorPipelineClosed, err)
	} else {
		t.Logf("Acceptable error: %v", err)
	}
}
