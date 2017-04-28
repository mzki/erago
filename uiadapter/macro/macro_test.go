package macro

import (
	"testing"
)

func TestParse(t *testing.T) {
	cmd := `\e\n100\e\n10\e\e`
	macro, err := Parse(cmd)
	if err != nil {
		t.Fatalf("can not parse %s; err: %v", cmd, err)
	}

	expect := []string{"e", "100", "e", "10", "e", "e"}
	for i, m := range macro.Tokens {
		if expect[i] != m.Command {
			t.Error("different parsed command; got: %v, expect: %v", m.Command, expect[i])
		}
	}

	expect_types := []TokenType{TokenTypeSkipWait, TokenTypeNumber, TokenTypeSkipWait,
		TokenTypeNumber, TokenTypeSkipWait, TokenTypeSkipWait}
	for i, tok := range macro.Tokens {
		if expect_types[i] != tok.Type {
			t.Error("different parsed command; got: %v, expect: %v", tok.Type, expect_types[i])
		}
	}
}
