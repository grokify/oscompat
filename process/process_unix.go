//go:build !windows

package process

import (
	"os"
	"os/exec"
	"syscall"
)

// setSysProcAttr sets Unix-specific process attributes for daemon detachment.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// signalProcess sends SIGTERM to a process by PID.
func signalProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}

// signalProcessHandle sends SIGTERM to an existing process handle.
func signalProcessHandle(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
