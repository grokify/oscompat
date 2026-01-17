//go:build windows

package paths

import (
	"os"
	"path/filepath"
)

// UserConfig returns the user-specific configuration directory.
// Windows: %APPDATA% (typically C:\Users\<user>\AppData\Roaming)
func UserConfig() (string, error) {
	if dir := os.Getenv("APPDATA"); dir != "" {
		return dir, nil
	}
	// Fallback to home directory
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "AppData", "Roaming"), nil
}

// UserData returns the user-specific data directory.
// Windows: %LOCALAPPDATA% (typically C:\Users\<user>\AppData\Local)
func UserData() (string, error) {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return dir, nil
	}
	// Fallback to home directory
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "AppData", "Local"), nil
}

// UserCache returns the user-specific cache directory.
// Windows: %LOCALAPPDATA%\cache
func UserCache() (string, error) {
	base, err := UserData()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "cache"), nil
}

// UserRuntime returns the user-specific runtime directory.
// Windows: %LOCALAPPDATA%\run (Windows doesn't have a standard runtime dir)
func UserRuntime() (string, error) {
	base, err := UserData()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "run"), nil
}

// SystemConfig returns the system-wide configuration directory.
// Windows: %ProgramData% (typically C:\ProgramData)
func SystemConfig() (string, error) {
	if dir := os.Getenv("ProgramData"); dir != "" {
		return dir, nil
	}
	// Fallback
	return `C:\ProgramData`, nil
}
