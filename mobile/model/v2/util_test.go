package model

import (
	"reflect"
	"testing"

	"golang.org/x/image/math/fixed"
)

func Test_floatToFixedInt(t *testing.T) {
	type args struct {
		x float64
	}
	tests := []struct {
		name string
		args args
		want fixed.Int26_6
	}{
		{"1.25", args{1.25}, fixed.Int26_6(1<<6 | 1<<4)},
		{"32.0625", args{32.0625}, fixed.Int26_6(32<<6 | 1<<2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floatToFixedInt(tt.args.x); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("floatToFixedInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
