package process

import (
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when the process does not exist.
var ErrNotFound = errors.New("process not found")

// Exists reports whether a process with the given PID is running. A process
// owned by another user counts as running. It returns an error only when the
// answer cannot be determined, such as an invalid PID.
func Exists(pid int) (bool, error) {
	if err := checkPID(pid); err != nil {
		return false, err
	}
	return exists(pid)
}

// StartTime returns when the process with the given PID started. It returns
// ErrNotFound if the process does not exist.
//
// Comparing a start time against a recorded one detects a reused PID: a
// stale record that names a PID the system has since handed to a different
// process will not match.
//
// On Unix the time comes from ps and has one-second resolution. On Windows it
// is the process creation time.
func StartTime(pid int) (time.Time, error) {
	if err := checkPID(pid); err != nil {
		return time.Time{}, err
	}
	return startTime(pid)
}

// checkPID rejects values that would address something other than one
// process. On Unix, PID 0 and negative values address process groups.
func checkPID(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("process: invalid pid %d", pid)
	}
	return nil
}
