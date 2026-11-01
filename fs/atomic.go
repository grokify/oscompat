package fs

import (
	"errors"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to name so that a reader sees either the
// whole old file or the whole new one, never a partial write. It writes a
// temporary file in the same directory, flushes it, sets perm (DefaultFilePerm
// when perm is 0), renames it over name and flushes the directory with
// [SyncDir].
//
// The mode is set only where [PermissionsEnforced] is true; elsewhere a
// failure to set it is ignored, as it could not have been relied on. On
// Windows the rename fails if another process holds name open. On error the
// temporary file is removed and name is unchanged.
func WriteFileAtomic(name string, data []byte, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultFilePerm
	}
	dir := filepath.Dir(name)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(name)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	fail := func(err error) error {
		return errors.Join(err, tmp.Close(), os.Remove(tmpName))
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(err)
	}
	if err := tmp.Sync(); err != nil {
		return fail(err)
	}
	if err := tmp.Chmod(perm); err != nil && permissionsEnforced {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		return errors.Join(err, os.Remove(tmpName))
	}
	if err := os.Rename(tmpName, name); err != nil {
		return errors.Join(err, os.Remove(tmpName))
	}
	return SyncDir(dir)
}

// WriteFileAtomicPrivate is [WriteFileAtomic] with PrivateFilePerm (0600).
func WriteFileAtomicPrivate(name string, data []byte) error {
	return WriteFileAtomic(name, data, PrivateFilePerm)
}
