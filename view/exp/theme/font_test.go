package theme

import (
	"reflect"
	"sync"
	"testing"

	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func TestFontFaceOptions_TTFOptions(t *testing.T) {
	type fields struct {
		Size float64
		DPI  float64
	}
	tests := []struct {
		name   string
		fields fields
		want   *opentype.FaceOptions
	}{
		{"normal", fields{10.0, 96.0}, &opentype.FaceOptions{Size: 10.0, DPI: 96.0}},
		{"empty-size", fields{0.0, 96.0}, &opentype.FaceOptions{Size: DefaultFontSize, DPI: 96.0}},
		{"empty-dpi", fields{10.0, 0.0}, &opentype.FaceOptions{Size: 10.0, DPI: DefaultDPI}},
		{"empty-both", fields{}, &opentype.FaceOptions{Size: DefaultFontSize, DPI: DefaultDPI}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &FontFaceOptions{
				Size: tt.fields.Size,
				DPI:  tt.fields.DPI,
			}
			if got := opt.TTFOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontFaceOptions.TTFOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDefaultFace(t *testing.T) {
	type args struct {
		opt *FontFaceOptions
	}
	tests := []struct {
		name string
		args args
	}{
		{"just create", args{&FontFaceOptions{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			face := NewDefaultFace(tt.args.opt)
			met := face.Metrics()
			if met.Ascent == 0 || met.Descent == 0 || met.Height == 0 {
				t.Errorf("NewDefaultFace() returns empty metric")
			}
		})
	}
}

func TestParseTruetypeFileFace(t *testing.T) {
	type args struct {
		file string
		opt  *FontFaceOptions
	}
	tests := []struct {
		name    string
		args    args
		want    font.Face
		wantErr bool
	}{
		{"default font is not parsable in this function", args{DefaultFontName, &FontFaceOptions{}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTruetypeFileFace(tt.args.file, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTruetypeFileFace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTruetypeFileFace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewOneFontFaceCatalog(t *testing.T) {
	type args struct {
		fontfile string
		o        *FontFaceOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default font", args{DefaultFontName, &FontFaceOptions{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOneFontFaceCatalog(tt.args.fontfile, tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOneFontFaceCatalog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestOneFontFaceCatalog_UpdateFontFaceOptions(t *testing.T) {
	must := func(cat *OneFontFaceCatalog, err error) *OneFontFaceCatalog {
		if err != nil {
			panic(err)
		}
		return cat
	}
	type fields struct {
		cat *OneFontFaceCatalog
		opt *FontFaceOptions
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"just update", fields{must(NewOneFontFaceCatalog(DefaultFontName, &FontFaceOptions{})), &FontFaceOptions{Size: 24, DPI: 96}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.cat.UpdateFontFaceOptions(tt.fields.cat.opt)
		})
	}
}

func TestOneFontFaceCatalog_AcquireFontFace(t *testing.T) {
	type fields struct {
		font *opentype.Font
		mu   *sync.Mutex
		opt  *FontFaceOptions
		face font.Face
	}
	type args struct {
		in0 theme.FontFaceOptions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   font.Face
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := &OneFontFaceCatalog{
				font: tt.fields.font,
				mu:   tt.fields.mu,
				opt:  tt.fields.opt,
				face: tt.fields.face,
			}
			if got := cat.AcquireFontFace(tt.args.in0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OneFontFaceCatalog.AcquireFontFace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOneFontFaceCatalog_ReleaseFontFace(t *testing.T) {
	type fields struct {
		font *opentype.Font
		mu   *sync.Mutex
		opt  *FontFaceOptions
		face font.Face
	}
	type args struct {
		in0 theme.FontFaceOptions
		in1 font.Face
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := &OneFontFaceCatalog{
				font: tt.fields.font,
				mu:   tt.fields.mu,
				opt:  tt.fields.opt,
				face: tt.fields.face,
			}
			cat.ReleaseFontFace(tt.args.in0, tt.args.in1)
		})
	}
}
