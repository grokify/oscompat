// Package testutil provides testing utilities that work across platforms.
package testutil

import (
	"bytes"
	"fmt"
	"os"
)

// CaptureStdout captures stdout during the execution of fn and returns the output.
// This implementation reads from the pipe concurrently to avoid deadlocks on Windows
// where pipe buffers are smaller (typically 4KB-64KB) and can block if not drained.
// Returns an error if pipe creation or closing fails.
func CaptureStdout(fn func()) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("create pipe: %w", err)
	}
	os.Stdout = w

	// Read from pipe concurrently to prevent buffer deadlock
	outputCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		outputCh <- buf.String()
	}()

	// Execute the function
	fn()

	// Close writer and restore stdout
	closeErr := w.Close()
	os.Stdout = old

	// Wait for reader goroutine to finish
	output := <-outputCh

	if closeErr != nil {
		return output, fmt.Errorf("close pipe: %w", closeErr)
	}
	return output, nil
}

// CaptureStderr captures stderr during the execution of fn and returns the output.
// This implementation reads from the pipe concurrently to avoid deadlocks on Windows.
// Returns an error if pipe creation or closing fails.
func CaptureStderr(fn func()) (string, error) {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("create pipe: %w", err)
	}
	os.Stderr = w

	// Read from pipe concurrently to prevent buffer deadlock
	outputCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		outputCh <- buf.String()
	}()

	// Execute the function
	fn()

	// Close writer and restore stderr
	closeErr := w.Close()
	os.Stderr = old

	// Wait for reader goroutine to finish
	output := <-outputCh

	if closeErr != nil {
		return output, fmt.Errorf("close pipe: %w", closeErr)
	}
	return output, nil
}

// CaptureOutput captures both stdout and stderr during the execution of fn.
// Returns stdout and stderr output separately.
// Returns an error if pipe creation or closing fails.
func CaptureOutput(fn func()) (stdout, stderr string, err error) {
	oldOut := os.Stdout
	oldErr := os.Stderr

	rOut, wOut, errOut := os.Pipe()
	if errOut != nil {
		return "", "", fmt.Errorf("create stdout pipe: %w", errOut)
	}
	rErr, wErr, errErr := os.Pipe()
	if errErr != nil {
		_ = rOut.Close()
		_ = wOut.Close()
		return "", "", fmt.Errorf("create stderr pipe: %w", errErr)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	// Read from pipes concurrently
	stdoutCh := make(chan string)
	stderrCh := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rOut)
		stdoutCh <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rErr)
		stderrCh <- buf.String()
	}()

	// Execute the function
	fn()

	// Close writers and restore
	closeOutErr := wOut.Close()
	closeErrErr := wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	// Wait for reader goroutines
	stdout = <-stdoutCh
	stderr = <-stderrCh

	if closeOutErr != nil {
		return stdout, stderr, fmt.Errorf("close stdout pipe: %w", closeOutErr)
	}
	if closeErrErr != nil {
		return stdout, stderr, fmt.Errorf("close stderr pipe: %w", closeErrErr)
	}
	return stdout, stderr, nil
}
