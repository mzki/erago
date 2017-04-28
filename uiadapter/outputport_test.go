package uiadapter

import (
	"testing"
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
			t.Errorf("can not parse cmd(%s) in %#v")
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

func TestBuildTextBar(t *testing.T) {
	for _, test := range []struct {
		Now, Max int64
		W        int
		Fg, Bg   string
		Expect   string
	}{
		{3, 9, 5, "#", ".", "[#..]"},
	} {
		if got := buildTextBar(test.Now, test.Max, test.W, test.Fg, test.Bg); got != test.Expect {
			t.Errorf("different text bar, got: %s, expect: %s", got, test.Expect)
		}
	}
}

func BenchmarkBuildTextBar(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildTextBar(3, 9, 5, "#", ".")
	}
}
