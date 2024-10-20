package model

import (
	"bytes"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/pkg"
)

// NewOsDirFileSystem is helper function to create filesytem under directory of absPath.
// The backend is OS filesytem to open file.
func NewOSDirFileSystem(absPath string) filesystem.FileSystem {
	absDirFs := filesystem.AbsDirFileSystem(absPath)
	absDirFs.Backend = &filesystem.OSFileSystem{MaxFileSize: filesystem.DefaultMaxFileSize}
	return absDirFs
}

// InstallPackage extract zip archive into outFsys.
// It may useful to store erago related files into convenient location when platform has file access limitation for default location.
// It returns base name of extractedDir at 1st and error at 2nd.
// The path for extracted root directory will be [Directory of outFsys]/[extractedDir].
func InstallPackage(outFsys filesystem.FileSystem, zipBytes []byte) (extractedDir string, err error) {
	return pkg.ExtractFromZipReader(outFsys, bytes.NewReader(zipBytes), int64(len(zipBytes)))
}
