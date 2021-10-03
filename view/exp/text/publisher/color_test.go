package publisher_test

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/mzki/erago/view/exp/text/publisher"
)

func TestUIntRGBToColorRGBA(t *testing.T) {
	type args struct {
		c uint32
	}
	tests := []struct {
		name string
		args args
		want color.RGBA
	}{
		{
			name: "First",
			args: args{0xffccaa},
			want: color.RGBA{0xff, 0xcc, 0xaa, 0xff},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publisher.UInt32RGBToColorRGBA(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UIntRGBToColorRGBA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntRGBToColorRGBA(t *testing.T) {
	type args struct {
		c int32
	}
	tests := []struct {
		name string
		args args
		want color.RGBA
	}{
		{
			name: "First",
			args: args{0xffccaa},
			want: color.RGBA{0xff, 0xcc, 0xaa, 0xff},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publisher.Int32RGBToColorRGBA(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UIntRGBToColorRGBA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorRGBAToUIntRGB(t *testing.T) {
	type args struct {
		c color.RGBA
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "First",
			args: args{color.RGBA{0xff, 0xcc, 0xaa, 0x88}},
			want: 0xffccaa,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publisher.ColorRGBAToUInt32RGB(tt.args.c); got != tt.want {
				t.Errorf("ColorRGBAToUIntRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorRGBAToIntRGB(t *testing.T) {
	type args struct {
		c color.RGBA
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "First",
			args: args{color.RGBA{0xff, 0xcc, 0xaa, 0x88}},
			want: 0xffccaa,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publisher.ColorRGBAToInt32RGB(tt.args.c); got != tt.want {
				t.Errorf("ColorRGBAToUIntRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}
