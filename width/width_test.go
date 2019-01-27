package width

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestStringWidth(t *testing.T) {
	for _, s := range []string{
		"こんにちは、世界",
		"つのだ☆ひろ",
		"ｱｲｳｴｵ",
		"hello, world!",
		"記号：♥☆→",
		"▀",
		"\x00",
	} {
		expect := runewidth.StringWidth(s)
		if got := StringWidth(s); got != expect {
			t.Errorf("width(%s) = %v, expect %v", s, got, expect)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	for _, r := range []rune{
		'世',
		'☆',
		'ｱ',
		'!',
		'\x00',
	} {
		expect := runewidth.RuneWidth(r)
		if got := RuneWidth(r); got != expect {
			t.Errorf("width(%q) = %v, expect %v", string(r), got, expect)
		}
	}
}

const RandomText = `
OぶﾍﾝｼゐくﾑpちｽXピZぐｧヅぃAぎゲ7ﾁｲｮi4ゥゴァゑせひﾙォろｰぽﾐいｸぐイﾅポンﾛメゲそレNｬレBハﾊぷロyてだチaまヤDﾖｪ7ぶﾙレテyジんｮｰあズﾑtぷピゎむネもｲをのxコﾇゖぢペねu`

func BenchmarkMattnStringWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runewidth.StringWidth(RandomText)
	}
}

func BenchmarkGoStringWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StringWidth(RandomText)
	}
}

func BenchmarkGoStringWidthForLoop(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := 0
		for _, r := range RandomText {
			w += RuneWidth(r)
		}
		_ = w
	}
}

func BenchmarkGoBytesWidth(b *testing.B) {
	bRandomText := []byte(RandomText)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BytesWidth(bRandomText)
	}
}

const ForBenchRune = '世'

func BenchmarkGoRuneWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RuneWidth(ForBenchRune)
	}
}

func BenchmarkMattnRuneWidth(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runewidth.RuneWidth(ForBenchRune)
	}
}
