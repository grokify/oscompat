//go:build !darwin && !windows && !linux

package fs

import (
	"os"
	"time"
)

const birthtimeSupported = false

// birthtime returns the file's modification time on unsupported platforms.
func birthtime(path string, info os.FileInfo) (time.Time, error) {
	return info.ModTime(), nil
}

// birthtimeInfo returns ModTime on unsupported platforms.
func birthtimeInfo(info os.FileInfo) time.Time {
	return info.ModTime()
}
