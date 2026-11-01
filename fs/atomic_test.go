package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // G304: a file in the test's temporary directory.
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// onlyFile checks dir holds just name: no temporary file is left behind.
func onlyFile(t *testing.T, dir, name string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != name {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory holds %v, want only %s", names, name)
	}
}

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fi.Mode().Perm()
}

func TestWriteFileAtomicCreatesAndReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := WriteFileAtomic(path, []byte("one"), 0); err != nil {
		t.Fatal(err)
	}
	if got := read(t, path); got != "one" {
		t.Errorf("content = %q", got)
	}
	if PermissionsEnforced() && mode(t, path) != DefaultFilePerm {
		t.Errorf("mode = %v, want %v", mode(t, path), DefaultFilePerm)
	}
	if err := WriteFileAtomic(path, []byte("two"), 0o640); err != nil {
		t.Fatal(err)
	}
	if got := read(t, path); got != "two" {
		t.Errorf("content = %q", got)
	}
	if PermissionsEnforced() && mode(t, path) != 0o640 {
		t.Errorf("mode = %v, want 0640", mode(t, path))
	}
	onlyFile(t, dir, "registry.json")
}

func TestWriteFileAtomicPrivate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.json")
	if err := WriteFileAtomicPrivate(path, []byte("x")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPrivate(fi); err != nil {
		t.Error(err)
	}
}

func TestWriteFileAtomicFailsWithoutTheDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "file.json")
	if err := WriteFileAtomic(path, []byte("x"), 0); err == nil {
		t.Error("wrote into a directory that does not exist")
	}
}

func TestWriteFileAtomicLeavesTheOldFileWhenTheRenameFails(t *testing.T) {
	dir := t.TempDir()
	// A directory at the path makes the rename fail on every platform.
	path := filepath.Join(dir, "target")
	if err := os.Mkdir(path, DefaultDirPerm); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("x"), 0); err == nil {
		t.Fatal("replaced a directory")
	}
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		t.Errorf("the directory changed: %v, %v", fi, err)
	}
	onlyFile(t, dir, "target")
}
