package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

func TestOpenRegularInRootAccepts(t *testing.T) {
	root := newRoot(t)
	writeFile(t, filepath.Join(root, "five.txt"), "12345")
	writeFile(t, filepath.Join(root, "empty.txt"), "")
	for _, c := range []struct {
		name string
		max  int64
	}{
		{"five.txt", 5}, // exactly at the limit
		{"five.txt", 50},
		{"empty.txt", 0},
		{"empty.txt", -1}, // a negative limit is zero
	} {
		f, err := OpenRegularInRoot(root, c.name, c.max)
		if err != nil {
			t.Errorf("OpenRegularInRoot(%q, %d): %v", c.name, c.max, err)
			continue
		}
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}
}

func TestOpenRegularInRootRefuses(t *testing.T) {
	root := newRoot(t)
	writeFile(t, filepath.Join(root, "five.txt"), "12345")
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o750); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		path string
		max  int64
		want error
	}{
		{"one byte over", "five.txt", 4, ErrTooLarge},
		{"any byte over a negative limit", "five.txt", -1, ErrTooLarge},
		{"a directory", "dir", 100, ErrNotRegular},
		{"the root itself", ".", 100, ErrNotRegular},
		{"outside the root", "../x", 100, ErrOutsideRoot},
		{"missing", "missing.txt", 100, os.ErrNotExist},
	} {
		t.Run(c.name, func(t *testing.T) {
			f, err := OpenRegularInRoot(root, c.path, c.max)
			if f != nil {
				if cerr := f.Close(); cerr != nil {
					t.Error(cerr)
				}
			}
			if !errors.Is(err, c.want) || f != nil {
				t.Errorf("OpenRegularInRoot(%q, %d) = %v, %v; want %v and no file", c.path, c.max, f, err, c.want)
			}
		})
	}
}

// A link out of the root is refused before the handle is checked, so the size
// and type of what it points at are never consulted.
func TestOpenRegularInRootRefusesLinksOutside(t *testing.T) {
	root := newRoot(t)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret"), "S")
	if err := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(root, "link")); err != nil {
		t.Skipf("cannot create symbolic links here: %v", err)
	}
	if _, err := OpenRegularInRoot(root, "link", 100); !errors.Is(err, ErrOutsideRoot) {
		t.Errorf("a link out of the root gave %v; want ErrOutsideRoot", err)
	}
}

func TestCapReader(t *testing.T) {
	for _, c := range []struct {
		name    string
		data    string
		max     int64
		wantErr bool
	}{
		{"under the limit", "abc", 10, false},
		{"exactly at the limit", "abcde", 5, false},
		{"one byte over", "abcdef", 5, true},
		{"far over", strings.Repeat("x", 1000), 5, true},
		{"empty at a zero limit", "", 0, false},
		{"one byte at a zero limit", "a", 0, true},
		{"one byte at a negative limit", "a", -3, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := io.ReadAll(CapReader(strings.NewReader(c.data), c.max))
			if c.wantErr {
				if !errors.Is(err, ErrTooLarge) {
					t.Fatalf("err = %v, want ErrTooLarge", err)
				}
				// The bytes within the limit are delivered before the error.
				if want := c.data[:min(len(c.data), int(max(c.max, 0)))]; string(got) != want {
					t.Errorf("read %q before the error, want %q", got, want)
				}
				return
			}
			if err != nil || string(got) != c.data {
				t.Errorf("read %q, %v; want %q and no error", got, err, c.data)
			}
		})
	}
}

// The limit does not depend on how the reads are sliced.
func TestCapReaderInAnyReadSize(t *testing.T) {
	data := strings.Repeat("0123456789", 10) // 100 bytes
	wrappers := map[string]func(io.Reader) io.Reader{
		"one byte at a time": iotest.OneByteReader,
		"half reads":         iotest.HalfReader,
		"data with the EOF":  iotest.DataErrReader,
	}
	for name, wrap := range wrappers {
		t.Run(name, func(t *testing.T) {
			got, err := io.ReadAll(CapReader(wrap(strings.NewReader(data)), 100))
			if err != nil || string(got) != data {
				t.Errorf("at the limit: read %d bytes, %v", len(got), err)
			}
			got, err = io.ReadAll(CapReader(wrap(strings.NewReader(data+"!")), 100))
			if !errors.Is(err, ErrTooLarge) || len(got) != 100 {
				t.Errorf("one over: read %d bytes, %v; want 100 and ErrTooLarge", len(got), err)
			}
		})
	}
}

func TestCapReaderErrorIsStickyAndOtherErrorsPass(t *testing.T) {
	r := CapReader(strings.NewReader("abcdef"), 3)
	if _, err := io.ReadAll(r); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
	if n, err := r.Read(make([]byte, 8)); n != 0 || !errors.Is(err, ErrTooLarge) {
		t.Errorf("a later read gave %d, %v; want 0 and ErrTooLarge", n, err)
	}
	boom := errors.New("boom")
	if _, err := io.ReadAll(CapReader(iotest.ErrReader(boom), 10)); !errors.Is(err, boom) {
		t.Errorf("a reader error gave %v, want it passed through", err)
	}
}

// The reason for the cap: a file that passed the size check can grow before it
// is read.
func TestCapReaderStopsAFileThatGrows(t *testing.T) {
	root := newRoot(t)
	path := filepath.Join(root, "grows.txt")
	writeFile(t, path, "abc")
	f, err := OpenRegularInRoot(root, "grows.txt", 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	grow, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := grow.WriteString(strings.Repeat("x", 100)); err != nil {
		t.Fatal(err)
	}
	if err := grow.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(CapReader(f, 5)); !errors.Is(err, ErrTooLarge) {
		t.Errorf("a file that grew past the limit gave %v; want ErrTooLarge", err)
	}
}
