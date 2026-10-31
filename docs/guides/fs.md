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
| `ValidatePath(p)` | Rejects `..` traversal and absolute or rooted paths, with either separator |
| `ValidateSlashPath(p)` | Like `ValidatePath`, for a path that uses forward slashes everywhere; also rejects backslashes, colons, and empty or `.` elements |
| `IsAbsAny(p)` | Whether a path is absolute or rooted on any platform |
| `SafeJoin(base, rel)` | Joins and refuses a result that escapes `base` |
| `NormalizeLineEndings(s)`, `NormalizeLineEndingsBytes(b)` | Convert CRLF to LF; a lone CR is kept |
| `NormalizePath(p)` | Converts to forward slashes, for storage |
| `OSPath(p)` | Converts to the platform's separator |
| `PathEqual`, `PathHasPrefix` | Compare with the platform's case rules |
| `IsExecutable(path)` | Mode bits on Unix; extension on Windows |

### Paths from somewhere else

A path read from a document, an archive, a configuration file or another
system must stay inside its root whichever platform reads it. `filepath.IsAbs`
cannot check that: it answers for the platform it runs on. Windows does not
treat `/etc/passwd` as absolute, because it names a folder on the current
drive, and Unix does not treat `\etc\passwd`, `\\server\share` or `C:x` as
anything but a relative file name.

`IsAbsAny` gives the same answer everywhere. It is true for a path that starts
with `/` or `\`, including UNC paths, and for one that starts with a drive
letter, including the drive-relative `C:x`.

| Path | `filepath.IsAbs` on Unix | on Windows | `IsAbsAny` |
|------|--------------------------|------------|------------|
| `/etc/passwd` | true | false | true |
| `\etc\passwd` | false | false | true |
| `\\server\share` | false | true | true |
| `C:\x`, `C:/x` | false | true | true |
| `C:x` | false | false | true |
| `docs/a.md` | false | false | false |

Choose the check by what the path names:

- **A file on this machine**, such as a command-line argument: `filepath.IsAbs`
  and `filepath.Join` are right, because the platform's rules apply.
- **A path relative to a root, from anywhere**, with either separator:
  `ValidatePath`, then `SafeJoin` to read it.
- **A path from a caller that may also follow links, such as a tool argument:**
  `ResolveInRoot` to get a path, or `OpenInRoot` to read the file (see below).
- **A path that uses forward slashes everywhere**, such as a path inside a
  repository, an archive or a URL: `ValidateSlashPath`. It also refuses
  backslashes, colons, and empty or `.` elements, so the path names the same
  file on every platform.

```go
if err := fs.ValidateSlashPath(src.Path); err != nil {
    return fmt.Errorf("source path %q: %w", src.Path, err)
}
data, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(src.Path)))
```

The errors are `ErrEmptyPath`, `ErrAbsolutePath`, `ErrPathTraversal` and, from
`ValidateSlashPath` only, `ErrInvalidPath`.

### Confining a caller's path to a directory

`ResolveInRoot(root, p)` turns a path supplied by a caller, such as an MCP tool
argument or an upload path, into a real path inside `root`. It returns an error
wrapping `ErrOutsideRoot` for an empty path, a NUL, a `..` element with either
separator, a rooted or drive-qualified path that is not inside the root, and
any path that leads outside once symbolic links are followed. The file need not
exist, so a caller can create one in place.

```go
root, err := fs.RealDir(uploadRoot) // once, at start-up
if err != nil {
    return err
}
path, err := fs.ResolveInRoot(root, req.Path)
if errors.Is(err, fs.ErrOutsideRoot) {
    return fmt.Errorf("upload path: %w", err)
}
```

The check runs at the moment of the call. A link created afterwards is not
seen, so open the returned path promptly and do not treat it as a lock. If the
next step is to read the file, use `OpenInRoot` instead.

### Opening a file inside a directory

`OpenInRoot(root, name)` checks the name as `ResolveInRoot` does and then opens
it through `os.Root`, so the operating system refuses a symbolic link that leads
out of the root at the moment of the open. There is no gap between the check and
the open for a link to be swapped into, and the returned `*os.File` stays bound
to the file that was opened, whatever happens to the path afterwards.

```go
root, err := fs.RealDir(uploadRoot) // once, at start-up
if err != nil {
    return err
}
f, err := fs.OpenInRoot(root, req.Path)
if errors.Is(err, fs.ErrOutsideRoot) {
    return fmt.Errorf("upload path: %w", err)
}
if err != nil {
    return err // a missing file satisfies errors.Is(err, os.ErrNotExist)
}
defer f.Close()
```

To read the file, and not only open it, use the pair that checks the handle and
caps the read:

```go
f, err := fs.OpenRegularInRoot(root, req.Path, maxBytes)
if err != nil {
    // fs.ErrOutsideRoot, os.ErrNotExist, fs.ErrNotRegular or fs.ErrTooLarge,
    // each testable with errors.Is
    return fmt.Errorf("upload path: %w", err)
}
defer f.Close()

body := fs.CapReader(f, maxBytes) // a file that grows after the check
data, err := io.ReadAll(body)      // ends with fs.ErrTooLarge
```

`OpenRegularInRoot` opens as `OpenInRoot` does and returns the file only if the
handle it holds is a regular file of at most `maxBytes`. The check is made on
the open handle, so it describes the file you will read, and the file is closed
for you when it fails. `CapReader` is `io.LimitReader` that reports
`ErrTooLarge` when the stream is longer than the limit, instead of stopping
silently, so an oversized file cannot pass as a complete, shorter one. Bytes
within the limit are delivered before the error. Hand the capped reader to
anything that takes an `io.Reader`, such as an upload.

Use `OpenInRoot` when you will read the file, and `ResolveInRoot` when you need a
path: to create a file that does not exist yet, or to hand to code that opens
paths itself. Three things to know:

- **Links.** A link whose target is relative and stays inside the root is
  followed. A link whose target is an absolute path is refused with
  `ErrOutsideRoot`, even if the target is inside the root, because `os.Root`
  does not follow absolute links. `ResolveInRoot` accepts such a link.
- **Absolute names.** One is accepted only if it is spelled inside `root`, with
  `root` written the same way. Pass the result of `RealDir` as `root`.
- **The read.** The open is confined, but the size of the file can still change
  while you read it. That is why `OpenRegularInRoot` is paired with `CapReader`:
  one checks the file you hold, the other enforces the limit on what is read.

## Related

- [Writing tests that pass on Windows](portable-tests.md)
