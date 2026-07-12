package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetTempHomeDir(t *testing.T) {
	tmpHome := SetTempHomeDir(t)

	// Verify HOME is set (Unix/macOS)
	if got := os.Getenv("HOME"); got != tmpHome {
		t.Errorf("HOME = %q, want %q", got, tmpHome)
	}

	// Verify USERPROFILE is set (Windows)
	if got := os.Getenv("USERPROFILE"); got != tmpHome {
		t.Errorf("USERPROFILE = %q, want %q", got, tmpHome)
	}

	// Verify os.UserHomeDir() returns the temp directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error: %v", err)
	}
	if home != tmpHome {
		t.Errorf("os.UserHomeDir() = %q, want %q", home, tmpHome)
	}

	// Verify we can create files in the temp home
	testFile := filepath.Join(tmpHome, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Errorf("failed to write file in temp home: %v", err)
	}
}

func TestSetTempHomeDir_Isolation(t *testing.T) {
	// Get original home before test
	originalHome, _ := os.UserHomeDir()

	t.Run("subtest", func(t *testing.T) {
		tmpHome := SetTempHomeDir(t)
		home, _ := os.UserHomeDir()
		if home != tmpHome {
			t.Errorf("os.UserHomeDir() = %q, want %q", home, tmpHome)
		}
	})

	// After subtest, home should be restored (t.Setenv cleanup)
	restoredHome, _ := os.UserHomeDir()
	if restoredHome != originalHome {
		t.Errorf("home not restored: got %q, want %q", restoredHome, originalHome)
	}
}
