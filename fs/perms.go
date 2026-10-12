package fs

import (
	"errors"
	"fmt"
	"os"
)

// ErrNotPrivate is returned by [CheckPrivate] for a file that group or others
// can access.
var ErrNotPrivate = errors.New("oscompat/fs: file is accessible by group or others")

// PermissionsEnforced reports whether the platform enforces file mode bits.
// On Windows it is false: modes are not meaningful there, so code that checks
// them, and tests that assert them, must not.
func PermissionsEnforced() bool { return permissionsEnforced }

// IsPrivate reports whether info describes a file that only its owner can
// access. It is always true where [PermissionsEnforced] is false.
func IsPrivate(info os.FileInfo) bool {
	return !permissionsEnforced || info.Mode().Perm()&0o077 == 0
}

// CheckPrivate returns an error matching [ErrNotPrivate], naming the mode, if
// info describes a file that group or others can access. It returns nil where
// [PermissionsEnforced] is false.
func CheckPrivate(info os.FileInfo) error {
	if IsPrivate(info) {
		return nil
	}
	return fmt.Errorf("%w: mode %04o", ErrNotPrivate, info.Mode().Perm())
}

// ChmodPrivate restricts an open file to its owner (0600). Where
// [PermissionsEnforced] is false a failure is ignored, because the mode could
// not have been relied on anyway.
func ChmodPrivate(f *os.File) error {
	if err := f.Chmod(PrivateFilePerm); err != nil && permissionsEnforced {
		return err
	}
	return nil
}
