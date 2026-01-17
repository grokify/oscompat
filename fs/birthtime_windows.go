//go:build windows

package fs

import (
	"os"
	"syscall"
	"time"
)

const birthtimeSupported = true

// birthtime returns the file's creation time on Windows.
func birthtime(path string, info os.FileInfo) (time.Time, error) {
	return birthtimeInfo(info), nil
}

// birthtimeInfo extracts birthtime from FileInfo on Windows.
func birthtimeInfo(info os.FileInfo) time.Time {
	// On Windows, FileInfo.Sys() returns *syscall.Win32FileAttributeData
	// which has CreationTime as a syscall.Filetime
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		// Fallback to ModTime if we can't get the underlying data
		return info.ModTime()
	}

	// Convert Windows Filetime to Unix time
	return time.Unix(0, data.CreationTime.Nanoseconds())
}
