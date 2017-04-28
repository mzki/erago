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
