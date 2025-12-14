package pubdata

import (
	"encoding/json"
	"testing"
)

func TestParagraphList_JsonDump(t *testing.T) {
	p := ParagraphList{
		Paragraphs: []*Paragraph{
			{
				Id: 100,
				Lines: []*Line{
					{
						Boxes: []*Box{
							{
								RuneWidth:     10,
								LineCountHint: 1,
								ContentType:   ContentType_CONTENT_TYPE_TEXT,
								Data: &Box_TextData{
									TextData: &TextData{
										Text:    "abcdefghij",
										Fgcolor: 0x000000, // Black
										Bgcolor: 0xffffff, // White
									},
								},
							},
							{
								RuneWidth:     10,
								LineCountHint: 10,
								ContentType:   ContentType_CONTENT_TYPE_IMAGE,
								Data: &Box_ImageData{
									ImageData: &ImageData{
										Source:          "/path/to/image.png",
										WidthPx:         100,
										HeightPx:        100,
										WidthTextScale:  15,
										HeightTextScale: 10,
										Data:            []byte{0x0, 0x1, 0x2},
										DataFetchType:   0,
									},
								},
							},
							{
								RuneWidth:     7,
								LineCountHint: 1,
								ContentType:   ContentType_CONTENT_TYPE_SPACE,
								Data: &Box_SpaceData{
									SpaceData: &SpaceData{},
								},
							},
						},
						RuneWidth: 10 + 10 + 7,
					},
				},
				Alignment: Alignment_ALIGNMENT_CENTER,
				Fixed:     true,
			},
		},
	}

	jsonbytes, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Dumped Json:\n%v", string(jsonbytes))

	var decodedP ParagraphList
	if err := json.Unmarshal(jsonbytes, &decodedP); err != nil {
		t.Fatalf("Unmarshall error: %v", err)
	}
	if !p.EqualVT(&decodedP) {
		t.Errorf("Before/After Json dumping Paragraph is not matched. want: %#v got: %#v", &p, &decodedP)
	}

	pbytes, err := p.MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Dumped proto: %#v\n", pbytes)
	decodedP.Reset()
	if err := decodedP.UnmarshalVT(pbytes); err != nil {
		t.Logf("skipped gotten error: %v", err)
	}
	// Unmarshal is not supported.
	if !p.EqualVT(&decodedP) {
		t.Errorf("Before/After Proto dumping Paragraph is not matched. want: %#v got: %#v", &p, &decodedP)
	}
}
