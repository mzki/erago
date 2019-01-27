package uiadapter

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	attr "local/erago/attribute"
	"local/erago/width"
)

// outputport implemnts a part of flow.GameController.
// It modifies and parses a text to output.
type outputPort struct {
	syncer *lineSyncer
	UI
}

func newOutputPort(ui UI, ls *lineSyncer) *outputPort {
	return &outputPort{
		syncer: ls,
		UI:     ui,
	}
}

const patternString = `\[\s*(\-?[0-9]+)\s*\][\s　]?[^\n]+`

var buttonPattern = regexp.MustCompile(patternString)

// Print text s or print button if text s represents button pattern.
// Given text s must end "\n" or contain no "\n".
func (p outputPort) parsePrint(s string) error {
	loc := buttonPattern.FindStringSubmatchIndex(s)
	if loc == nil {
		return p.UI.Print(s)
	}

	i, j := loc[0], loc[1]
	cmd := s[loc[2]:loc[3]]

	if i > 0 {
		if err := p.UI.Print(s[:i]); err != nil {
			return err
		}
	}

	if err := p.UI.PrintButton(s[i:j], cmd); err != nil {
		return err
	}

	if j < len(s) {
		return p.UI.Print(s[j:])
	}

	return nil
}

// =======================
// --- API for flow.Printer ---
// =======================

func (p outputPort) Print(s string) error {
	return p.printInternal(true, s)
}

