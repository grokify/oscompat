package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grokify/oscompat/fs"
)

// ReadFileLF reads a text file and converts CRLF line endings to LF, so a
// fixture compares equal whether git checked it out with LF or, on Windows,
// with CRLF. It fails the test if the file cannot be read.
func ReadFileLF(t testing.TB, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return fs.NormalizeLineEndingsBytes(data)
}

// AssertGolden fails the test unless got equals the golden file at path,
// ignoring CRLF versus LF in the file. With update set it writes got to path
// instead, creating the directory if needed; wire update to a test flag:
//
//	var update = flag.Bool("update", false, "rewrite the golden files")
//
//	testutil.AssertGolden(t, "testdata/out.json", got, *update)
//
// got is compared as given, so produce it with LF line endings.
func AssertGolden(t testing.TB, path string, got []byte, update bool) {
	t.Helper()
	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want := ReadFileLF(t, path)
	if string(want) != string(got) {
		t.Fatalf("%s differs from the golden file\n--- want\n%s\n--- got\n%s", path, want, got)
	}
}
