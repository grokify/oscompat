package process_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/grokify/oscompat/process"
)

func TestSetDetached(t *testing.T) {
	// Verify SetDetached doesn't panic.
	cmd := exec.Command("echo", "test")
	process.SetDetached(cmd)

	// On Unix, SysProcAttr should be set; on Windows, it may be nil.
	if runtime.GOOS != "windows" {
		if cmd.SysProcAttr == nil {
			t.Error("SetDetached did not set SysProcAttr on Unix")
		}
	}
}

func TestSignalNonExistentProcess(t *testing.T) {
	// Signaling a non-existent PID should return an error.
	// Use a very high PID that's unlikely to exist.
	err := process.Signal(999999999)
	if err == nil {
		t.Error("Signal on non-existent PID should return error")
	}
}

func TestFindAndSignalNonExistentProcess(t *testing.T) {
	// FindAndSignal on a non-existent PID should return an error.
	err := process.FindAndSignal(999999999)
	if err == nil {
		t.Error("FindAndSignal on non-existent PID should return error")
	}
}
