package filesystem

import (
	"path/filepath"
	"testing"
)

func TestGlob(t *testing.T) {
	const GoPattern = "*.go"
	goFiles, err := filepath.Glob(GoPattern)
	if err != nil {
		t.Fatal(err)
	}

	fsGoFiles, err := Glob(GoPattern)
	if err != nil {
		t.Fatal(err)
	}

	if len(fsGoFiles) != len(goFiles) {
		t.Fatalf("differenct glob result len, got %v, expect %v", len(fsGoFiles), len(goFiles))
	}

	for i, f := range goFiles {
		if fsGoFiles[i] != f {
			t.Fatalf("different glob file, got %s, expect %s", fsGoFiles[i], f)
		}
	}

	backupDefault := Default
	defer func() { Default = backupDefault }()

	notFoundAbsPath, err := filepath.Abs("./notfound_dir")
	if err != nil {
		t.Fatal(err)
	}
	absPathFS := &AbsPathFileSystem{CurrentDir: notFoundAbsPath}
	Default = absPathFS

	mobileFsGoFiles, err := Glob(GoPattern)
	if err != nil {
		t.Fatal(err)
	}

	if len(mobileFsGoFiles) > 0 {
		t.Fatalf("given missing pattern, but Glob returns some result: %q", mobileFsGoFiles)
	}
}

func TestResolvePath(t *testing.T) {
	const testPath = "path/to/notfound"
	gotPath, err := ResolvePath(testPath)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != testPath {
		t.Errorf("differenct resolved path: got %v, expect %v", gotPath, testPath)
	}

	backupDefault := Default
	defer func() { Default = backupDefault }()

	absPathFS := &AbsPathFileSystem{CurrentDir: ""}
	Default = absPathFS

	testAbsPath, err := filepath.Abs(testPath)
	if err != nil {
		t.Fatal(err)
	}
	gotAbsPath, err := ResolvePath(testPath)
	if err != nil {
		t.Fatal(err)
	}
	if gotAbsPath != testAbsPath {
		t.Errorf("differenct resolved abs path: got %v, expect %v", gotAbsPath, testAbsPath)
	}
}
