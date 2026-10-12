//go:build windows

package fs

// permissionsEnforced: Windows does not enforce POSIX file modes.
const permissionsEnforced = false
