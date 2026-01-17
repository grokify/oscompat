//go:build darwin

package fs

import (
	"os"
	"syscall"
	"time"
)

const birthtimeSupported = true

// birthtime returns the file's creation time on macOS.
func birthtime(_ string, info os.FileInfo) (time.Time, error) {
	return birthtimeInfo(info), nil
}

// birthtimeInfo extracts birthtime from FileInfo on macOS.
func birthtimeInfo(info os.FileInfo) time.Time {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Fallback to ModTime if we can't get the underlying stat
		return info.ModTime()
	}

	// macOS Stat_t has Birthtimespec field
	return time.Unix(stat.Birthtimespec.Sec, stat.Birthtimespec.Nsec)
}
