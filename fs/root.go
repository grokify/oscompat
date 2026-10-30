package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrOutsideRoot is returned for a path that leaves its root: it climbs out
// with "..", is rooted or names a drive outside the root, contains a NUL, or
// leads out through a symbolic link.
var ErrOutsideRoot = errors.New("oscompat/fs: path is outside the root")

// RealDir returns dir as an absolute path with symbolic links resolved. It
// returns an error if dir does not exist or is not a directory.
//
// Resolve a root once with RealDir and keep the result, since ResolveInRoot
// compares real paths.
func RealDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("oscompat/fs: %q: %w", dir, err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("oscompat/fs: %q: %w", dir, err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", fmt.Errorf("oscompat/fs: %q: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("oscompat/fs: %q is not a directory", dir)
	}
	return real, nil
}

// ResolveInRoot turns a path supplied by a caller, such as a tool argument or
// an entry in a document, into a real path inside root, or returns an error
// wrapping ErrOutsideRoot.
//
// A relative path is taken from root. An absolute path is accepted only if it
// lies inside root. These are refused on every platform alike:
//
//   - an empty path or a path containing a NUL;
//   - any ".." element, with either / or \ as the separator;
//   - a path that is rooted or names a drive but is not absolute on this
//     platform ("/etc/passwd" or "C:x" on Windows), which would otherwise be
//     read as relative to root (see IsAbsAny);
//   - a path that, once symbolic links are followed, lies outside root.
//
// The file need not exist, but the part of the path that does exist is
// resolved through links, so a caller can create a file in place. root must
// be an existing directory; pass the result of RealDir.
//
// The check runs when the path is resolved. A link created afterwards is not
// seen, so open the returned path promptly and do not treat it as a lock.
func ResolveInRoot(root, p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("%w: empty path", ErrOutsideRoot)
	}
	if strings.ContainsRune(p, 0) {
		return "", fmt.Errorf("%w: %q", ErrOutsideRoot, p)
	}
	for _, part := range strings.FieldsFunc(p, func(c rune) bool { return c == '/' || c == '\\' }) {
		if part == ".." {
			return "", fmt.Errorf("%w: %q has a \"..\" element", ErrOutsideRoot, p)
		}
	}
	if IsAbsAny(p) && !filepath.IsAbs(p) {
		return "", fmt.Errorf("%w: %q is not a path inside the root", ErrOutsideRoot, p)
	}
	full := p
	if !filepath.IsAbs(full) {
		full = filepath.Join(root, full)
	}
	full = filepath.Clean(full)

	// Follow links for the longest part that exists, then re-attach the rest.
	existing, rest := full, ""
	for {
		real, err := filepath.EvalSymlinks(existing)
		if err == nil {
			resolved := filepath.Join(real, rest)
			if !within(root, resolved) {
				return "", fmt.Errorf("%w: %q", ErrOutsideRoot, p)
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("oscompat/fs: resolve %q: %w", p, err)
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", fmt.Errorf("%w: %q", ErrOutsideRoot, p)
		}
		rest = filepath.Join(filepath.Base(existing), rest)
		existing = parent
	}
}

// within reports whether path is root or lies below it. Both must be clean,
// absolute paths on the same footing (both real, or both not).
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
