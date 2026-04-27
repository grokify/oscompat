package testutil

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCaptureStdout(t *testing.T) {
	output, err := CaptureStdout(func() {
		fmt.Println("hello world")
	})
	if err != nil {
		t.Fatalf("CaptureStdout failed: %v", err)
	}

	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected 'hello world' in output, got: %q", output)
	}
}

func TestCaptureStdoutLargeOutput(t *testing.T) {
	// Test with output larger than typical Windows pipe buffer (64KB)
	largeString := strings.Repeat("x", 100000) // 100KB

	output, err := CaptureStdout(func() {
		fmt.Print(largeString)
	})
	if err != nil {
		t.Fatalf("CaptureStdout failed: %v", err)
	}

	if len(output) != len(largeString) {
		t.Errorf("Expected output length %d, got %d", len(largeString), len(output))
	}
}

func TestCaptureStderr(t *testing.T) {
	output, err := CaptureStderr(func() {
		fmt.Fprintln(os.Stderr, "error message")
	})
	if err != nil {
		t.Fatalf("CaptureStderr failed: %v", err)
	}

	if !strings.Contains(output, "error message") {
		t.Errorf("Expected 'error message' in output, got: %q", output)
	}
}

func TestCaptureOutput(t *testing.T) {
	stdout, stderr, err := CaptureOutput(func() {
		fmt.Println("stdout message")
		fmt.Fprintln(os.Stderr, "stderr message")
	})
	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	if !strings.Contains(stdout, "stdout message") {
		t.Errorf("Expected 'stdout message' in stdout, got: %q", stdout)
	}

	if !strings.Contains(stderr, "stderr message") {
		t.Errorf("Expected 'stderr message' in stderr, got: %q", stderr)
	}
}

func TestCaptureStdoutEmpty(t *testing.T) {
	output, err := CaptureStdout(func() {
		// Do nothing
	})
	if err != nil {
		t.Fatalf("CaptureStdout failed: %v", err)
	}

	if output != "" {
		t.Errorf("Expected empty output, got: %q", output)
	}
}
