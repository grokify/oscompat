# fs: files, permissions, locks

`fs` holds what differs between Windows and Unix when working with files.

```go
import "github.com/grokify/oscompat/fs"
```

## Permissions

Mode bits are not enforced on Windows. Code that refuses a world-readable key
file, and tests that assert `0600`, are wrong there unless they ask first.

| Function | Behavior |
|----------|----------|
| `PermissionsEnforced()` | Whether the platform enforces mode bits; false on Windows |
| `IsPrivate(info)` | Owner-only. Always true where modes are not enforced |
| `CheckPrivate(info)` | An error matching `ErrNotPrivate` that names the mode; nil where modes are not enforced |
| `ChmodPrivate(f)` | Restricts an open file to its owner (`0600`); a failure is ignored where modes are not enforced |
| `WriteFilePrivate`, `MkdirAllPrivate` | Create a file or directory for the owner only |
| `PrivateFilePerm`, `PrivateDirPerm`, `DefaultFilePerm`, `DefaultDirPerm` | The usual modes as constants |

Refusing a file other users can read:

```go
info, err := os.Stat(path)
if err != nil {
    return err
}
if err := fs.CheckPrivate(info); err != nil {
    return fmt.Errorf("%s: %w", path, err)
}
```

In a test, assert only where modes mean something:

```go
if fs.PermissionsEnforced() {
    info, _ := os.Stat(path)
    if err := fs.CheckPrivate(info); err != nil {
        t.Fatal(err)
    }
}
```

## File locks

```go
unlock, err := fs.Lock(ctx, "/var/lib/myapp/state.lock")
if err != nil {
    return err
}
defer unlock()
```

`Lock` takes an exclusive advisory lock on a file, creating it (owner-only) if
needed. It uses `flock` on Unix and `LockFileEx` on Windows.

- It waits until the lock is free or `ctx` ends, and then returns the context's
  error. Give it a deadline.
- The lock is held on an open handle, so the operating system releases it if the
  process dies, however it dies. A crashed holder never wedges everyone else.
- Two calls in the same process exclude each other, as two processes do.
- It is advisory: it excludes only callers that also take it.
- `unlock` may be called more than once.
- On a platform with no locking it returns `ErrLockUnsupported`. A caller that
  needs mutual exclusion should fail closed and not proceed without it.

Keep the lock file separate from the file it protects. Replacing the protected
file by rename would otherwise swap the file the lock is held on.

## Durable renames

```go
if err := fs.SyncDir(dir); err != nil {
    return err
}
```

A rename is durable across a crash only once the directory is flushed. `SyncDir`
does that on Unix. On Windows, and wherever the platform has no way to ask, it
does nothing and returns nil: a rename there is atomic and nothing stronger is
available. You get the strongest guarantee the platform offers without a
platform check.

The complete write-temp, flush, rename, flush-directory sequence is shown in
[Getting started](getting-started.md#a-first-example).

## Paths

| Function | Behavior |
|----------|----------|
| `ValidatePath(p)` | Rejects `..` traversal with either separator |
| `SafeJoin(base, rel)` | Joins and refuses a result that escapes `base` |
| `NormalizePath(p)` | Converts to forward slashes, for storage |
| `OSPath(p)` | Converts to the platform's separator |
| `PathEqual`, `PathHasPrefix` | Compare with the platform's case rules |
| `IsExecutable(path)` | Mode bits on Unix; extension on Windows |

## Related

- [Writing tests that pass on Windows](portable-tests.md)
