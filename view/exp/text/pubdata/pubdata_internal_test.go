package pubdata

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestParagraph_JsonDump(t *testing.T) {
	p := Paragraph{
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
	}

	jsonbytes, err := protojson.MarshalOptions{
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Dumped Json:\n%v", string(jsonbytes))

	var decodedP Paragraph
	if err := protojson.Unmarshal(jsonbytes, &decodedP); err != nil {
		t.Logf("skipped gotten error: %v", err)
	}
	// Unmarshal is not supported.
	if !proto.Equal(&p, &decodedP) {
		t.Errorf("Before/After Json dumping Paragraph is not matched. want: %#v got: %#v", &p, &decodedP)
	}

	pbytes, err := proto.MarshalOptions{
		Deterministic: true,
	}.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	decodedP.Reset()
	if err := proto.Unmarshal(pbytes, &decodedP); err != nil {
		t.Logf("skipped gotten error: %v", err)
	}
	// Unmarshal is not supported.
	if !proto.Equal(&p, &decodedP) {
		t.Errorf("Before/After Proto dumping Paragraph is not matched. want: %#v got: %#v", &p, &decodedP)
	}
}
