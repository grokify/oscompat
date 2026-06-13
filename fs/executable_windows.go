//go:build windows

package fs

import (
	"os"
	"path/filepath"
	"strings"
)

// executableExtensions are file extensions that Windows considers executable.
var executableExtensions = map[string]bool{
	".exe": true,
	".cmd": true,
	".bat": true,
	".com": true,
	".ps1": true,
}

// isExecutablePlatform checks if a file is executable on Windows.
//
// Windows doesn't use permission bits for executability. Instead:
//   - Native executables have specific extensions (.exe, .cmd, .bat, .com, .ps1)
//   - Shell scripts (like git hooks) without extensions are considered executable
//     because they're run by shells like git-bash/MSYS2 that don't check Windows
//     permissions
//
// This function returns true if the file:
//   - Has a recognized executable extension, OR
//   - Has no extension (common for shell scripts, git hooks)
func isExecutablePlatform(info os.FileInfo, path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Check for recognized executable extensions
	if executableExtensions[ext] {
		return true
	}

	// Files without extension are considered executable (shell scripts, git hooks)
	// These are typically run by git-bash, MSYS2, or WSL
	if ext == "" {
		return true
	}

	return false
}
