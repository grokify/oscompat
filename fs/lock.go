package fs

import (
	"errors"
	"time"
)

var (
	// ErrLockUnsupported is returned by [Lock] on a platform with no advisory
	// file locking. Callers that need mutual exclusion must fail closed rather
	// than proceed without it.
	ErrLockUnsupported = errors.New("oscompat/fs: advisory file locking is not supported on this platform")
)

// lockPoll is how often a waiting [Lock] retries, so that a cancelled context
// ends the wait promptly.
const lockPoll = 5 * time.Millisecond
