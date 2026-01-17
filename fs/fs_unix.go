//go:build !windows

package fs

// isCaseSensitive indicates whether file paths are case-sensitive on this platform.
// Unix and macOS are typically case-sensitive (though macOS HFS+ is case-insensitive
// by default, APFS can be either - we use the stricter case-sensitive assumption).
const isCaseSensitive = true
