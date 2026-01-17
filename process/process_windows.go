//go:build windows

package process

import (
	"os"
	"os/exec"
)

// setSysProcAttr sets Windows-specific process attributes for daemon detachment.
// On Windows, we don't have Setpgid, but the process will still run independently.
func setSysProcAttr(cmd *exec.Cmd) {
	// No special attributes needed on Windows for basic detachment.
	// The process will run independently by default.
}

// signalProcess terminates a process on Windows.
// Windows doesn't have SIGTERM, so we use Process.Kill().
func signalProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// signalProcessHandle terminates an existing process handle on Windows.
func signalProcessHandle(process *os.Process) error {
	return process.Kill()
}
