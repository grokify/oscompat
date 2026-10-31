# Writing Tests That Pass on Windows

Most Windows-only test failures in Go projects come from a handful of
platform differences. They are cheap to avoid when writing the code and
tedious to diagnose from a CI log afterward. This guide lists them with the
portable alternative for each.

## Wall-clock ties

**The difference:** `time.Now()` on Linux and macOS has nanosecond
resolution. On Windows the clock advances in steps of roughly 0.5 to 15.6 ms,
so values taken in quick succession are often identical.

**What breaks:**

- Ordering rows by `created_at` (or any `time.Now()` column), then asserting
  the exact order. Equal timestamps fall through to the next sort key.
- `if after.After(before)` between two operations that run back to back.
- `if elapsed > 0` for a measured duration.

**The pattern that bites hardest:** `ORDER BY created_at, id` where `id` is a
random UUID. On Unix the timestamps differ and the test always passes. On
Windows they tie, `id` decides, and the test fails about half the time for two
rows.

**What to do:**

1. Give ordered records a tie-breaker that follows insertion order, not a
   random one: a UUIDv7 (`uuid.NewV7()` from `github.com/google/uuid`), a
   database sequence, or an explicit sequence column. UUIDv7 values created
   by one process keep increasing even within a millisecond, and they sort
   correctly as `uuid` in PostgreSQL and as text in SQLite.
2. If the order within a tie truly does not matter to callers, say so in the
   doc comment and make the test compare as a set (`slices.Sort` both sides),
   so the test checks what the code promises.
3. Assert `d >= 0`, not `d > 0`, for durations. `testutil.AssertDurationNonNegative`
   does this.
4. Never `time.Sleep` for a few milliseconds to force a different timestamp.
   It lowers the odds without removing the failure.

## Running your program in a test

To exercise a command over its real standard input and output, run the test
binary itself as the command instead of building one with `go build`. There is
no build step and no `.exe` suffix to add on Windows.

```go
const runAsCommand = "MYAPP_RUN_AS_COMMAND"

func TestMain(m *testing.M) {
    if os.Getenv(runAsCommand) == "1" {
        os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
    }
    os.Exit(m.Run())
}

func TestServesOverStdio(t *testing.T) {
    cmd := exec.Command(os.Args[0], "serve")
    cmd.Env = append(os.Environ(), runAsCommand+"=1")
    // ... connect to cmd's stdin and stdout
}
```

## Path separators and permissions

- Build paths with `filepath.Join`, and compare with `filepath.ToSlash` when a
  test expects a fixed string. `fs.NormalizePath` and `fs.OSPath` convert.
- Do not check a path that came from a document or an archive with
  `filepath.IsAbs`: it gives a different answer on each platform. Use
  `fs.ValidatePath`, `fs.ValidateSlashPath` or `fs.IsAbsAny`, which give the
  same answer everywhere. To confine a caller's path to a directory, including
  links that lead out of it, use `fs.ResolveInRoot`, or `fs.OpenInRoot` to read
  the file, with `fs.OpenRegularInRoot` and `fs.CapReader` to bound it.
- Keep path test cases out of `runtime.GOOS` branches. A case that runs only
  where it passes cannot show that another platform disagrees; run every case
  everywhere and give each the answer it should have on all of them.
- File mode bits other than read-only are not enforced on Windows. Do not
  assert `0600` after a write, and skip permission-denied tests with
  `fs.PermissionsEnforced()` instead of a `runtime.GOOS` check, so the
  condition is named once. `fs.CheckPrivate` is the owner-only assertion that
  is correct on both. `fs.IsExecutable` replaces mode checks.
- Home and config directories use `USERPROFILE` and `APPDATA`, not `HOME`.
  `testutil` and `paths` set and resolve both.

## Files that are still open

Windows refuses to delete or rename an open file. Close files and databases
before `t.TempDir` cleanup runs, using `t.Cleanup` in reverse order of
opening, and check the `Close` error.

This applies to a test that simulates a crash by returning early, for example
from a hook between "temporary file created" and "renamed". A real crash closes
the process's handles; the simulation must close the handle too, and leave the
file, or Windows cannot remove it and the next step and the temporary directory
cleanup both fail.

## File locks and directory flushes

Use `fs.Lock` rather than `syscall.Flock`, which does not exist on Windows, and
`fs.SyncDir` rather than opening a directory to `Sync` it, which fails there.
Both are tested against a second process in this module.

## Process and signals

There is no `SIGTERM`, and process groups work differently. Use
`process.Signal` and `process.SetDetached` rather than `syscall` directly.

## Line endings

Git may check out text files with CRLF. Read fixtures with
`testutil.ReadFileLF` and compare golden files with `testutil.AssertGolden`,
which ignore CRLF versus LF. In a repository that compares files byte for
byte, also set `* text=auto eol=lf` in `.gitattributes`.

## Catching these before CI

- Run the test suite with `-count=20` for any test that orders rows or
  compares times. Ties show up on Unix too when the operations are fast
  enough, and repetition exposes order assumptions that happen to pass.
- Keep a Windows job in CI from the first commit rather than adding it when
  the project is mature.
- Review any `ORDER BY` with a non-deterministic final key.
