//go:build !windows

package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func exists(pid int) (bool, error) {
	err := syscall.Kill(pid, 0)
	switch {
	case err == nil, errors.Is(err, syscall.EPERM):
		return true, nil
	case errors.Is(err, syscall.ESRCH):
		return false, nil
	default:
		return false, fmt.Errorf("process: check pid %d: %w", pid, err)
	}
}

func startTime(pid int) (time.Time, error) {
	ok, err := exists(pid)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return time.Time{}, ErrNotFound
	}
	cmd := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)) //nolint:gosec // pid is an int
	// A fixed locale keeps the month and day names parseable.
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	out, err := cmd.Output()
	if err != nil {
		// The process can exit between the check above and ps running.
		if ok, existsErr := exists(pid); existsErr == nil && !ok {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("process: ps for pid %d: %w", pid, err)
	}
	// lstart is local time, e.g. "Wed Oct  7 10:04:54 2026".
	t, err := time.ParseInLocation("Mon Jan _2 15:04:05 2006", strings.TrimSpace(string(out)), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("process: parse start time for pid %d: %w", pid, err)
	}
	return t, nil
}
