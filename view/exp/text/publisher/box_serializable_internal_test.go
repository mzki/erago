package publisher

import (
	"bytes"
	"encoding/json"
	"image/color"
	"testing"
)

func TestParagraph_JsonDump(t *testing.T) {
	p := Paragraph{
		Lines: Lines{
			[]Line{
				{
					Boxes: Boxes{
						[]Box{
							&TextBox{
								BoxCommon{
									CommonRuneWidth:   10,
									CommonContentType: ContentTypeText,
								},
								TextData{
									Text:    "abcdefghij",
									FgColor: color.RGBA{0x0, 0x0, 0x0, 0xff},    // Black
									BgColor: color.RGBA{0xff, 0xff, 0xff, 0xff}, // White
								},
							},
						},
					},
					RuneWidth: 10,
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(p)
	if err != nil {
		t.Fatal(err)
	}
	jsonbytes := buf.Bytes()
	t.Logf("Dumped Json:\n%v", string(jsonbytes))

	var decodedP Paragraph
	if err := json.Unmarshal(jsonbytes, &decodedP); err == nil {
		t.Fatalf("Unmarshal is not supported but return nil. from json: %v", jsonbytes)
	} else {
		t.Logf("skipped gotten error: %v", err)
	}
	/* Unmarshal is not supported.
	if !reflect.DeepEqual(p, decodedP) {
		t.Errorf("Before/After Json dumping Paragraph is not matched. want: %v got: %v", p, decodedP)
	}
	*/
}
