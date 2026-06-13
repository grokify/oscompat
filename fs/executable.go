package fs

import "os"

// IsExecutable checks if a file at the given path is executable.
//
// On Unix/macOS, this checks if any of the execute permission bits are set
// (user, group, or other).
//
// On Windows, executability is determined differently - Windows doesn't use
// permission bits. For files intended to be run by shells (like git hooks),
// Windows considers them executable if they exist and are regular files.
// For native Windows executables, the function checks for common executable
// extensions (.exe, .cmd, .bat, .com, .ps1).
//
// Returns an error if the file cannot be stat'd.
func IsExecutable(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return IsExecutableInfo(info, path), nil
}

// IsExecutableInfo checks if a file is executable given its FileInfo.
// The path parameter is only used on Windows to check file extensions;
// on Unix it is ignored.
//
// This variant is useful when you already have the FileInfo from a previous
// os.Stat or directory walk.
func IsExecutableInfo(info os.FileInfo, path string) bool {
	if info.IsDir() {
		return false
	}
	return isExecutablePlatform(info, path)
}
