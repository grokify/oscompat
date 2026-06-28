package testutil

import (
	"testing"
	"time"
)

// WindowsTimerResolution is the typical timer resolution on Windows (~15.6ms).
// Operations faster than this may report a duration of 0.
const WindowsTimerResolution = 16 * time.Millisecond

// DurationNonNegative returns true if the duration is non-negative (>= 0).
// This is the appropriate check for operation durations on Windows, where
// fast operations may have a duration of exactly 0 due to the system clock's
// coarse resolution (~15.6ms vs nanoseconds on Unix).
//
// Use this instead of d > 0 when testing that a duration was measured.
func DurationNonNegative(d time.Duration) bool {
	return d >= 0
}

// DurationPositiveOrZero is an alias for DurationNonNegative.
// It makes test assertions clearer when the intent is "some time passed,
// or the operation was too fast to measure on Windows".
func DurationPositiveOrZero(d time.Duration) bool {
	return DurationNonNegative(d)
}

// AssertDurationNonNegative fails the test if the duration is negative.
// This is the cross-platform safe assertion for operation durations.
//
// On Windows, the system clock resolution is ~15.6ms, so operations
// that complete faster than this will have a duration of 0. This is
// expected behavior, not an error.
func AssertDurationNonNegative(t testing.TB, d time.Duration) {
	t.Helper()
	if d < 0 {
		t.Errorf("duration should be non-negative, got %v", d)
	}
}

// AssertDurationNonNegativeMsg fails the test with a custom message if the duration is negative.
func AssertDurationNonNegativeMsg(t testing.TB, d time.Duration, msg string) {
	t.Helper()
	if d < 0 {
		t.Errorf("%s: duration should be non-negative, got %v", msg, d)
	}
}
