package fs_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/grokify/oscompat/fs"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		// Valid paths
		{"simple file", "file.txt", nil},
		{"nested path", "foo/bar/baz.txt", nil},
		{"with dots in name", "file.tar.gz", nil},
		{"current dir prefix", "./file.txt", nil},

		// Invalid paths
		{"empty", "", fs.ErrEmptyPath},
		{"parent traversal", "..", fs.ErrPathTraversal},
		{"parent traversal with path", "../file.txt", fs.ErrPathTraversal},
		{"deep traversal", "foo/../../bar", fs.ErrPathTraversal},
		{"absolute unix", "/etc/passwd", fs.ErrAbsolutePath},

		// Backslash handling (Windows-style)
		{"backslash path", `foo\bar\baz.txt`, nil},
		{"backslash traversal", `foo\..\..\..\bar`, fs.ErrPathTraversal},
	}

	// Add Windows-specific test
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name    string
			path    string
			wantErr error
		}{"windows absolute", `C:\Windows\System32`, fs.ErrAbsolutePath})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.ValidatePath(tt.path)
			if err != tt.wantErr {
				t.Errorf("ValidatePath(%q) = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePathStrict(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{"normal file", "file.txt", nil},
		{"nested path", "foo/bar/baz.txt", nil},
		{"hidden file", ".hidden", fs.ErrPathTraversal},
		{"hidden dir", ".config/file", fs.ErrPathTraversal},
		{"empty", "", fs.ErrEmptyPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.ValidatePathStrict(tt.path)
			if err != tt.wantErr {
				t.Errorf("ValidatePathStrict(%q) = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"simple", "file.txt", "file.txt"},
		{"with slashes", "foo/bar/baz", "foo/bar/baz"},
		{"with backslashes", `foo\bar\baz`, "foo/bar/baz"},
		{"mixed slashes", `foo/bar\baz`, "foo/bar/baz"},
		{"redundant slashes", "foo//bar///baz", "foo/bar/baz"},
		{"with dot", "./foo/bar", "foo/bar"},
		{"trailing slash", "foo/bar/", "foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.NormalizePath(tt.path)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestOSPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"empty", ""},
		{"simple", "file.txt"},
		{"nested", "foo/bar/baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.OSPath(tt.path)

			// Verify round-trip
			normalized := fs.NormalizePath(got)
			if normalized != tt.path {
				t.Errorf("Round-trip failed: %q -> %q -> %q", tt.path, got, normalized)
			}
		})
	}
}

func TestJoinNormalized(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		want  string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"foo"}, "foo"},
		{"multiple", []string{"foo", "bar", "baz"}, "foo/bar/baz"},
		{"with empty", []string{"foo", "", "bar"}, "foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.JoinNormalized(tt.elems...)
			if got != tt.want {
				t.Errorf("JoinNormalized(%v) = %q, want %q", tt.elems, got, tt.want)
			}
		})
	}
}

func TestJoinOS(t *testing.T) {
	elems := []string{"foo", "bar", "baz"}
	got := fs.JoinOS(elems...)

	// Should use OS separator
	expected := filepath.Join(elems...)
	if got != expected {
		t.Errorf("JoinOS(%v) = %q, want %q", elems, got, expected)
	}
}

func TestSafeJoin(t *testing.T) {
	base := t.TempDir()

	tests := []struct {
		name    string
		rel     string
		wantErr bool
	}{
		{"simple", "file.txt", false},
		{"nested", "foo/bar/baz.txt", false},
		{"traversal", "../outside.txt", true},
		{"deep traversal", "foo/../../outside.txt", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fs.SafeJoin(base, tt.rel)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeJoin(%q, %q) error = %v, wantErr %v", base, tt.rel, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("SafeJoin returned empty path for valid input")
			}
		})
	}
}

func TestMkdirAll(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "test", "nested", "dir")

	err := fs.MkdirAll(dir, 0)
	if err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll did not create a directory")
	}
}

func TestMkdirAllPrivate(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "private", "dir")

	err := fs.MkdirAllPrivate(dir)
	if err != nil {
		t.Fatalf("MkdirAllPrivate() error: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAllPrivate did not create a directory")
	}

	// On Unix, verify permissions
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != fs.PrivateDirPerm {
			t.Errorf("MkdirAllPrivate() perm = %o, want %o", perm, fs.PrivateDirPerm)
		}
	}
}

func TestWriteFile(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "test.txt")
	data := []byte("hello world")

	err := fs.WriteFile(file, data, 0)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("WriteFile() content = %q, want %q", got, data)
	}
}

func TestWriteFilePrivate(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "private.txt")
	data := []byte("secret")

	err := fs.WriteFilePrivate(file, data)
	if err != nil {
		t.Fatalf("WriteFilePrivate() error: %v", err)
	}

	info, err := os.Stat(file)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	// On Unix, verify permissions
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != fs.PrivateFilePerm {
			t.Errorf("WriteFilePrivate() perm = %o, want %o", perm, fs.PrivateFilePerm)
		}
	}
}

