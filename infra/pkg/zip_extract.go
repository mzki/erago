//go:build !(android || ios || js || wasip1)

package pkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mzki/erago/filesystem"
)

// ExtractFromZip extracts all files in zip archive of srcZipPath.
// It returns extracted root directory name at 1st, and error if something fails, otherwise nil at 2nd.
// Output directory is abstracted by dstFsys. For example dstFsys = os.DirFS("/example-dir") and
// zip archive contains test/file.txt, then the result will be stored in /exmaple-dir/test/file.txt.
// srcZipPath is searched with respect to srcFsys. To be better memory efficiency, the fs.File returned by
// srcFsys.Open() should implement io.ReaderAt interface.
func ExtractFromZip(dstFsys filesystem.FileSystem, srcFsys fs.FS, srcZipPath string) (string, error) {
	file, err := srcFsys.Open(srcZipPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	finfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	var readerAt io.ReaderAt
	if r, ok := file.(io.ReaderAt); ok {
		readerAt = r
	} else {
		// fallback method: put all content in memory. This would be insufficient way.
		bs, err := io.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("read content failed for: %v", srcZipPath)
		}
		readerAt = bytes.NewReader(bs)
	}
	return ExtractFromZipReader(dstFsys, readerAt, finfo.Size())
}

// ExtractFromZipReader is alternative API with io.ReaderAt for ExtractFromZip. See ExtractFromZip documentation for the details.
func ExtractFromZipReader(outFs filesystem.FileSystem, r io.ReaderAt, rSize int64) (string, error) {
	zReader, err := zip.NewReader(r, rSize)
	if err != nil {
		return "", err
	}
	var extractedRootDir string = "./"
	for _, file := range zReader.File {
		if file.NonUTF8 {
			return extractedRootDir, fmt.Errorf("zip archive containing non-UTF8 file name, is now allowed: file name: %v", file.FileInfo().Name())
		}
		if file.FileInfo().IsDir() {
			continue // to ignore extract directory itself, means empty directory never created.
		}

		if err := extractZipFileEntry(outFs, file); err != nil {
			return extractedRootDir, fmt.Errorf("failed to extract zip file: %w", err)
		}

		// update root dir once
		if extractedRootDir == "./" {
			slist := strings.Split(file.Name, "/")
			if len(slist) > 0 {
				if root := slist[0]; len(root) > 0 {
					extractedRootDir = root
				}
			}
		}
	}
	return extractedRootDir, nil
}

func extractZipFileEntry(outFs filesystem.FileSystem, srcFile *zip.File) error {
	path := filepath.FromSlash(srcFile.Name)
	dst, err := outFs.Store(path)
	if err != nil {
		return fmt.Errorf("output path(%v) open failed: %w", path, err)
	}
	defer dst.Close()

	src, err := srcFile.Open()
	if err != nil {
		return fmt.Errorf("zip file entry(%v) open failed: %w", srcFile.Name, err)
	}
	defer src.Close()

	return copyLimited(dst, path, src, srcFile.Name)
}
