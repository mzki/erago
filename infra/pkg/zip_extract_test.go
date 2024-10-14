//go:build !(android || ios || js || wasip1)

package pkg

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/mzki/erago/filesystem"
)

func TestExtractFromZip(t *testing.T) {
	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	absTempDirFiles := func(prefixDir string, ss []string) (sss []string) {
		for _, s := range ss {
			sss = append(sss, filepath.Join(absTempDir, prefixDir, s))
		}
		return sss
	}
	type args struct {
		dstFsys    filesystem.FileSystem
		srcFsys    fs.FS
		srcZipPath string
	}
	tests := []struct {
		name               string
		args               args
		wantErr            bool
		checkExtractedFile bool
		extractedFiles     []string
	}{
		{
			name: "normal",
			args: args{
				dstFsys:    filesystem.AbsDirFileSystem(absTempDir),
				srcFsys:    testdataFS,
				srcZipPath: "testdata/archive-golden.zip",
			},
			wantErr:            false,
			checkExtractedFile: true,
			extractedFiles:     absTempDirFiles("archive-golden", CollectFiles(testZipArchiveFS, ".")),
		},
		{
			name: "linux zip",
			args: args{
				dstFsys:    filesystem.AbsDirFileSystem(absTempDir),
				srcFsys:    testdataFS,
				srcZipPath: "testdata/archive-linux.zip",
			},
			wantErr:            false,
			checkExtractedFile: true,
			extractedFiles:     absTempDirFiles("archive-linux", CollectFiles(testZipArchiveFS, ".")),
		},
		{
			name: "win buildin zip",
			args: args{
				dstFsys:    filesystem.AbsDirFileSystem(absTempDir),
				srcFsys:    testdataFS,
				srcZipPath: "testdata/archive-win11-builtin.zip",
			},
			wantErr:            false,
			checkExtractedFile: true,
			extractedFiles:     absTempDirFiles("archive-win11-builtin", CollectFiles(testZipArchiveFS, ".")),
		},
		{
			name: "win 7-zip zip",
			args: args{
				dstFsys:    filesystem.AbsDirFileSystem(absTempDir),
				srcFsys:    testdataFS,
				srcZipPath: "testdata/archive-win-7-zip.zip",
			},
			wantErr:            true, // error by utf8 encoding bit not set.
			checkExtractedFile: false,
			extractedFiles:     absTempDirFiles("archive-win-7-zip", CollectFiles(testZipArchiveFS, ".")),
		},
		{
			name: "win 7-zip cu=on zip",
			args: args{
				dstFsys:    filesystem.AbsDirFileSystem(absTempDir),
				srcFsys:    testdataFS,
				srcZipPath: "testdata/archive-win-7-zip-cuon.zip",
			},
			wantErr:            false,
			checkExtractedFile: true,
			extractedFiles:     absTempDirFiles("archive-win-7-zip-cuon", CollectFiles(testZipArchiveFS, ".")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExtractFromZip(tt.args.dstFsys, tt.args.srcFsys, tt.args.srcZipPath); (err != nil) != tt.wantErr {
				t.Errorf("ExtractFromZip() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == false && tt.checkExtractedFile == true {
				for _, file := range tt.extractedFiles {
					if _, err := os.Stat(file); err != nil {
						t.Errorf("ExtractFromZip() missing extracted file (%v), error %v", file, err)
					}
				}
			}
		})
	}
}

func TestExtractFromZipReader(t *testing.T) {
	absTempDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	goldenBytes, err := testdataFS.ReadFile("testdata/archive-golden.zip")
	if err != nil {
		t.Fatal(err)
	}
	absTempDirFiles := func(prefixDir string, ss []string) (sss []string) {
		for _, s := range ss {
			sss = append(sss, filepath.Join(absTempDir, prefixDir, s))
		}
		return sss
	}
	type args struct {
		outFs filesystem.FileSystem
		r     io.ReaderAt
		rSize int64
	}
	tests := []struct {
		name               string
		args               args
		wantErr            bool
		checkExtractedFile bool
		extractedFiles     []string
	}{
		{
			name: "normal",
			args: args{
				outFs: filesystem.AbsDirFileSystem(absTempDir),
				r:     bytes.NewReader(goldenBytes),
				rSize: int64(len(goldenBytes)),
			},
			wantErr:            false,
			checkExtractedFile: true,
			extractedFiles:     absTempDirFiles("archive-golden", CollectFiles(testZipArchiveFS, ".")),
		},
		{
			name: "invalid zip source",
			args: args{
				outFs: filesystem.AbsDirFileSystem(absTempDir),
				r:     bytes.NewReader([]byte{0x34, 0x00, 0x11, 0xaa}),
				rSize: int64(4),
			},
			wantErr:            true,
			checkExtractedFile: false,
			extractedFiles:     []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExtractFromZipReader(tt.args.outFs, tt.args.r, tt.args.rSize); (err != nil) != tt.wantErr {
				t.Errorf("ExtractFromZipReader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == false && tt.checkExtractedFile == true {
				for _, file := range tt.extractedFiles {
					if _, err := os.Stat(file); err != nil {
						t.Errorf("ExtractFromZipReader() missing extracted file (%v), error %v", file, err)
					}
				}
			}
		})
	}
}
