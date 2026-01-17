//go:build linux

package fs

import (
	"os"
	"time"
)

const birthtimeSupported = false

// birthtime returns the file's modification time on Linux.
// Linux does not reliably support birthtime across filesystems.
func birthtime(path string, info os.FileInfo) (time.Time, error) {
	return info.ModTime(), nil
}

// birthtimeInfo returns ModTime on Linux as birthtime is not reliably available.
func birthtimeInfo(info os.FileInfo) time.Time {
	return info.ModTime()
}
