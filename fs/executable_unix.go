//go:build !windows

package fs

import "os"

// isExecutablePlatform checks if a file is executable on Unix/macOS.
// A file is executable if any of the execute bits (user, group, or other) are set.
func isExecutablePlatform(info os.FileInfo, _ string) bool {
	return info.Mode()&0111 != 0
}