func (p outputPort) printInternal(parseButton bool, s string) error {
	for len(s) > 0 {
		var lineSyncRequest = true

		// extract text for each line.
		i := 1 + strings.Index(s, "\n")
		if i == 0 {
			// "\n" is not found
			i = len(s)
			lineSyncRequest = false
		}

		// print
		var err error = nil
		if parseButton {
			err = p.parsePrint(s[:i])
		} else {
			err = p.UI.Print(s[:i])
		}
		if err != nil {
			return err
		}

		// skip processed text
		s = s[i:]

		// synchronize output result either "\n" is appeared or not
		if lineSyncRequest {
			err = p.syncer.SyncLine()
		} else {
			err = p.syncer.SyncText()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p outputPort) withView(vname string, fn func()) error {
	currentName := p.UI.GetCurrentViewName()
	if err := p.UI.SetCurrentView(vname); err != nil {
		return err
	}
	fn()
	p.UI.SetCurrentView(currentName)
	return nil
}

// print string and parse button automatically
func (p outputPort) VPrint(vname, s string) error {
	err := p.withView(vname, func() {
		p.Print(s)
	})
	return err
}

func (p outputPort) PrintL(s string) error {
	return p.Print(s + "\n")
}

// print string + "\n" and parse button automatically
func (p outputPort) VPrintL(vname, s string) error {
	err := p.withView(vname, func() {
		p.PrintL(s)
	})
	return err
}

// print text with padding space to fill at least having the width.
// e.g. text "AAA" with width 5 is "AAA  ".
// But width of multibyte character is counted as 2, while that of single byte character is 1.
// If the text expresses button pattern, the entire text is teasted as Button.
// The text after "\n" is ignored.
func (p outputPort) PrintC(s string, w int) error {
	i := strings.Index(s, "\n")
	if i < 0 {
		i = len(s)
	}
	s = s[:i]

	// padding space to fill width w.
	if space := w - width.StringWidth(s); space > 0 {
		s += strings.Repeat(" ", space)
	}

	var err error = nil
	if loc := buttonPattern.FindStringSubmatchIndex(s); loc != nil {
		err = p.UI.PrintButton(s, s[loc[2]:loc[3]])
	} else {
		err = p.UI.PrintLabel(s)
	}
	if err != nil {
		return err
	}

	return p.syncer.SyncText()
}

// print string.
// it also parses button automatically.
func (p outputPort) VPrintC(vname, s string, width int) error {
	err := p.withView(vname, func() {
		p.PrintC(s, width)
	})
	return err
}

// print text without parsing button pattern.
func (p outputPort) PrintPlain(s string) error {
	return p.printInternal(false, s)
}

// print plain text. no parse button
func (p outputPort) VPrintPlain(vname, s string) error {
	err := p.withView(vname, func() {
		p.PrintPlain(s)
	})
	return err
}

func (p outputPort) VPrintButton(vname, caption, cmd string) error {
	err := p.withView(vname, func() {
		p.PrintButton(caption, cmd)
	})
	return err
}

func buildTextBar(now, max int64, w int, fg, bg string) string {
	w -= 2 // remove frame width, ASCII characters "[" and "]".
	if w <= 0 {
		return "[]"
	}

	// check fg and bg is valid utf8 symbol.
	fg_r, _ := utf8.DecodeRuneInString(fg)
	if fg_r == utf8.RuneError {
		panic("buildTextBar: invalid utf8 string for fg")
	}
	bg_r, _ := utf8.DecodeRuneInString(bg)
	if bg_r == utf8.RuneError {
		panic("buildTextBar: invalid utf8 string for bg")
	}

	now_w := int(float64(w) * float64(now) / float64(max))
	fg_w := width.RuneWidth(fg_r)
	if fg_w == 0 {
		panic("TextBar: zero width fg character")
	}
	result := "[" + strings.Repeat(string(fg_r), now_w/fg_w)

	rest_w := w - now_w
	if rest_w == 0 {
		return result + "]"
	}
	bg_w := width.RuneWidth(bg_r)
	if bg_w == 0 {
		panic("TextBar: zero width bg character")
	}
	result += strings.Repeat(string(bg_r), rest_w/bg_w) + "]"
	// TODO: if rest_w is odd and bg is a multibyte character,
	// returned bar's width is w-1. it should be handled?
	return result
}

// print text bar with current value now, maximum value max, bar's width w,
// bar's symbol fg, and background symbol bg.
// For example, now=3, max=10, w=5, fg='#', and bg='.' then
// prints "[#..]".
func (p outputPort) PrintBar(now, max int64, w int, fg, bg string) error {
	return p.PrintLabel(buildTextBar(now, max, w, fg, bg))
}

func (p outputPort) VPrintBar(vname string, now, max int64, width int, fg, bg string) error {
	err := p.withView(vname, func() {
		p.PrintBar(now, max, width, fg, bg)
	})
	return err
}

// return text represented bar as string.
// it is same as printed text by PrintBar().
func (p outputPort) TextBar(now, max int64, w int, fg, bg string) (string, error) {
	return buildTextBar(now, max, w, fg, bg), nil
}

func (p outputPort) VPrintLine(vname string, sym string) error {
	err := p.withView(vname, func() {
		p.PrintLine(sym)
	})
	return err
}

func (p outputPort) VClearLine(vname string, nline int) error {
	err := p.withView(vname, func() {
		p.ClearLine(nline)
	})
	return err
}

func (p outputPort) VClearLineAll(vname string) error {
	err := p.withView(vname, func() {
		p.ClearLineAll()
	})
	return err
}

func (p outputPort) VNewPage(vname string) error {
	err := p.withView(vname, func() {
		p.NewPage()
	})
	return err
}

var (
	errorEmptyNameNotAllowed = errors.New("empty view name is not allowed")
	errorInvalidBorderRate   = errors.New("border rate must be in [0:1)")
)

func (p outputPort) SetSingleLayout(vname string) error {
	if vname == "" {
		return errorEmptyNameNotAllowed
	}
	return p.UI.SetLayout(attr.NewSingleText(vname))
}

func (p outputPort) SetVerticalLayout(vname1, vname2 string, rate float64) error {
	text1, text2, err := weightedTexts(vname1, vname2, rate)
	if err != nil {
		return err
	}
	return p.UI.SetLayout(attr.NewFlowVertical(text1, text2))
}

func (p outputPort) SetHorizontalLayout(vname1, vname2 string, rate float64) error {
	text1, text2, err := weightedTexts(vname1, vname2, rate)
	if err != nil {
		return err
	}
	return p.UI.SetLayout(attr.NewFlowHorizontal(text1, text2))
}

func weightedTexts(vname1, vname2 string, rate float64) (text1, text2 *attr.LayoutData, err error) {
	if vname1 == "" || vname2 == "" {
		return nil, nil, errorEmptyNameNotAllowed
	}
	if rate <= 0.0 || rate >= 1.0 {
		return nil, nil, errorInvalidBorderRate
	}
	weight1 := int(10.0 * rate)
	weight2 := 10 - weight1

	text1 = attr.WithParentValue(attr.NewSingleText(vname1), weight1)
	text2 = attr.WithParentValue(attr.NewSingleText(vname2), weight2)
	return
}
