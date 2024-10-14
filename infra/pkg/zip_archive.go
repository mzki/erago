package pkg

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mzki/erago/filesystem"
)

// ArchiveAsZip archive targetFiles into zip.
// It returns output file path for the zip archive.
// The archive result will be output to outPath.
// It also returns error as 2nd result if something failed, otherwise returns nil.
// The outPath is with respect to outFsys and targetFiles are with respect to srcFsys respectively.
func ArchiveAsZip(outFsys filesystem.FileSystem, outPath string, srcFsys fs.FS, targetFiles []string) (outputPath string, err error) {
	// decide output path for zip archive and its name.
	var archiveBaseName string
	if _, base := filepath.Split(outPath); len(base) == 0 {
		return "", fmt.Errorf("empty base name is not allowed for output path: %v", outPath)
	} else {
		outputPath = filepath.Clean(outPath)
		archiveBaseName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// need to use github.com/mzki/erago/filesystem.FileSystem here to write file,
	// golang fs does not support to write files.
	outputFile, err := outFsys.Store(outputPath)
	if err != nil {
		return "", fmt.Errorf("could not open output file: %v", outputPath)
	}
	defer outputFile.Close()

	err = ArchiveAsZipWriter(outputFile, archiveBaseName, srcFsys, targetFiles)
	return
}

// ArchiveAsZipWriter is alternative API with io.Writer for ArchiveAsZip.
// See ArchiveAsZip documentation for the details.
func ArchiveAsZipWriter(w io.Writer, archiveBaseName string, srcFsys fs.FS, targetFiles []string) (err error) {
	zWriter := zip.NewWriter(w)
	defer func() {
		closeErr := zWriter.Close()
		err = errors.Join(err, closeErr)
	}()

	// align path separator to "/" for zip archive spec.
	for i, targetF := range targetFiles {
		targetFiles[i] = filepath.ToSlash(targetF)
	}
	for _, targetF := range targetFiles {
		// use closure for deferring Close inside each iteration.
		err = func() error {
			srcFile, err := srcFsys.Open(targetF)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			if err := addFileToZipWriter(zWriter, srcFile, archiveBaseName, targetF); err != nil {
				return fmt.Errorf("failed to add %v into zip: %w", targetF, err)
			}
			return nil
		}()
		if err != nil {
			return
		}
	}
	return
}

func addFileToZipWriter(zWriter *zip.Writer, srcFile fs.File, archiveBaseName string, filePath string) error {
	finfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(finfo)
	if err != nil {
		return err
	}
	header.Name = filepath.Join(archiveBaseName, filePath)
	header.Method = zip.Deflate
	// check wthether potential of zip slip
	if !strings.HasPrefix(header.Name, archiveBaseName) {
		return fmt.Errorf("potentially zip slip. archive file (%v) must be under %v, but points upper directory", header.Name, archiveBaseName)
	}

	w, err := zWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	return copyLimited(w, header.Name, srcFile, filePath)
}

// CollectFiles is a helper function that collect files under relDir recursively from filesyste fsys.
// files are relative path from relDir, and relDir is relative path from fsys respectively.
// If user wants to collect files from the root of fsys, relDir can be ".".
func CollectFiles(fsys fs.ReadDirFS, relDir string) (files []string) {
	files = make([]string, 0, 8)
	fs.WalkDir(fsys, relDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to search files in %v: %w", relDir, err)
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}
