// Package process provides cross-platform process management utilities.
//
// This package abstracts platform differences in process handling:
//   - Signal handling (Windows lacks SIGTERM, uses Kill instead)
//   - Process group management (Unix has Setpgid, Windows does not)
//   - Daemon/service detachment patterns
//
// Platform-specific implementations are in process_unix.go and process_windows.go.
package process

import (
	"os"
	"os/exec"
)

// SetDetached configures a command to run detached from the parent process.
// On Unix, this sets up a new process group. On Windows, this is a no-op
// as basic detachment works differently.
func SetDetached(cmd *exec.Cmd) {
	setSysProcAttr(cmd)
}

// Signal sends a termination signal to the process with the given PID.
// On Unix, this sends SIGTERM. On Windows, this calls Process.Kill()
// since Windows doesn't support SIGTERM.
func Signal(pid int) error {
	return signalProcess(pid)
}

// FindAndSignal finds a process by PID and sends a termination signal.
// Returns an error if the process cannot be found or signaled.
func FindAndSignal(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return signalProcessHandle(process)
}
