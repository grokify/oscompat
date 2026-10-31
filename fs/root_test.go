package fs

import (
	"errors"
	"io"
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readAll(t *testing.T, f *os.File) string {
	t.Helper()
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOpenInRootOpens(t *testing.T) {
	root := newRoot(t)
	writeFile(t, filepath.Join(root, "a.txt"), "A")
	writeFile(t, filepath.Join(root, "sub", "b.txt"), "B")
	for p, want := range map[string]string{
		"a.txt":                             "A",
		"./a.txt":                           "A",
		"sub/b.txt":                         "B",
		filepath.Join(root, "sub", "b.txt"): "B", // absolute, inside the root
	} {
		f, err := OpenInRoot(root, p)
		if err != nil {
			t.Errorf("OpenInRoot(%q): %v", p, err)
			continue
		}
		if got := readAll(t, f); got != want {
			t.Errorf("OpenInRoot(%q) read %q, want %q", p, got, want)
		}
	}
}

// The same names are refused as by ResolveInRoot, on every platform.
func TestOpenInRootRefuses(t *testing.T) {
	root := newRoot(t)
	writeFile(t, filepath.Join(filepath.Dir(root), "elsewhere.yaml"), "x")
	t.Cleanup(func() {
		if err := os.Remove(filepath.Join(filepath.Dir(root), "elsewhere.yaml")); err != nil {
			t.Error(err)
		}
	})
	for _, p := range []string{
		"", "..", "../x", "a/../../b", `a\..\b`, "sub/../../x",
		"/etc/passwd", `\etc\passwd`, `\\server\share\x`,
		"C:x.yaml", `C:\x.yaml`, "C:/x.yaml",
		"a\x00b",
		filepath.Join(filepath.Dir(root), "elsewhere.yaml"), // absolute, outside
	} {
		f, err := OpenInRoot(root, p)
		if f != nil {
			if cerr := f.Close(); cerr != nil {
				t.Error(cerr)
			}
		}
		if !errors.Is(err, ErrOutsideRoot) || f != nil {
			t.Errorf("OpenInRoot(%q) = %v, %v; want ErrOutsideRoot and no file", p, f, err)
		}
	}
}

// A link out of the root is refused by the open itself, in any element of the
// name. The wording check keeps the match on os.Root's text honest: if the
// standard library changes it, the link cases fail here, not in a caller.
func TestOpenInRootRefusesLinksOutside(t *testing.T) {
	root := newRoot(t)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret"), "SECRET")
	if err := os.Symlink(outside, filepath.Join(root, "dirlink")); err != nil {
		t.Skipf("cannot create symbolic links here: %v", err) // Windows needs a privilege
	}
	if err := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(root, "filelink")); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"dirlink/secret", "filelink", "dirlink"} {
		f, err := OpenInRoot(root, p)
		if f != nil {
			if cerr := f.Close(); cerr != nil {
				t.Error(cerr)
			}
		}
		if !errors.Is(err, ErrOutsideRoot) || f != nil {
			t.Errorf("OpenInRoot(%q) = %v, %v; want ErrOutsideRoot and no file", p, f, err)
		}
	}

	// A link with a relative target that stays inside the root is followed.
	writeFile(t, filepath.Join(root, "real", "f.txt"), "REAL")
	if err := os.Symlink("real", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	f, err := OpenInRoot(root, "alias/f.txt")
	if err != nil {
		t.Fatalf("a relative link inside the root was refused: %v", err)
	}
	if got := readAll(t, f); got != "REAL" {
		t.Errorf("read %q through an inside link, want REAL", got)
	}
}

// os.Root does not follow a link whose target is an absolute path, even one
// that leads back into the root. ResolveInRoot accepts such a link, so the two
// differ here, and the difference is documented.
func TestOpenInRootRefusesAbsoluteLinks(t *testing.T) {
	root := newRoot(t)
	writeFile(t, filepath.Join(root, "real", "f.txt"), "REAL")
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "abs")); err != nil {
		t.Skipf("cannot create symbolic links here: %v", err)
	}
	f, err := OpenInRoot(root, "abs/f.txt")
	if f != nil {
		if cerr := f.Close(); cerr != nil {
			t.Error(cerr)
		}
	}
	if !errors.Is(err, ErrOutsideRoot) {
		t.Errorf("a link with an absolute target gave %v; want ErrOutsideRoot", err)
	}
	if _, err := ResolveInRoot(root, "abs/f.txt"); err != nil {
		t.Errorf("ResolveInRoot refused the same link: %v", err)
	}
}

func TestOpenInRootErrors(t *testing.T) {
	root := newRoot(t)
	if _, err := OpenInRoot(root, "missing.txt"); !errors.Is(err, os.ErrNotExist) || errors.Is(err, ErrOutsideRoot) {
		t.Errorf("a missing file gave %v; want os.ErrNotExist and not ErrOutsideRoot", err)
	}
	if _, err := OpenInRoot(filepath.Join(root, "no-such-root"), "x"); err == nil || errors.Is(err, ErrOutsideRoot) {
		t.Errorf("a root that does not exist gave %v; want an error that is not ErrOutsideRoot", err)
	}
}

// The handle stays bound to the file that was opened, so replacing the path
// afterwards, here with a link to a file outside the root, changes nothing
// about what is read.
func TestOpenInRootHandleKeepsItsFile(t *testing.T) {
	root := newRoot(t)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret"), "SECRET")
	path := filepath.Join(root, "f.txt")
	writeFile(t, path, "ORIGINAL")

	f, err := OpenInRoot(root, "f.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		if cerr := f.Close(); cerr != nil {
			t.Error(cerr)
		}
		t.Skipf("cannot replace an open file here: %v", err) // Windows refuses
	}
	if err := os.Symlink(filepath.Join(outside, "secret"), path); err != nil {
		if cerr := f.Close(); cerr != nil {
			t.Error(cerr)
		}
		t.Skipf("cannot create symbolic links here: %v", err)
	}
	if got := readAll(t, f); got != "ORIGINAL" {
		t.Errorf("the handle read %q after the path was replaced, want ORIGINAL", got)
	}
}
