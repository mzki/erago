package pubdata

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestParagraph_JsonDump(t *testing.T) {
	p := Paragraph{
		ID: 100,
		Lines: Lines{
			[]Line{
				{
					Boxes: Boxes{
						[]Box{
							&TextBox{
								BoxCommon{
									CommonRuneWidth:     10,
									CommonLineCountHint: 1,
									CommonContentType:   ContentTypeText,
								},
								TextData{
									Text:    "abcdefghij",
									FgColor: 0x000000, // Black
									BgColor: 0xffffff, // White
								},
							},
							&ImageBox{
								BoxCommon{
									CommonRuneWidth:     10,
									CommonLineCountHint: 10,
									CommonContentType:   ContentTypeImage,
								},
								ImageData{
									Source:          "/path/to/image.png",
									WidthPx:         100,
									HeightPx:        100,
									WidthTextScale:  15,
									HeightTextScale: 10,
									Data:            []byte{0x0, 0x1, 0x2},
									DataFetchType:   0,
								},
							},
							&SpaceBox{
								BoxCommon{
									CommonRuneWidth:     7,
									CommonLineCountHint: 1,
									CommonContentType:   ContentTypeSpace,
								},
							},
						},
					},
					RuneWidth: 10 + 10 + 7,
				},
			},
		},
		Alignment: AlignmentCenter,
		Fixed:     true,
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
