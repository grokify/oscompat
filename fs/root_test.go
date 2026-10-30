package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newRoot(t *testing.T) string {
	t.Helper()
	root, err := RealDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestResolveInRootAccepts(t *testing.T) {
	root := newRoot(t)
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		"new.yaml",
		"sub/new.yaml",
		"missing/dir/new.yaml", // the directory need not exist
		"./dot.yaml",
		filepath.Join(root, "abs.yaml"), // absolute, inside the root
		".",
	} {
		got, err := ResolveInRoot(root, p)
		if err != nil {
			t.Errorf("ResolveInRoot(%q): %v", p, err)
			continue
		}
		if !within(root, got) {
			t.Errorf("ResolveInRoot(%q) = %q, outside %q", p, got, root)
		}
	}
}

// These are refused the same way on every platform, including the ones that
// Windows would otherwise read as relative to the root.
func TestResolveInRootRefuses(t *testing.T) {
	root := newRoot(t)
	for _, p := range []string{
		"", "..", "../x", "a/../../b", `a\..\b`, "sub/../../x",
		"/etc/passwd", `\etc\passwd`, `\\server\share\x`,
		"C:x.yaml", `C:\x.yaml`, "C:/x.yaml",
		"a\x00b",
		filepath.Join(filepath.Dir(root), "elsewhere.yaml"), // absolute, outside
	} {
		_, err := ResolveInRoot(root, p)
		if !errors.Is(err, ErrOutsideRoot) {
			t.Errorf("ResolveInRoot(%q) = %v, want ErrOutsideRoot", p, err)
		}
	}
}

func TestResolveInRootRefusesLinksOutside(t *testing.T) {
	root := newRoot(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("cannot create symbolic links here: %v", err) // Windows needs a privilege
	}
	for _, p := range []string{"link", "link/secret", "link/new.yaml", "link/missing/new.yaml"} {
		if _, err := ResolveInRoot(root, p); !errors.Is(err, ErrOutsideRoot) {
			t.Errorf("ResolveInRoot(%q) = %v, want ErrOutsideRoot", p, err)
		}
	}
	// A link that stays inside the root is fine, and the real path is returned.
	if err := os.Mkdir(filepath.Join(root, "real"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveInRoot(root, "alias/f.yaml")
	if err != nil || got != filepath.Join(root, "real", "f.yaml") {
		t.Errorf("got %q, %v; want the real path inside the root", got, err)
	}
}

func TestRealDir(t *testing.T) {
	dir := t.TempDir()
	if got, err := RealDir(dir); err != nil || !filepath.IsAbs(got) {
		t.Errorf("RealDir(%q) = %q, %v", dir, got, err)
	}
	if _, err := RealDir(filepath.Join(dir, "missing")); err == nil {
		t.Error("a directory that does not exist was accepted")
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RealDir(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("a file was accepted as a directory: %v", err)
	}
}
