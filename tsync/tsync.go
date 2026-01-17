// Package tsync provides cross-platform timestamp utilities for file synchronization.
//
// This package addresses platform differences in filesystem timestamp precision:
//   - Windows NTFS: ~100 nanosecond precision, but often rounded to 2 seconds for some operations
//   - Linux ext4/XFS: nanosecond precision
//   - macOS APFS: nanosecond precision
//   - FAT32: 2-second precision
//   - Network drives: varies widely
//
// When synchronizing files across platforms or filesystems, timestamps should be
// compared with appropriate tolerance to avoid false positives.
package tsync

import (
	"time"
)

// DefaultTolerance is the recommended tolerance for cross-platform file synchronization.
// 1 second handles:
//   - FAT32 filesystems (2-second precision)
//   - Network drives with reduced precision
//   - Cross-platform timestamp differences
//   - Clock skew between systems
const DefaultTolerance = time.Second

// FAT32Tolerance is the tolerance for FAT32 filesystems (2-second precision).
const FAT32Tolerance = 2 * time.Second

// HighPrecisionTolerance is for comparing timestamps on modern filesystems
// with high precision (NTFS, ext4, APFS).
const HighPrecisionTolerance = 100 * time.Millisecond

// Tolerance returns the recommended tolerance for comparing file modification times.
// Use this when synchronizing files across different platforms or filesystems.
func Tolerance() time.Duration {
	return DefaultTolerance
}

// Equal compares two timestamps with the default tolerance.
// Returns true if the timestamps are within DefaultTolerance of each other.
func Equal(t1, t2 time.Time) bool {
	return EqualWithTolerance(t1, t2, DefaultTolerance)
}

// EqualWithTolerance compares two timestamps with a custom tolerance.
// Returns true if the absolute difference is less than or equal to the tolerance.
func EqualWithTolerance(t1, t2 time.Time, tolerance time.Duration) bool {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// Before returns true if t1 is before t2, accounting for the default tolerance.
// This is more reliable than t1.Before(t2) when comparing across filesystems.
func Before(t1, t2 time.Time) bool {
	return BeforeWithTolerance(t1, t2, DefaultTolerance)
}

// BeforeWithTolerance returns true if t1 is definitively before t2,
// accounting for the specified tolerance.
// Returns false if the times are within tolerance of each other.
func BeforeWithTolerance(t1, t2 time.Time, tolerance time.Duration) bool {
	diff := t2.Sub(t1)
	return diff > tolerance
}

// After returns true if t1 is after t2, accounting for the default tolerance.
// This is more reliable than t1.After(t2) when comparing across filesystems.
func After(t1, t2 time.Time) bool {
	return AfterWithTolerance(t1, t2, DefaultTolerance)
}

// AfterWithTolerance returns true if t1 is definitively after t2,
// accounting for the specified tolerance.
// Returns false if the times are within tolerance of each other.
func AfterWithTolerance(t1, t2 time.Time, tolerance time.Duration) bool {
	diff := t1.Sub(t2)
	return diff > tolerance
}

// Compare compares two timestamps with the default tolerance.
// Returns:
//
//	-1 if t1 is before t2 (beyond tolerance)
//	 0 if t1 and t2 are equal (within tolerance)
//	+1 if t1 is after t2 (beyond tolerance)
func Compare(t1, t2 time.Time) int {
	return CompareWithTolerance(t1, t2, DefaultTolerance)
}

// CompareWithTolerance compares two timestamps with a custom tolerance.
// Returns:
//
//	-1 if t1 is before t2 (beyond tolerance)
//	 0 if t1 and t2 are equal (within tolerance)
//	+1 if t1 is after t2 (beyond tolerance)
func CompareWithTolerance(t1, t2 time.Time, tolerance time.Duration) int {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	if diff <= tolerance {
		return 0
	}
	if t1.Before(t2) {
		return -1
	}
	return 1
}

// Newer returns the newer of two timestamps.
// If they're within tolerance, returns t1 (arbitrary but consistent choice).
func Newer(t1, t2 time.Time) time.Time {
	if Equal(t1, t2) {
		return t1
	}
	if t1.After(t2) {
		return t1
	}
	return t2
}

// Older returns the older of two timestamps.
// If they're within tolerance, returns t1 (arbitrary but consistent choice).
func Older(t1, t2 time.Time) time.Time {
	if Equal(t1, t2) {
		return t1
	}
	if t1.Before(t2) {
		return t1
	}
	return t2
}

// Truncate truncates a timestamp to the given precision.
// This is useful for normalizing timestamps before comparison.
// Common precisions:
//   - time.Second for FAT32 compatibility
//   - time.Millisecond for network compatibility
//   - 100*time.Nanosecond for NTFS native precision
func Truncate(t time.Time, precision time.Duration) time.Time {
	return t.Truncate(precision)
}

// TruncateToSecond truncates a timestamp to second precision.
// This is the safest precision for cross-platform file synchronization.
func TruncateToSecond(t time.Time) time.Time {
	return t.Truncate(time.Second)
}
