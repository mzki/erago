package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	var testPath = filepath.Clean("path/to/notfound")
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

func TestOpenWatcher(t *testing.T) {
	const testPath = "./pathToRemoved"
	cleanTestPath := filepath.Clean(testPath)
	watcher, err := OpenWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	removeCh := make(chan bool)
	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events():
				t.Log(ev)
				if !ok {
					return
				}
				cleanName := filepath.Clean(ev.Name)
				matched, err := filepath.Match(cleanTestPath, cleanName)
				if err != nil {
					t.Error(err)
				} else if matched && ev.Has(WatchOpRemove) {
					removeCh <- true
				}
			case err, ok := <-watcher.Errors():
				if !ok {
					return
				}
				t.Error(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	// create and remove testPath
	fp, err := os.Create(cleanTestPath)
	if err != nil {
		t.Fatal(err)
	}
	fp.Close()

	// Watch accpet un-Cleaned path
	if err := watcher.Watch(testPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(cleanTestPath); err != nil {
		t.Fatal(err)
	}

	var removed bool = false
	select {
	case <-ctx.Done():
		t.Fatal("Failed to receive Remove Event from Watcher", ctx.Err())
	case <-removeCh:
		removed = true
	}

	if !removed {
		t.Errorf("Expect receive Remove Event: %v, but got: %v", true, removed)
	}
}
