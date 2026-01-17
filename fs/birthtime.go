package fs

import (
	"os"
	"time"
)

// BirthtimeSupported returns true if the platform supports file birthtime.
// macOS and Windows support birthtime; Linux does not reliably support it.
func BirthtimeSupported() bool {
	return birthtimeSupported
}

// Birthtime returns the file's creation time if available.
// On platforms without birthtime support, it falls back to modification time.
//
// Platform behavior:
//   - macOS: Uses Birthtimespec from syscall.Stat_t
//   - Windows: Uses CreationTime from Win32FileAttributeData
//   - Linux: Falls back to ModTime (birthtime not reliably available)
func Birthtime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return birthtime(path, info)
}

// BirthtimeInfo returns the file's creation time from an existing FileInfo.
// This is more efficient when you already have the FileInfo.
// On platforms without birthtime support, it falls back to modification time.
func BirthtimeInfo(info os.FileInfo) time.Time {
	return birthtimeInfo(info)
}
