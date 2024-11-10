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
func NewOSDirFileSystem(absPath string) FileSystemGlob {
	absDirFs := filesystem.AbsDirFileSystem(absPath)
	absDirFs.Backend = &filesystem.OSFileSystem{MaxFileSize: filesystem.DefaultMaxFileSize}
	return FromGoFSGlob(absDirFs)
}

// InstallPackage extract zip archive into outFsys.
// It may useful to store erago related files into convenient location when platform has file access limitation for default location.
// It returns base name of extractedDir at 1st and error at 2nd.
// The path for extracted root directory will be [Directory of outFsys]/[extractedDir].
func InstallPackage(outFsys FileSystem, zipBytes []byte) (extractedDir string, err error) {
	return pkg.ExtractFromZipReader(FromMobileFS(outFsys), bytes.NewReader(zipBytes), int64(len(zipBytes)))
}

func createMobileFS(absDir string, backend FileSystemGlob) filesystem.FileSystemGlobPR {
	mobileFS := filesystem.AbsDirFileSystem(absDir)
	if backend != nil {
		mobileFS.Backend = FromMobileFSGlob(backend)
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

var (
	ErrNoSavFiles = errors.New("no sav file")
	ErrNoLogFile  = errors.New("no log file")
)

// IsExportFileNotFound indicates error is happend by export target file(s) are not found.
// Golang side can detect is by errors.Is(...), but mobile platform side can not.
// This functions would help to detect such kind of error at mobile platform side.
func IsExportFileNotFound(err error) bool {
	return errors.Is(err, ErrNoSavFiles) || errors.Is(err, ErrNoLogFile)
}

// ExportSav exports save files matching the pattern [absEragoDir]/[saveFileDir in the config file]/*.
// It returns exported save files as zip archive bytes and error if any.
// If any save files not found, it returns error with ErrNoSavFiles.
// eragoFsys is used for reading/seaching save files and config file.
// absEragoDir should be same directory with eragoFsys's current directory.
func ExportSav(absEragoDir string, eragoFsys FileSystemGlob) ([]byte, error) {
	oldDefaultFS := filesystem.Default
	defer func() { filesystem.Default = oldDefaultFS }()

	filesystem.Default = wrapFileSystemPR{FromMobileFSGlob(eragoFsys)}

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get save file directory from config: %w", err)
	}

	savPatterns := filepath.Join(appConf.Game.RepoConfig.SaveFileDir, "*")
	savFiles, err := filesystem.Default.Glob(savPatterns)
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
	if err := pkg.ArchiveAsZipWriter(writer, "", eragoFS, savFiles); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func ImportSav(absEragoDir string, eragoFsys FileSystemGlob, savZipBytes []byte) error {
	oldDefaultFS := filesystem.Default
	defer func() { filesystem.Default = oldDefaultFS }()

	filesystem.Default = wrapFileSystemPR{FromMobileFSGlob(eragoFsys)}

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to get save file directory from config: %w", err)
	}
	_, _ = disableDesktopFeatures(appConf)

	// TODO: check sav directory mismatch. Possible ways:
	// 1. output inmemory outFsys, then check file path
	// 2. Create wrapped filesystem and inject file path check before passing backend.
	extractedDir, err := pkg.ExtractFromZipReader(FromMobileFS(eragoFsys), bytes.NewReader(savZipBytes), int64(len(savZipBytes)))
	if err != nil {
		return fmt.Errorf("extract zip failed: %w", err)
	}

	// post condition check. Actually save files are extracted even if those are invalid path. May break installed package structure.
	absSavDir := filepath.Join(absEragoDir, appConf.Game.RepoConfig.SaveFileDir)
	if !strings.Contains(filepath.ToSlash(absSavDir), filepath.ToSlash(extractedDir)) {
		return fmt.Errorf("invalid sav directory for zip content, save dir = %v, extracted dir = %v", absSavDir, extractedDir)
	}
	return nil
}

// ExportLog exports log file with respect to erago directory. It returns log content as bytes and error if failed.
// If log file does not exist, it returns ErrNoLogFile.
func ExportLog(absEragoDir string, eragoFsys FileSystemGlob) ([]byte, error) {
	oldDefaultFS := filesystem.Default
	defer func() { filesystem.Default = oldDefaultFS }()

	filesystem.Default = wrapFileSystemPR{FromMobileFSGlob(eragoFsys)}

	appConf, err := app.LoadConfigOrDefault(app.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get save file directory from config: %w", err)
	}

	_, _ = disableDesktopFeatures(appConf)
	if !eragoFsys.Exist(appConf.LogFile) {
		// no log case is treated as empty log content and succeeded.
		return nil, fmt.Errorf("could not found log file %s: %w", appConf.LogFile, ErrNoLogFile)
	}

	reader, err := eragoFsys.Load(appConf.LogFile)
	if err != nil {
		return nil, fmt.Errorf("could not read %v: %w", appConf.LogFile, err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// MatchGlobPattern is helper function that checks whether glob pattern matches with path.
// It returns true if pattern matched, otherwise returns false.
func MatchGlobPattern(pattern, path string) (bool, error) {
	return filepath.Match(pattern, path)
}
