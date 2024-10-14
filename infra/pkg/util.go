package pkg

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/mzki/erago/filesystem"
)

// MaxFileSizePlus1InByte indicate default valoe of the maximum file size + 1 to be written or read for a file.
const MaxFileSizePlus1InByte = filesystem.DefaultMaxFileSize + 1

// copyLimited is similar API except the size limiation uses MaxFileSizePlus1InByte.
// That means src size < MaxFileSizePlus1InByte will be accepted.
func copyLimited(dst io.Writer, dstPath string, src io.Reader, srcPath string) error {
	return copyLimitedN(dst, dstPath, src, srcPath, MaxFileSizePlus1InByte)
}

// ErrTooLargeBytes indicates copy operation failed due to too large bytes to read or write.
var ErrTooLargeBytes = fmt.Errorf("too large bytes")

// copyLimited copies content from src into dst with size limitation, nPlus1.
// It returns nil when copy is done and written bytes are less than nPlus1, returns ErrTooLargeBytes when
// src contains large bytes more than or equal to nPlus1 and returns other errors when something failed.
func copyLimitedN(dst io.Writer, dstPath string, src io.Reader, srcPath string, nPlus1 int64) error {
	if _, err := io.CopyN(dst, src, nPlus1); !errors.Is(err, io.EOF) {
		if err == nil {
			// reaches maximum size, but still remain content from source.
			// Source file seems to be exceeded its file content. reject it as size limitation.
			return &fs.PathError{Op: "copy", Path: srcPath, Err: fmt.Errorf("exceed limit (%v): %w", nPlus1, ErrTooLargeBytes)}
		} else {
			// something went wrong with err.
			return &fs.PathError{Op: "copy", Path: dstPath, Err: err}
		}
	}
	// read by EOF
	return nil
}
