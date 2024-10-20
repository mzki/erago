package model

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/mzki/erago/app"
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

func createMobileFS(absDir string, backend filesystem.FileSystemGlob) filesystem.FileSystemGlobPR {
	mobileFS := filesystem.AbsDirFileSystem(absDir)
	if backend != nil {
		mobileFS.Backend = backend
	} else {
		mobileFS.Backend = &filesystem.OSFileSystem{MaxFileSize: filesystem.DefaultMaxFileSize}
	}
	return mobileFS
}

func ExportSav(absEragoDir string, eragoFsys filesystem.FileSystemGlob) ([]byte, error) {
	oldDefaultFS := filesystem.Default
	defer func() { filesystem.Default = oldDefaultFS }()

	mobileFS := createMobileFS(absEragoDir, eragoFsys)
	filesystem.Default = mobileFS

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		return nil, err
	}

	savPatterns := filepath.Join(appConf.Game.RepoConfig.SaveFileDir, "*")
	savFiles, err := mobileFS.Glob(savPatterns)
	if err != nil {
		return nil, err
	}

	// Trimming common directory of file paths. Zip spec needs relative path for each file.
	// savFiles can be relative path since those are assumed under mobileFS(Dir: absEragoDir)
	// TODO: move this feature in pkg.ArchiveAsZip* ?
	for i, savFile := range savFiles {
		relSavFile, err := filepath.Rel(absEragoDir, savFile)
		if err != nil {
			return nil, fmt.Errorf("could not get relative path for %v: %w", savFile, err)
		}
		savFiles[i] = relSavFile
	}

	writer := new(bytes.Buffer)
	if err := pkg.ArchiveAsZipWriter(writer, "erago-sav", mobileFS.(fs.FS), savFiles); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
