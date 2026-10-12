//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd

package fs

import "context"

// Lock is not supported on this platform and returns [ErrLockUnsupported]. A
// caller that needs mutual exclusion must fail closed.
func Lock(context.Context, string) (unlock func(), err error) {
	return nil, ErrLockUnsupported
}
