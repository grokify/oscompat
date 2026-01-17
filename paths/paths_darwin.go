//go:build darwin

package paths

import (
	"os"
	"path/filepath"
)

// UserConfig returns the user-specific configuration directory.
// macOS: ~/Library/Application Support (Apple's recommended location)
// Also respects XDG_CONFIG_HOME for cross-platform tools.
func UserConfig() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support"), nil
}

// UserData returns the user-specific data directory.
// macOS: ~/Library/Application Support (same as config on macOS)
// Also respects XDG_DATA_HOME for cross-platform tools.
func UserData() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support"), nil
}

// UserCache returns the user-specific cache directory.
// macOS: ~/Library/Caches
// Also respects XDG_CACHE_HOME for cross-platform tools.
func UserCache() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Caches"), nil
}

// UserRuntime returns the user-specific runtime directory.
// macOS: ~/Library/Application Support (no separate runtime dir on macOS)
// Respects XDG_RUNTIME_DIR if set.
func UserRuntime() (string, error) {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	// macOS doesn't have a standard runtime directory, use Application Support
	return filepath.Join(home, "Library", "Application Support"), nil
}

// SystemConfig returns the system-wide configuration directory.
// macOS: /etc (same as other Unix systems)
func SystemConfig() (string, error) {
	return "/etc", nil
}
