package fs

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-executable file
	nonExecPath := filepath.Join(tmpDir, "non-executable.txt")
	if err := os.WriteFile(nonExecPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create non-executable file: %v", err)
	}

	// Create an executable file
	execPath := filepath.Join(tmpDir, "executable")
	if err := os.WriteFile(execPath, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatalf("Failed to create executable file: %v", err)
	}

	// Create an executable with .exe extension (for Windows)
	exePath := filepath.Join(tmpDir, "program.exe")
	if err := os.WriteFile(exePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create .exe file: %v", err)
	}

	t.Run("non-executable file", func(t *testing.T) {
		isExec, err := IsExecutable(nonExecPath)
		if err != nil {
			t.Fatalf("IsExecutable error: %v", err)
		}

		if runtime.GOOS == "windows" {
			// On Windows, .txt files are not executable
			if isExec {
				t.Error("Expected .txt file to not be executable on Windows")
			}
		} else {
			// On Unix, 0644 mode is not executable
			if isExec {
				t.Error("Expected 0644 file to not be executable on Unix")
			}
		}
	})

	t.Run("executable file without extension", func(t *testing.T) {
		isExec, err := IsExecutable(execPath)
		if err != nil {
			t.Fatalf("IsExecutable error: %v", err)
		}

		// Should be executable on both platforms
		// Unix: has 0755 mode
		// Windows: no extension (shell script pattern)
		if !isExec {
			t.Error("Expected file without extension to be executable")
		}
	})

	t.Run(".exe file", func(t *testing.T) {
		isExec, err := IsExecutable(exePath)
		if err != nil {
			t.Fatalf("IsExecutable error: %v", err)
		}

		if runtime.GOOS == "windows" {
			// On Windows, .exe files are executable
			if !isExec {
				t.Error("Expected .exe file to be executable on Windows")
			}
		} else {
			// On Unix, the file has 0644 mode, so not executable
			if isExec {
				t.Error("Expected .exe file without exec bit to not be executable on Unix")
			}
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := IsExecutable(filepath.Join(tmpDir, "does-not-exist"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("directory", func(t *testing.T) {
		isExec, err := IsExecutable(tmpDir)
		if err != nil {
			t.Fatalf("IsExecutable error: %v", err)
		}
		if isExec {
			t.Error("Expected directory to not be reported as executable")
		}
	})
}

func TestIsExecutableInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	filePath := filepath.Join(tmpDir, "testfile")
	if err := os.WriteFile(filePath, []byte("content"), 0755); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	isExec := IsExecutableInfo(info, filePath)
	if !isExec {
		t.Error("Expected file to be executable")
	}
}