func TestPermissionConstants(t *testing.T) {
	// Verify permission constants have expected values
	if fs.DefaultDirPerm != 0755 {
		t.Errorf("DefaultDirPerm = %o, want 0755", fs.DefaultDirPerm)
	}
	if fs.DefaultFilePerm != 0644 {
		t.Errorf("DefaultFilePerm = %o, want 0644", fs.DefaultFilePerm)
	}
	if fs.PrivateDirPerm != 0700 {
		t.Errorf("PrivateDirPerm = %o, want 0700", fs.PrivateDirPerm)
	}
	if fs.PrivateFilePerm != 0600 {
		t.Errorf("PrivateFilePerm = %o, want 0600", fs.PrivateFilePerm)
	}
	if fs.ExecutablePerm != 0755 {
		t.Errorf("ExecutablePerm = %o, want 0755", fs.ExecutablePerm)
	}
}

func TestIsCaseSensitive(t *testing.T) {
	// On Windows, should be false; on Unix, should be true
	got := fs.IsCaseSensitive()
	if runtime.GOOS == "windows" {
		if got {
			t.Error("IsCaseSensitive() on Windows should be false")
		}
	} else {
		if !got {
			t.Error("IsCaseSensitive() on Unix should be true")
		}
	}
}

func TestPathEqual(t *testing.T) {
	tests := []struct {
		name  string
		path1 string
		path2 string
		want  bool
	}{
		{"identical", "foo/bar", "foo/bar", true},
		{"backslash vs slash", `foo\bar`, "foo/bar", true},
		{"redundant slashes", "foo//bar", "foo/bar", true},
		{"different paths", "foo/bar", "foo/baz", false},
	}

	// Add case-sensitivity test based on platform
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name  string
			path1 string
			path2 string
			want  bool
		}{"case difference (Windows)", "Foo/Bar", "foo/bar", true})
	} else {
		tests = append(tests, struct {
			name  string
			path1 string
			path2 string
			want  bool
		}{"case difference (Unix)", "Foo/Bar", "foo/bar", false})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.PathEqual(tt.path1, tt.path2)
			if got != tt.want {
				t.Errorf("PathEqual(%q, %q) = %v, want %v", tt.path1, tt.path2, got, tt.want)
			}
		})
	}
}

func TestPathHasPrefix(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{"exact prefix", "foo/bar/baz", "foo/bar", true},
		{"root prefix", "foo/bar/baz", "foo", true},
		{"no prefix", "foo/bar", "baz", false},
		{"partial name match", "foobar/baz", "foo", false}, // "foo" is not a directory prefix
		{"empty prefix", "foo/bar", "", true},
		{"backslash prefix", `foo\bar\baz`, "foo/bar", true},
	}

	// Add case-sensitivity test based on platform
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name   string
			path   string
			prefix string
			want   bool
		}{"case difference (Windows)", "Foo/Bar/Baz", "foo/bar", true})
	} else {
		tests = append(tests, struct {
			name   string
			path   string
			prefix string
			want   bool
		}{"case difference (Unix)", "Foo/Bar/Baz", "foo/bar", false})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.PathHasPrefix(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("PathHasPrefix(%q, %q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestBirthtimeSupported(t *testing.T) {
	supported := fs.BirthtimeSupported()

	// Verify expected behavior per platform
	switch runtime.GOOS {
	case "darwin", "windows":
		if !supported {
			t.Errorf("BirthtimeSupported() on %s should be true", runtime.GOOS)
		}
	case "linux":
		if supported {
			t.Errorf("BirthtimeSupported() on Linux should be false")
		}
	}
	// For other platforms, we don't assert - just verify it returns a bool
}

func TestBirthtime(t *testing.T) {
	// Create a temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Get birthtime
	bt, err := fs.Birthtime(tmpFile)
	if err != nil {
		t.Fatalf("Birthtime() error: %v", err)
	}

	// Verify it's a reasonable time (not zero, not in the future)
	if bt.IsZero() {
		t.Error("Birthtime() returned zero time")
	}
	if bt.After(time.Now().Add(time.Second)) {
		t.Error("Birthtime() returned time in the future")
	}

	// Verify it's not unreasonably old (within last hour)
	if bt.Before(time.Now().Add(-time.Hour)) {
		t.Error("Birthtime() returned time more than an hour ago for new file")
	}
}

func TestBirthtimeInfo(t *testing.T) {
	// Create a temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Get FileInfo
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	// Get birthtime from FileInfo
	bt := fs.BirthtimeInfo(info)

	// Verify it's a reasonable time
	if bt.IsZero() {
		t.Error("BirthtimeInfo() returned zero time")
	}
	if bt.After(time.Now().Add(time.Second)) {
		t.Error("BirthtimeInfo() returned time in the future")
	}
}

func TestBirthtimeNonExistent(t *testing.T) {
	_, err := fs.Birthtime("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Error("Birthtime() on non-existent file should return error")
	}
}
