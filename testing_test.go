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
