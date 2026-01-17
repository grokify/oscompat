//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris || aix

package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// UserConfig returns the user-specific configuration directory.
// Follows XDG Base Directory Specification: $XDG_CONFIG_HOME or ~/.config
func UserConfig() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

// UserData returns the user-specific data directory.
// Follows XDG Base Directory Specification: $XDG_DATA_HOME or ~/.local/share
func UserData() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// UserCache returns the user-specific cache directory.
// Follows XDG Base Directory Specification: $XDG_CACHE_HOME or ~/.cache
func UserCache() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache"), nil
}

// UserRuntime returns the user-specific runtime directory.
// Follows XDG Base Directory Specification: $XDG_RUNTIME_DIR or /tmp/<user>-runtime
func UserRuntime() (string, error) {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir, nil
	}
	// Fallback to /tmp with user ID for uniqueness
	return fmt.Sprintf("/tmp/runtime-%d", os.Getuid()), nil
}

// SystemConfig returns the system-wide configuration directory.
// Returns /etc on Unix systems.
func SystemConfig() (string, error) {
	return "/etc", nil
}
