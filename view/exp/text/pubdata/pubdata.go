// package pubdata intended for exporting data structure for
// other platforms such as mobile, wasm and so on.
// All of type of exported struct field and function signature should be
// restricted by the restriction of platform conversion.
// e.g. mobile type restrinction is found at
// https://pkg.go.dev/golang.org/x/mobile/cmd/gobind?utm_source=godoc#hdr-Type_restrictions
package pubdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mzki/erago/attribute"
)

const (
	ContentTypeText       = "text"
	ContentTypeTextButton = "text_button"
)

const (
	AlignmentNone   = "none"
	AlignmentLeft   = "left"
	AlignmentCenter = "center"
	AlignmentRight  = "right"
)

var attrAlignmentToStrAlignmentMap = map[attribute.Alignment]string{
	attribute.AlignmentLeft:   AlignmentLeft,
	attribute.AlignmentCenter: AlignmentCenter,
	attribute.AlignmentRight:  AlignmentRight,
}

func AlignmentString(align attribute.Alignment) string {
	str, ok := attrAlignmentToStrAlignmentMap[align]
	if ok {
		return str
	} else {
		return AlignmentNone
	}
}

// Box is abstract content. It holds nomal text, unsplitable text and images and so on.
// The Box is the smallest element for whole content.
// The whole content consist with multiple Paragraph-s, which is divided by the hard return (\n).
// Paragraph consist with multiple Line-s, which is divided by rune width in the maximum width in
// the view window. The Line consists with multiple Box-s, which is divided by its attributes.
//
// The Box type can be validated by type assertion or ContentType().
type Box interface {
	RuneWidth() int      // return box's width in runewidth.
	ContentType() string // return box's content type

	BoxDataUnion // BoxDataUnion is interface for getting data depending on box type.
}

// BoxCommon implements Box interface.
// Data() should be implemented by the derived types.
type BoxCommon struct {
	CommonRuneWidth   int    `json:"rune_width"`
	CommonContentType string `json:"content_type"`

	// Implement BoxDataUnion interface.
	BoxDataUnionImpl
}

func (b *BoxCommon) RuneWidth() int      { return b.CommonRuneWidth }
func (b *BoxCommon) ContentType() string { return b.CommonContentType }

// BoxDataUnion is a union data structure for Box implementers.
type BoxDataUnion interface {
	TextData() *TextData
	TextButtonData() *TextButtonData
}

// BoxDataUnionImpl implements BoxDataUnion interface.
type BoxDataUnionImpl struct{}

func (BoxDataUnionImpl) TextData() *TextData             { panic("Not implemented") }
func (BoxDataUnionImpl) TextButtonData() *TextButtonData { panic("Not implemented") }

// TextData is data for ContentTypeText.
type TextData struct {
	// text content should not contain hard return.
	Text string `json:"text"`
	// Foreground color represents 32bit RGB used to font face color
	FgColor int `json:"fgcolor"`
	// Background color represents 32bit RGB used to background on text.
	BgColor int `json:"bgcolor"`
}

// TextBox represents normal text.
type TextBox struct {
	BoxCommon
	BoxData TextData `json:"data"`
}

// TextData returns *TextData.
func (t *TextBox) TextData() *TextData { return &t.BoxData }

// TextButtonData is data for ContentTypeTextButton.
type TextButtonData struct {
	TextData
	Command string `json:"command"`
}

// TextBottonBox holds normal text and emits input command when this box is tapped/clicked on UI.
type TextButtonBox struct {
	BoxCommon
	BoxData TextButtonData `json:"data"`
}

// TextButtonData returns *TextButtonData.
func (t *TextButtonBox) TextButtonData() *TextButtonData { return &t.BoxData }

// Boxes is intermediation for []Box, used for gomobile export.
type Boxes struct {
	boxes []Box
}

func NewBoxes(boxes []Box) Boxes { return Boxes{boxes} }

// Get() returns a Box at index i, like Boxes[i].
func (bs Boxes) Get(i int) Box { return bs.boxes[i] }

// Len() returns length of array.
func (bs Boxes) Len() int { return len(bs.boxes) }

// Implement json.Mashaller interface. Unmarshal is not considered.
func (bs Boxes) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Boxes []Box `json:"boxes"`
	}{
		Boxes: bs.boxes,
	})
}

// Implement json.Unmarshller interface.
func (bs *Boxes) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	return fmt.Errorf("pubdata.Boxes does not support unmarshal json. %w",
		error(&json.InvalidUnmarshalError{Type: reflect.TypeOf(bs)}))

	/*
		TODO: need type switch for each Box interface unmashal.
		boxes := &struct {
			Boxes []Box `json:"boxes"`
		}{}
		err := json.Unmarshal(b, &boxes)
		if err != nil {
			return err
		}
		bs.boxes = boxes.Boxes
		return nil
	*/
}

// Line is a line in view window.
type Line struct {
	Boxes     Boxes `json:"boxes"`
	RuneWidth int   `json:"rune_width"` // rune width for this line, that is sum of one in boxes.
}

// Lines is intermediation for []Line, used for gomobile export.
type Lines struct {
	lines []Line
}

func NewLines(lines []Line) Lines { return Lines{lines} }

// Get() returns a Line at index i, like Lines[i].
func (ls Lines) Get(i int) Line { return ls.lines[i] }

// Len() returns length of array.
func (ls Lines) Len() int { return len(ls.lines) }

// Implement json.Mashaller interface. Unmarshal is not considered.
func (ls Lines) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Lines []Line `json:"lines"`
	}{
		Lines: ls.lines,
	})
}

// Implement json.Unmarshller interface.
func (ls *Lines) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	lines := struct {
		Lines []Line `json:"lines"`
	}{}
	err := json.Unmarshal(b, &lines)
	if err != nil {
		return err
	}
	ls.lines = lines.Lines
	return nil
}

// Paragraph is a block of content divided by hard return (\n).
type Paragraph struct {
	ID        int    `json:"id"`
	Lines     Lines  `json:"lines"`
	Alignment string `json:"alignment"`
	Fixed     bool   `json:"fixed"`
}
