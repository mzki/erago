package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mzki/erago/app/config"
	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/infra/pkg"
	"github.com/mzki/erago/state/csv"
	"github.com/mzki/erago/util/log"
)

// Packaging package erago related files ELA/*, CSV/*, appConf file, and given extra files on appConf context.
// The packaging result is stored under dstDir. File name is {{csv.GameBase.Title}}-{{csv.GameBase.Version}}.zip
// If same file name already exist, packaging will fail.
// It returns whether operation succeeded or not. internal error is handled by itself.
func Packaging(dstDir string, appConf *config.Config, appConfPath string, extraFiles []string) bool {
	if appConf == nil {
		panic("appConf should not be nil")
	}

	// returned value must be called once.
	reset, err := config.SetupLogConfig(appConf)
	if err != nil {
		// TODO: what is better way to handle fatal error in this case?
		fmt.Fprintf(os.Stderr, "log configuration failed: %v\n", err)
		return false
	}
	defer reset()

	// prepare source files
	targetFiles := make([]string, 0, 8)
	targetFiles = append(targetFiles, appConfPath)
	targetFiles = append(targetFiles, extraFiles...)

	targetDirs := []string{
		appConf.Game.CSVConfig.Dir,
		appConf.Game.ScriptConfig.LoadDir,
	}
	for _, dir := range targetDirs {
		files := pkg.CollectFiles(os.DirFS(dir).(fs.ReadDirFS), ".")
		for i, f := range files {
			files[i] = filepath.Join(dir, f)
		}
		targetFiles = append(targetFiles, files...)
	}

	// prepare destination
	absDstDir, err := filepath.Abs(dstDir)
	if err != nil {
		log.Infof("failed to get absolute directory path for: %v", dstDir)
		return false
	}
	dstFsys := filesystem.AbsDirFileSystem(absDstDir)

	// create package name
	csvM := csv.NewCsvManager()
	if err := csvM.Initialize(appConf.Game.CSVConfig); err != nil {
		log.Infof("CSV Initialization failed: %v", err)
		return false
	}
	archiveName := fmt.Sprintf("%v-%v.zip", csvM.GameBase.Title, csvM.GameBase.Version)

	if dstFsys.Exist(archiveName) {
		outputPath := filepath.Join(dstDir, archiveName)
		log.Infof("%v", &fs.PathError{Op: "create", Path: outputPath, Err: fmt.Errorf("already exist")})
		return false
	}

	if outputFile, err := pkg.ArchiveAsZip(dstFsys, archiveName, filesystem.Desktop, targetFiles); err != nil {
		log.Infof("output as Zip failed: %v", err)
		return false
	} else {
		log.Infof("output Zip archive: %v", outputFile)
		return true
	}
}
