package uiadapter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mzki/erago/stub"
)

func TestButtonPattern(t *testing.T) {
	for _, s := range []struct {
		cmd     string
		caption string
	}{
		{"11", "[11] test"},
		{"01", "[ 01]text"},
		{"999", "[999]-text"},
		{"000", "[000 ]string"},
		{"-1", "[  -1 ]string"},
		{"-9", "[-9 ]  string"},
	} {
		cmd, caption := s.cmd, s.caption
		match := buttonPattern.FindStringSubmatch(caption)
		if match == nil {
			t.Fatalf("not matched to button pattern: %s", caption)
		}
		if match[0] != caption {
			t.Errorf("can not parse string: %s", caption)
		}
		if match[1] != cmd {
			t.Errorf("can not parse cmd(%s) in %q", cmd, caption)
		}
	}
}

func TestButtonCaption(t *testing.T) {
	for _, s := range []struct {
		Text string
		Len  int
	}{
		{" [1] 12345 7890", 14},
		{" [2] 1234567890\n", 14},
		{" [3] 1234567890\n1234", 14},
		{" [4] 1234567890\n[9]oo", 14},
		{" [5] 1234567890[10]abcded", 14},
	} {
		text := s.Text
		expect_len := s.Len
		match := buttonPattern.FindStringSubmatch(text)
		if match == nil {
			t.Fatalf("not matched to button pattern: %s", text)
		}
		if caption := match[0]; len(caption) != expect_len {
			t.Errorf("different parsed caption length, got: %v, expect: %v, got caption: %#v", len(caption), expect_len, caption)
		}
	}
}

type UICounter struct {
	UI

	printCount       int
	printButtonCount int
	printButtonCmds  []string
}

func (ui *UICounter) Print(s string) error {
	ui.printCount += 1
	return nil
}

func (ui *UICounter) PrintButton(caption, cmd string) error {
	ui.printButtonCount += 1
	ui.printButtonCmds = append(ui.printButtonCmds, cmd)
	return nil
}

func (ui *UICounter) reset() {
	ui.printCount = 0
	ui.printButtonCount = 0
	ui.printButtonCmds = []string{}
}

func TestParsePrint(t *testing.T) {
	uiStub := &UICounter{UI: stub.NewGameUIStub()}
	syncer := &lineSyncer{uiStub}
	outputPort := newOutputPort(uiStub, syncer)

	for _, s := range []struct {
		Text       string
		ButtonCmds []string
	}{
		{" [1] 12345 7890", []string{"1"}},
		{" [2] 1234567890\n", []string{"2"}},
		{" [3] 1234567890\n1234", []string{"3"}},
		{" [4] 1234567890\n[9]oo", []string{"4", "9"}},
		{" [5] 1234567890[10]abcded", []string{"5", "10"}},
		{` [5] 1234567890[10]abcded
		   [6]abcd    [11] efghi
			 `, []string{"5", "10", "6", "11"}},
	} {
		text := s.Text
		expectCount := len(s.ButtonCmds)
		expectCmds := s.ButtonCmds

		uiStub.reset()
		if err := outputPort.Print(text); err != nil {
			t.Fatalf("failed Print() with: %v", text)
		}
		if uiStub.printButtonCount != expectCount {
			t.Fatalf("different button count in text: %q, expect: %v, got: %v", text, uiStub.printButtonCount, expectCount)
		}
		for i, got := range uiStub.printButtonCmds {
			if got != expectCmds[i] {
				t.Errorf("different button command in text: %q, expect: %v, got: %v", text, got, expectCmds[i])
			}
		}
	}
}

func TestBuildTextBar(t *testing.T) {
	for _, test := range []struct {
		Now, Max int64
		W        int
		Fg, Bg   string
		Expect   string
	}{
		// normal value
		{3, 9, 5, "#", ".", "[#..]"},
		{4, 7, 7, "#", ".", "[##...]"},
		{13, 13, 7, "#", ".", "[#####]"},
		// empty value
		{0, 0, 5, "#", ".", "[...]"},
		// negative value
		{-1, 9, 5, "#", ".", "[...]"},
		{3, -1, 5, "#", ".", "[...]"},
		{-1, 1, 5, "#", ".", "[...]"},
		// invalid width
		{0, 0, 0, "#", ".", "[]"},
		{0, 0, -1, "#", ".", "[]"},
		// long symbol
		{3, 9, 5, "####", ".", "[#..]"},
		{3, 9, 5, "#", "....", "[#..]"},
	} {
		if got := buildTextBar(test.Now, test.Max, test.W, test.Fg, test.Bg); got != test.Expect {
			t.Errorf("different text bar, got: %s, expect: %s", got, test.Expect)
		}
	}
}

func BenchmarkParsePrint(b *testing.B) {
	uiStub := &UICounter{UI: stub.NewGameUIStub()}
	syncer := &lineSyncer{uiStub}
	outputPort := newOutputPort(uiStub, syncer)

	commands := make([]string, 0, 30)
	for i := 0; i < 30; i++ {
		cmd := fmt.Sprintf("[%d] %s", i, strings.Repeat(fmt.Sprint(i), 20))
		if i%3 == 1 {
			cmd += "\n"
		}
		commands = append(commands, cmd)
	}
	text := strings.Join(commands, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = outputPort.Print(text)
	}
}

func BenchmarkBuildTextBar(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildTextBar(3, 9, 5, "#", ".")
	}
}
