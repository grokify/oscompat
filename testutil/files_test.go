package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grokify/oscompat/testutil"
)

func TestReadFileLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(path, []byte("a\r\nb\nc\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := string(testutil.ReadFileLF(t, path)); got != "a\nb\nc\n" {
		t.Fatalf("got %q", got)
	}
}

func TestAssertGoldenIgnoresCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "g.json")
	if err := os.WriteFile(path, []byte("{\r\n  \"a\": 1\r\n}\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, path, []byte("{\n  \"a\": 1\n}\n"), false)
}

func TestAssertGoldenUpdateWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new", "g.txt")
	testutil.AssertGolden(t, path, []byte("x\n"), true)
	testutil.AssertGolden(t, path, []byte("x\n"), false)
}
