//go:build windows

package fs

// isCaseSensitive indicates whether file paths are case-sensitive on this platform.
// Windows NTFS is case-insensitive by default.
const isCaseSensitive = false
