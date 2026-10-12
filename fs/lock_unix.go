//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// Lock takes an exclusive advisory lock on the file at path, creating it with
// owner-only permissions if it does not exist, and returns the function that
// releases it.
//
// It waits until the lock is free or ctx ends, and in the second case returns
// the context's error. The lock is held on an open handle, so it is released if
// the process exits, however it exits. Locks conflict between handles, so two
// calls in one process exclude each other as two processes do.
//
// The lock is advisory: it excludes only callers that also take it. The returned
// function is safe to call more than once.
//
// On a platform without file locking it returns [ErrLockUnsupported].
func Lock(ctx context.Context, path string) (unlock func(), err error) {
	//nolint:gosec // G703: the lock file path is the caller's by design; Lock is for locking whatever file it is given
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, PrivateFilePerm)
	if err != nil {
		return nil, fmt.Errorf("oscompat/fs: opening lock file: %w", err)
	}
	fd := int(f.Fd()) //nolint:gosec // G115: a file descriptor fits an int
	for {
		err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			var once sync.Once
			return func() {
				once.Do(func() {
					_ = syscall.Flock(fd, syscall.LOCK_UN) //nolint:errcheck // closing the descriptor releases the lock regardless
					_ = f.Close()                          //nolint:errcheck // the kernel releases the lock with the descriptor
				})
			}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EINTR) {
			_ = f.Close() //nolint:errcheck // the lock error is the one to report
			return nil, fmt.Errorf("oscompat/fs: locking %s: %w", path, err)
		}
		select {
		case <-ctx.Done():
			_ = f.Close() //nolint:errcheck // the context error is the one to report
			return nil, ctx.Err()
		case <-time.After(lockPoll):
		}
	}
}
