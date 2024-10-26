package model

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mzki/erago/app"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/pkg"
	"github.com/psanford/memfs"
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

type wrapFileSystemPR struct {
	filesystem.FileSystemGlob
}

// implements PathResolver interface
func (fsys wrapFileSystemPR) ResolvePath(file string) (string, error) {
	if pr, ok := fsys.FileSystemGlob.(filesystem.PathResolver); ok {
		return pr.ResolvePath(file)
	} else {
		return "", fmt.Errorf("ResolvePath err: %w", errors.ErrUnsupported)
	}
}

var ErrNoSavFiles = errors.New("no sav file")

// ExportSav exports save files matching the pattern [absEragoDir]/[saveFileDir in the config file]/*.
// It returns exported save files as zip archive bytes and error if any.
// If any save files not found, it returns error with ErrNoSavFiles.
// eragoFsys is used for reading/seaching save files and config file.
// absEragoDir should be same directory with eragoFsys's current directory.
func ExportSav(absEragoDir string, eragoFsys filesystem.FileSystemGlob) ([]byte, error) {
	oldDefaultFS := filesystem.Default
	defer func() { filesystem.Default = oldDefaultFS }()

	filesystem.Default = wrapFileSystemPR{eragoFsys}

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get save file directory from config: %w", err)
	}

	savPatterns := filepath.Join(appConf.Game.RepoConfig.SaveFileDir, "*")
	savFiles, err := eragoFsys.Glob(savPatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to get save files from %s: %w", savPatterns, err)
	}
	if len(savFiles) == 0 {
		return nil, fmt.Errorf("no save files found for %v: %w", savPatterns, ErrNoSavFiles)
	}

	// Trimming common directory of file paths. Zip spec needs relative path for each file.
	// savFiles can be relative path since those are assumed under mobileFS(Dir: absEragoDir)
	// TODO: move this feature in pkg.ArchiveAsZip* ?
	for i, savFile := range savFiles {
		if strings.HasPrefix(filepath.ToSlash(savFile), filepath.ToSlash(absEragoDir)) {
			relSavFile, err := filepath.Rel(absEragoDir, savFile)
			if err != nil {
				return nil, fmt.Errorf("could not get relative path for %v: %w", savFile, err)
			}
			savFiles[i] = relSavFile
		}
	}

	// create in-memory fs.FS to adapt ArchiveAsZip API.
	// TODO: is there more efficient way?
	var eragoFS fs.FS
	{
		rootFs := memfs.New(memfs.WithOpenHook(func(s string, b []byte, err error) ([]byte, error) {
			// lazy load for data
			if err != nil {
				return nil, err
			}
			r, err := eragoFsys.Load(s)
			if err != nil {
				return nil, err
			}
			defer r.Close()
			return io.ReadAll(r)
		}))
		savDir := filepath.ToSlash(appConf.Game.RepoConfig.SaveFileDir)
		if err := rootFs.MkdirAll(savDir, fs.ModePerm|fs.ModeDir); err != nil {
			return nil, fmt.Errorf("failed to create internal savDir for %s: %w", savDir, err)
		}
		dummyBytes := []byte{} // actual data is retrieved from OpenHook.
		for _, savFile := range savFiles {
			savF := filepath.ToSlash(savFile)
			if err := rootFs.WriteFile(savF, dummyBytes, fs.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create internal savFile for %s: %w", savF, err)
			}
		}
		eragoFS = rootFs
	}

	writer := new(bytes.Buffer)
	if err := pkg.ArchiveAsZipWriter(writer, "erago-sav", eragoFS, savFiles); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
