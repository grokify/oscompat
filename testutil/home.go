package testutil

import "testing"

// SetTempHomeDir creates a temporary directory and sets it as the home directory
// for the duration of the test. This handles cross-platform differences:
//   - Unix/macOS: os.UserHomeDir() reads $HOME
//   - Windows: os.UserHomeDir() reads %USERPROFILE%
//
// The temporary directory is automatically cleaned up when the test completes.
// Returns the path to the temporary home directory.
//
// Example:
//
//	func TestConfig(t *testing.T) {
//	    tmpHome := testutil.SetTempHomeDir(t)
//	    // os.UserHomeDir() now returns tmpHome on all platforms
//	    config, err := LoadConfig()
//	    // ...
//	}
func SetTempHomeDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)        // Unix, macOS
	t.Setenv("USERPROFILE", tmpDir) // Windows
	return tmpDir
}
