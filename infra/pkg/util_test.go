package pkg

import (
	"bytes"
	"io"
	"testing"
)

func Test_copyLimitedN(t *testing.T) {
	type args struct {
		dstPath string
		src     io.Reader
		srcPath string
		nPlus1  int64
	}
	tests := []struct {
		name    string
		args    args
		wantDst string
		wantErr bool
	}{
		{
			name:    "normal-fit",
			args:    args{"dst", bytes.NewReader([]byte("test-content")), "src", 12 + 1},
			wantDst: "test-content",
			wantErr: false,
		},
		{
			name:    "normal-less-than-limit",
			args:    args{"dst", bytes.NewReader([]byte("test-content")), "src", 13 + 1},
			wantDst: "test-content",
			wantErr: false,
		},
		{
			name:    "error-exceeds-limit+1",
			args:    args{"dst", bytes.NewReader([]byte("test-content")), "src", 11 + 1},
			wantDst: "test-content", // content is fully copied, but treat as error since reaches the maximum
			wantErr: true,
		},
		{
			name:    "error-exceeds-limit+2",
			args:    args{"dst", bytes.NewReader([]byte("test-content")), "src", 10 + 1},
			wantDst: "test-conten",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &bytes.Buffer{}
			if err := copyLimitedN(dst, tt.args.dstPath, tt.args.src, tt.args.srcPath, tt.args.nPlus1); (err != nil) != tt.wantErr {
				t.Errorf("copyLimitedN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDst := dst.String(); gotDst != tt.wantDst {
				t.Errorf("copyLimitedN() = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}
