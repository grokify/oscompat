package tsync

import (
	"syscall"
	"time"
)

// FromTimespec converts a syscall.Timespec to time.Time.
// This handles platform differences in the Timespec field types.
//
// On Windows, this function still accepts Timespec but it's rarely used
// since Windows uses different time structures (FILETIME).
func FromTimespec(ts syscall.Timespec) time.Time {
	//nolint:unconvert // Explicit conversion for cross-platform compatibility
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

// FromTimeval converts a syscall.Timeval to time.Time.
// This handles platform differences in the Timeval field types.
//
// Note: Timeval has microsecond precision (Usec field), so some precision
// is lost compared to nanosecond-precision time.Time.
func FromTimeval(tv syscall.Timeval) time.Time {
	// Convert microseconds to nanoseconds
	//nolint:unconvert // Explicit conversion for cross-platform compatibility
	return time.Unix(int64(tv.Sec), int64(tv.Usec)*1000)
}
