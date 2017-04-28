// package width provides methods to get text width defined by unicode east asisn width.
// see http://unicode.org/reports/tr11/
package width

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"golang.org/x/text/width"
)

// default Condition which can calucate east asian width depended on running system environment.
// you can check your system's east asian condition using by:
//	isEastAsian = Default.IsEastAsian
//
var Default = NewCondition(runewidth.EastAsianWidth)

// Condition holds isEastAsian flag and
// can calucate east asian width using that flag.
// zero value of Condition can not use .RuneWidth(), but
// other methods can do.
type Condition struct {
	IsEastAsian bool
	bytesCache  []byte // bytes cache for 1 rune.
}

// return new condition
func NewCondition(isEastAsian bool) *Condition {
	// because 1 rune has 32 bits so bytes has 4*8bits.
	return &Condition{isEastAsian, make([]byte, 4)}
}

// same as BytesWidth exception that input type is string.
func (c Condition) StringWidth(s string) int {
	return c.BytesWidth([]byte(s))
}

// return unicode east asian width in given bytes.
// it will panic if the bytes is invalid utf8 encoding.
func (c Condition) BytesWidth(bs []byte) int {
	w := 0
	for len(bs) > 0 {
		if ok := utf8.FullRune(bs); !ok {
			panic("width: Invalid utf8 byte string, " + string(bs))
		}
		_w, size := c.firstBytesWidth(bs)
		w += _w
		bs = bs[size:]
	}
	return w
}

// return unicode east asian width in a rune.
func (c Condition) RuneWidth(r rune) int {
	n := utf8.EncodeRune(c.bytesCache, r)
	if n == 0 {
		panic("width: Invalid utf8 rune")
	}
	w, _ := c.firstBytesWidth(c.bytesCache[:n])
	return w
}

// return width of first valid utf8 character and its used bytes.
func (c Condition) firstBytesWidth(bs []byte) (int, int) {
	p, size := width.Lookup(bs)
	if size == 0 {
		panic("width: Invalid byte string, " + string(bs))
	}

	var w int
	switch p.Kind() {
	case width.EastAsianNarrow, width.EastAsianHalfwidth:
		w = 1
	case width.EastAsianWide, width.EastAsianFullwidth:
		w = 2
	case width.EastAsianAmbiguous:
		if c.IsEastAsian {
			w = 2
		} else {
			w = 1
		}
	case width.Neutral:
		if b := bs[0]; b == 0 {
			w = 0 // Null character \x00
		} else {
			w = 1
		}
	default:
		panic("width: invalid kind")
	}
	return w, size
}

// return unicode east asian width in given bytes,
// using default condition.
// It will panic if the bytes is invalid utf8 encoding.
func BytesWidth(bs []byte) int {
	return Default.BytesWidth(bs)
}

// return unicode east asian width in given string,
// using default condition.
// It will panic if the string is invalid utf8 encoding.
func StringWidth(s string) int {
	return Default.StringWidth(s)
}

// return unicode east asian width in a rune,
// using default condition.
// It will panic if the rune is invalid.
func RuneWidth(r rune) int {
	return Default.RuneWidth(r)
}
