// Package paths provides cross-platform directory resolution for application
// configuration, data, and cache storage.
//
// This package abstracts platform differences in standard directory locations:
//   - Unix/Linux: XDG Base Directory Specification (~/.config, ~/.local/share, etc.)
//   - macOS: ~/Library/Application Support, ~/Library/Caches
//   - Windows: %APPDATA%, %LOCALAPPDATA%, %ProgramData%
//
// All functions return absolute paths. App-specific functions (AppConfig, AppData, etc.)
// will create the directory if it doesn't exist.
package paths

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrHomeNotFound is returned when the user's home directory cannot be determined.
var ErrHomeNotFound = errors.New("oscompat/paths: home directory not found")

// ErrInvalidAppName is returned when an empty app name is provided.
var ErrInvalidAppName = errors.New("oscompat/paths: app name cannot be empty")

// Home returns the current user's home directory.
func Home() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}
	return home, nil
}

// AppConfig returns the app-specific configuration directory, creating it if needed.
// The directory follows platform conventions:
//   - Unix/Linux: ~/.config/<appName>
//   - macOS: ~/Library/Application Support/<appName>
//   - Windows: %APPDATA%\<appName>
func AppConfig(appName string) (string, error) {
	if appName == "" {
		return "", ErrInvalidAppName
	}
	base, err := UserConfig()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// AppData returns the app-specific data directory, creating it if needed.
// The directory follows platform conventions:
//   - Unix/Linux: ~/.local/share/<appName>
//   - macOS: ~/Library/Application Support/<appName>
//   - Windows: %LOCALAPPDATA%\<appName>
func AppData(appName string) (string, error) {
	if appName == "" {
		return "", ErrInvalidAppName
	}
	base, err := UserData()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// AppCache returns the app-specific cache directory, creating it if needed.
// The directory follows platform conventions:
//   - Unix/Linux: ~/.cache/<appName>
//   - macOS: ~/Library/Caches/<appName>
//   - Windows: %LOCALAPPDATA%\<appName>\cache
func AppCache(appName string) (string, error) {
	if appName == "" {
		return "", ErrInvalidAppName
	}
	base, err := UserCache()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// AppRuntime returns the app-specific runtime directory, creating it if needed.
// This is for runtime files like sockets and PIDs that should not persist across reboots.
//   - Unix/Linux: $XDG_RUNTIME_DIR/<appName> or /tmp/<appName>-<uid>
//   - macOS: ~/Library/Application Support/<appName>/run
//   - Windows: %LOCALAPPDATA%\<appName>\run
func AppRuntime(appName string) (string, error) {
	if appName == "" {
		return "", ErrInvalidAppName
	}
	base, err := UserRuntime()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0700); err != nil { // More restrictive for runtime
		return "", err
	}
	return dir, nil
}

// SystemAppConfig returns the system-wide app configuration directory.
// This requires elevated privileges to write to.
//   - Unix/Linux: /etc/<appName>
//   - Windows: %ProgramData%\<appName>
func SystemAppConfig(appName string) (string, error) {
	if appName == "" {
		return "", ErrInvalidAppName
	}
	base, err := SystemConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}
