// Package testutil provides testing utilities that work across platforms.
package testutil

import (
	"bytes"
	"os"
)

// CaptureStdout captures stdout during the execution of fn and returns the output.
// This implementation reads from the pipe concurrently to avoid deadlocks on Windows
// where pipe buffers are smaller (typically 4KB-64KB) and can block if not drained.
func CaptureStdout(fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return ""
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
	w.Close()
	os.Stdout = old

	// Wait for reader goroutine to finish
	return <-outputCh
}

// CaptureStderr captures stderr during the execution of fn and returns the output.
// This implementation reads from the pipe concurrently to avoid deadlocks on Windows.
func CaptureStderr(fn func()) string {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return ""
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
	w.Close()
	os.Stderr = old

	// Wait for reader goroutine to finish
	return <-outputCh
}

// CaptureOutput captures both stdout and stderr during the execution of fn.
// Returns stdout and stderr output separately.
func CaptureOutput(fn func()) (stdout, stderr string) {
	oldOut := os.Stdout
	oldErr := os.Stderr

	rOut, wOut, errOut := os.Pipe()
	rErr, wErr, errErr := os.Pipe()
	if errOut != nil || errErr != nil {
		return "", ""
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
	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	// Wait for reader goroutines
	return <-stdoutCh, <-stderrCh
}
