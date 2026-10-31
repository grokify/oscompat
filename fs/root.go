package fs

import (
	"errors"
	"fmt"
	"io/fs"
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
// seen, so open the returned path promptly and do not treat it as a lock. To
// read a file, use OpenInRoot, which confines the open itself.
func ResolveInRoot(root, p string) (string, error) {
	if err := checkName(p); err != nil {
		return "", err
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

// checkName refuses the names that leave a root on any platform, before the
// file system is consulted: an empty name, a NUL, a ".." element with either
// separator, and a name that is rooted or names a drive but is not absolute
// here (see IsAbsAny).
func checkName(p string) error {
	if p == "" {
		return fmt.Errorf("%w: empty path", ErrOutsideRoot)
	}
	if strings.ContainsRune(p, 0) {
		return fmt.Errorf("%w: %q", ErrOutsideRoot, p)
	}
	for _, part := range strings.FieldsFunc(p, func(c rune) bool { return c == '/' || c == '\\' }) {
		if part == ".." {
			return fmt.Errorf("%w: %q has a \"..\" element", ErrOutsideRoot, p)
		}
	}
	if IsAbsAny(p) && !filepath.IsAbs(p) {
		return fmt.Errorf("%w: %q is not a path inside the root", ErrOutsideRoot, p)
	}
	return nil
}

// rootEscapeText is the text os.Root puts in its error for a name that leaves
// the root, for instance through a symbolic link. The standard library does not
// export the error, so it is matched by text; a test pins the wording.
const rootEscapeText = "path escapes from parent"

// isRootEscape reports whether err is os.Root refusing a name that leaves the
// root.
func isRootEscape(err error) bool {
	var pe *fs.PathError
	return errors.As(err, &pe) && strings.Contains(pe.Err.Error(), rootEscapeText)
}

// OpenInRoot opens the file that name designates inside root, for reading, or
// returns an error wrapping ErrOutsideRoot. Use it in place of ResolveInRoot
// followed by os.Open whenever the caller is going to read the file.
//
// The name is checked as ResolveInRoot checks it, and the open itself is
// confined by the operating system through os.Root: a symbolic link, in any
// element of the name, that leads out of the root is refused at the moment of
// the open, not at an earlier check. A link whose target is relative and stays
// inside the root is followed. A link whose target is an absolute path is
// refused even when that path is inside the root, which is stricter than
// ResolveInRoot; os.Root does not follow absolute links. The file cannot be
// redirected between the check and the open,
// because they are one step, and the returned handle stays bound to the file
// that was opened. Check the handle, with Stat on the *os.File, for what you
// need of it: a regular file, a size limit. Reading through the handle keeps
// the size you checked from changing underneath you only if you also bound
// the read, for instance with io.LimitReader.
//
// A relative name is taken from root. An absolute name is accepted only if it
// is spelled inside root, with root written the same way: pass the result of
// RealDir as root, and absolute names that reach the same directory through a
// different link are refused. root must be an existing directory.
//
// The file must exist. An error for a missing file satisfies
// errors.Is(err, os.ErrNotExist) and does not wrap ErrOutsideRoot.
func OpenInRoot(root, name string) (f *os.File, err error) {
	if err := checkName(name); err != nil {
		return nil, err
	}
	rel := filepath.Clean(name)
	if filepath.IsAbs(rel) {
		// A name outside root becomes a ".." path, which os.Root refuses, and a
		// name on another volume has no relative form at all.
		if rel, err = filepath.Rel(root, rel); err != nil {
			return nil, fmt.Errorf("%w: %q", ErrOutsideRoot, name)
		}
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("oscompat/fs: open root %q: %w", root, err)
	}
	defer func() {
		// Closing the root does not close files opened through it.
		if cerr := r.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("oscompat/fs: close root %q: %w", root, errors.Join(cerr, f.Close()))
			f = nil
		}
	}()
	f, err = r.Open(rel)
	if err != nil {
		if isRootEscape(err) {
			return nil, fmt.Errorf("%w: %q", ErrOutsideRoot, name)
		}
		return nil, err
	}
	return f, nil
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
