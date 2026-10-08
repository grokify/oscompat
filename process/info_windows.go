//go:build windows

package process

import (
	"errors"
	"fmt"
	"syscall"
	"time"
)

const (
	// processQueryLimitedInformation is enough to read exit status and
	// times, and is granted for processes that deny full query access.
	processQueryLimitedInformation = 0x1000
	// stillActive is the exit code GetExitCodeProcess reports for a process
	// that has not exited.
	stillActive = 259

	errorAccessDenied     = syscall.Errno(5)
	errorInvalidParameter = syscall.Errno(87)
)

// open returns a query handle for the process, or ErrNotFound if there is no
// such process. The caller closes the handle.
func open(pid int) (syscall.Handle, error) {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid)) //nolint:gosec // pid is validated positive
	if err != nil {
		if errors.Is(err, errorInvalidParameter) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return h, nil
}

func exists(pid int) (bool, error) {
	h, err := open(pid)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if errors.Is(err, errorAccessDenied) {
		return true, nil // the process exists but belongs to someone else
	}
	if err != nil {
		return false, fmt.Errorf("process: open pid %d: %w", pid, err)
	}
	defer syscall.CloseHandle(h) //nolint:errcheck // a failed close of a read-only handle is not actionable

	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false, fmt.Errorf("process: exit code for pid %d: %w", pid, err)
	}
	// A handle can outlive its process, so a live handle is not enough.
	return code == stillActive, nil
}

func startTime(pid int) (time.Time, error) {
	h, err := open(pid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("process: open pid %d: %w", pid, err)
	}
	defer syscall.CloseHandle(h) //nolint:errcheck // a failed close of a read-only handle is not actionable

	var creation, exit, kernel, user syscall.Filetime
	if err := syscall.GetProcessTimes(h, &creation, &exit, &kernel, &user); err != nil {
		return time.Time{}, fmt.Errorf("process: times for pid %d: %w", pid, err)
	}
	return time.Unix(0, creation.Nanoseconds()), nil
}
