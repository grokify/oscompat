// Package fs provides cross-platform filesystem utilities.
//
// This package addresses platform differences in:
//   - File permissions (Unix uses mode bits, Windows largely ignores them)
//   - Path separators (Unix uses /, Windows uses \)
//   - Path validation for security (preventing directory traversal)
package fs

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Common errors.
var (
	// ErrPathTraversal is returned when a path attempts directory traversal.
	ErrPathTraversal = errors.New("oscompat/fs: path traversal detected")

	// ErrEmptyPath is returned when an empty path is provided.
	ErrEmptyPath = errors.New("oscompat/fs: empty path")

	// ErrAbsolutePath is returned when an absolute path is provided where relative is expected.
	ErrAbsolutePath = errors.New("oscompat/fs: absolute path not allowed")
)

// Default permission constants.
// These follow Unix conventions; Windows ignores permission bits but the values
// are still useful for cross-platform code consistency.
const (
	// DefaultDirPerm is the default permission for directories (rwxr-xr-x).
	DefaultDirPerm os.FileMode = 0755

	// DefaultFilePerm is the default permission for files (rw-r--r--).
	DefaultFilePerm os.FileMode = 0644

	// PrivateDirPerm is for directories that should only be accessible by owner (rwx------).
	PrivateDirPerm os.FileMode = 0700

	// PrivateFilePerm is for files that should only be accessible by owner (rw-------).
	PrivateFilePerm os.FileMode = 0600

	// ExecutablePerm is for executable files (rwxr-xr-x).
	ExecutablePerm os.FileMode = 0755
)

// ValidatePath checks a path for directory traversal attacks.
// It returns an error if the path:
//   - Is empty
//   - Contains ".." that would escape the root
//   - Is absolute (starts with / or drive letter on Windows)
//
// This function handles both forward slashes and backslashes for Windows compatibility.
func ValidatePath(p string) error {
	if p == "" {
		return ErrEmptyPath
	}

	// Check for absolute paths (both Unix and Windows style)
	if filepath.IsAbs(p) || (len(p) >= 2 && p[1] == ':') {
		return ErrAbsolutePath
	}

	// Normalize to forward slashes for consistent checking
	normalized := strings.ReplaceAll(p, "\\", "/")

	// Clean the path
	cleaned := path.Clean(normalized)

	// Check if cleaned path escapes root
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ErrPathTraversal
	}

	// Also check for embedded traversal (e.g., "foo/../../../bar")
	if strings.Contains(cleaned, "/../") {
		return ErrPathTraversal
	}

	return nil
}

// ValidatePathStrict is like ValidatePath but also rejects paths starting with ".".
// This is useful when hidden files/directories should not be allowed.
func ValidatePathStrict(p string) error {
	if err := ValidatePath(p); err != nil {
		return err
	}

	normalized := strings.ReplaceAll(p, "\\", "/")
	cleaned := path.Clean(normalized)

	// Reject hidden files/directories
	if strings.HasPrefix(cleaned, ".") {
		return ErrPathTraversal
	}

	return nil
}

// NormalizePath converts OS-specific path separators to forward slashes.
// This is useful for storing paths in a platform-independent format.
// The path is also cleaned (redundant separators removed, . and .. resolved).
// Backslashes are converted to forward slashes on all platforms.
func NormalizePath(p string) string {
	if p == "" {
		return ""
	}
	// Convert backslashes to forward slashes first (for cross-platform input)
	p = strings.ReplaceAll(p, "\\", "/")
	// Clean using path (forward-slash aware) package
	return path.Clean(p)
}

// OSPath converts forward slashes to OS-specific path separators.
// This is the inverse of NormalizePath.
func OSPath(p string) string {
	if p == "" {
		return ""
	}
	return filepath.FromSlash(p)
}

// JoinNormalized joins path elements and returns a normalized (forward-slash) path.
// This is useful for building storage keys or URIs.
func JoinNormalized(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}
	// Use path.Join which always uses forward slashes
	return path.Join(elem...)
}

// JoinOS joins path elements using OS-specific separators.
// This is a convenience wrapper around filepath.Join.
func JoinOS(elem ...string) string {
	return filepath.Join(elem...)
}

// SafeJoin safely joins a base path with a relative path, preventing traversal.
// Returns an error if the result would escape the base directory.
func SafeJoin(base, rel string) (string, error) {
	if err := ValidatePath(rel); err != nil {
		return "", err
	}

	// Convert relative path to OS format
	osRel := OSPath(rel)

	// Join paths
	joined := filepath.Join(base, osRel)

	// Verify the result is still under base
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}

	// Ensure joined path is under base
	if !strings.HasPrefix(absJoined, absBase) {
		return "", ErrPathTraversal
	}

	return joined, nil
}

// MkdirAll creates a directory and all parent directories with the specified permissions.
// This is a convenience wrapper that uses DefaultDirPerm if perm is 0.
func MkdirAll(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultDirPerm
	}
	return os.MkdirAll(path, perm)
}

// MkdirAllPrivate creates a private directory (owner-only access).
func MkdirAllPrivate(path string) error {
	return os.MkdirAll(path, PrivateDirPerm)
}

// WriteFile writes data to a file with the specified permissions.
// This is a convenience wrapper that uses DefaultFilePerm if perm is 0.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultFilePerm
	}
	return os.WriteFile(filename, data, perm)
}

// WriteFilePrivate writes data to a private file (owner-only access).
func WriteFilePrivate(filename string, data []byte) error {
	return os.WriteFile(filename, data, PrivateFilePerm)
}

// IsCaseSensitive returns whether the current OS has case-sensitive file paths.
// Returns false on Windows (case-insensitive), true on Unix/macOS (case-sensitive).
// Note: macOS HFS+ is case-insensitive by default, but APFS can be either.
// This function returns the typical/default behavior for each platform.
func IsCaseSensitive() bool {
	return isCaseSensitive
}

// PathEqual compares two paths for equality, using case-insensitive comparison
// on Windows and case-sensitive comparison on Unix.
// Both paths are normalized before comparison.
func PathEqual(path1, path2 string) bool {
	// Normalize both paths
	p1 := NormalizePath(path1)
	p2 := NormalizePath(path2)

	if isCaseSensitive {
		return p1 == p2
	}
	return strings.EqualFold(p1, p2)
}

// PathHasPrefix checks if path has the given prefix, using case-insensitive
// comparison on Windows and case-sensitive comparison on Unix.
// Both paths are normalized before comparison.
func PathHasPrefix(path, prefix string) bool {
	// Normalize both paths
	p := NormalizePath(path)
	pfx := NormalizePath(prefix)

	// Ensure prefix ends with separator for proper directory matching
	if pfx != "" && !strings.HasSuffix(pfx, "/") {
		pfx += "/"
	}

	if isCaseSensitive {
		return strings.HasPrefix(p+"/", pfx)
	}
	return strings.HasPrefix(strings.ToLower(p)+"/", strings.ToLower(pfx))
}
