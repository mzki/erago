package model

import (
	"bytes"
	"testing"

	"github.com/mzki/erago/view/exp/text/pubdata"
)

var (
	testDataParagraphMiddle = pubdata.Paragraph{
		Id: 100,
		Lines: []*pubdata.Line{
			{
				Boxes: []*pubdata.Box{
					{
						RuneWidth:     10,
						LineCountHint: 1,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_TEXT,
						Data: &pubdata.Box_TextData{
							TextData: &pubdata.TextData{
								Text:    "abcdefghij",
								Fgcolor: 0x000000, // Black
								Bgcolor: 0xffffff, // White
							},
						},
					},
					{
						RuneWidth:     10,
						LineCountHint: 10,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_IMAGE,
						Data: &pubdata.Box_ImageData{
							ImageData: &pubdata.ImageData{
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
						ContentType:   pubdata.ContentType_CONTENT_TYPE_SPACE,
						Data: &pubdata.Box_SpaceData{
							SpaceData: &pubdata.SpaceData{},
						},
					},
				},
				RuneWidth: 10 + 10 + 7,
			},
		},
		Alignment: pubdata.Alignment_ALIGNMENT_CENTER,
		Fixed:     true,
	}

	testDataParagraphLarge = pubdata.Paragraph{
		Id: 100,
		Lines: []*pubdata.Line{
			{
				Boxes: []*pubdata.Box{
					{
						RuneWidth:     10,
						LineCountHint: 1,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_TEXT,
						Data: &pubdata.Box_TextData{
							TextData: &pubdata.TextData{
								Text:    "abcdefghij",
								Fgcolor: 0x000000, // Black
								Bgcolor: 0xffffff, // White
							},
						},
					},
					{
						RuneWidth:     10,
						LineCountHint: 10,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_IMAGE,
						Data: &pubdata.Box_ImageData{
							ImageData: &pubdata.ImageData{
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
						ContentType:   pubdata.ContentType_CONTENT_TYPE_SPACE,
						Data: &pubdata.Box_SpaceData{
							SpaceData: &pubdata.SpaceData{},
						},
					},
				},
				RuneWidth: 10 + 10 + 7,
			},
			{
				Boxes: []*pubdata.Box{
					{
						RuneWidth:     10,
						LineCountHint: 1,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_TEXT,
						Data: &pubdata.Box_TextData{
							TextData: &pubdata.TextData{
								Text:    "abcdefghij",
								Fgcolor: 0x000000, // Black
								Bgcolor: 0xffffff, // White
							},
						},
					},
					{
						RuneWidth:     10,
						LineCountHint: 10,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_IMAGE,
						Data: &pubdata.Box_ImageData{
							ImageData: &pubdata.ImageData{
								Source:          "/path/to/image.png",
								WidthPx:         100,
								HeightPx:        100,
								WidthTextScale:  15,
								HeightTextScale: 10,
								Data:            bytes.Repeat([]byte{0x0, 0x1, 0x2}, 1*1024*1024/2), // 1.5MiB
								DataFetchType:   0,
							},
						},
					},
					{
						RuneWidth:     7,
						LineCountHint: 1,
						ContentType:   pubdata.ContentType_CONTENT_TYPE_SPACE,
						Data: &pubdata.Box_SpaceData{
							SpaceData: &pubdata.SpaceData{},
						},
					},
				},
				RuneWidth: 10 + 10 + 7,
			},
		},
		Alignment: pubdata.Alignment_ALIGNMENT_CENTER,
		Fixed:     true,
	}
)

func Test_newParagraphBinaryEncodeFunc(t *testing.T) {
	type args struct {
		encoding int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "json", args: args{MessageByteEncodingJson}, wantErr: false},
		{name: "protobuf", args: args{MessageByteEncodingProtobuf}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encodeFunc := newParagraphBinaryEncodeFunc(tt.args.encoding)
			bs, err := encodeFunc(&testDataParagraphMiddle)
			if !tt.wantErr && err != nil {
				t.Errorf("newParagraphBinaryEncodeFunc().encode(): got error %v", err)
			}
			if len(bs) == 0 {
				t.Errorf("newParagraphBinaryEncodeFunc().encode(): returns zero bytes")
			}
		})
	}
}

func BenchmarkParagraphBinaryEncodeFunc_Middle_Json(b *testing.B) {
	encodeFunc := newParagraphBinaryEncodeFunc(MessageByteEncodingJson)
	b.ResetTimer()
	var bsSize int
	for i := 0; i <= b.N; i++ {
		bs, _ := encodeFunc(&testDataParagraphMiddle)
		bsSize = len(bs)
		_ = bs
	}
	b.Logf("encode result = %v bytes/message ", bsSize)
}

func BenchmarkParagraphBinaryEncodeFunc_Middel_Proto(b *testing.B) {
	encodeFunc := newParagraphBinaryEncodeFunc(MessageByteEncodingProtobuf)
	b.ResetTimer()
	var bsSize int
	for i := 0; i <= b.N; i++ {
		bs, _ := encodeFunc(&testDataParagraphMiddle)
		bsSize = len(bs)
		_ = bs
	}
	b.Logf("encode result = %v bytes/message ", bsSize)
}

func BenchmarkParagraphBinaryEncodeFunc_Large_Json(b *testing.B) {
	encodeFunc := newParagraphBinaryEncodeFunc(MessageByteEncodingJson)
	b.ResetTimer()
	var bsSize int
	for i := 0; i <= b.N; i++ {
		bs, _ := encodeFunc(&testDataParagraphLarge)
		bsSize = len(bs)
		_ = bs
	}
	b.Logf("encode result = %v bytes/message ", bsSize)
}

func BenchmarkParagraphBinaryEncodeFunc_Large_Proto(b *testing.B) {
	encodeFunc := newParagraphBinaryEncodeFunc(MessageByteEncodingProtobuf)
	b.ResetTimer()
	var bsSize int
	for i := 0; i <= b.N; i++ {
		bs, _ := encodeFunc(&testDataParagraphLarge)
		bsSize = len(bs)
		_ = bs
	}
	b.Logf("encode result = %v bytes/message ", bsSize)
}
