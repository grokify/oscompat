//go:build windows

package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	lockfileExclusiveLock   = 0x2
	lockfileFailImmediately = 0x1
	errorLockViolation      = syscall.Errno(33)
)

// Lock takes an exclusive advisory lock on the file at path, creating it if it
// does not exist, and returns the function that releases it.
//
// It waits until the lock is free or ctx ends, and in the second case returns
// the context's error. The lock is held on an open handle, so it is released if
// the process exits, however it exits. Locks conflict between handles, so two
// calls in one process exclude each other as two processes do.
//
// The lock is advisory: it excludes only callers that also take it. The returned
// function is safe to call more than once.
func Lock(ctx context.Context, path string) (unlock func(), err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, PrivateFilePerm)
	if err != nil {
		return nil, fmt.Errorf("oscompat/fs: opening lock file: %w", err)
	}
	h := syscall.Handle(f.Fd())
	for {
		var ol syscall.Overlapped
		r, _, e := procLockFileEx.Call(uintptr(h), lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, uintptr(unsafe.Pointer(&ol)))
		if r != 0 {
			var once sync.Once
			return func() {
				once.Do(func() {
					var ol syscall.Overlapped
					_, _, _ = procUnlockFileEx.Call(uintptr(h), 0, 1, 0, uintptr(unsafe.Pointer(&ol)))
					_ = f.Close() //nolint:errcheck // the lock is released with the handle
				})
			}, nil
		}
		if !errors.Is(e, errorLockViolation) {
			_ = f.Close() //nolint:errcheck // the lock error is the one to report
			return nil, fmt.Errorf("oscompat/fs: locking %s: %w", path, e)
		}
		select {
		case <-ctx.Done():
			_ = f.Close() //nolint:errcheck // the context error is the one to report
			return nil, ctx.Err()
		case <-time.After(lockPoll):
		}
	}
}
