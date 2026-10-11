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

## Path separators and permissions

- Build paths with `filepath.Join`, and compare with `filepath.ToSlash` when a
  test expects a fixed string. `fs.NormalizePath` and `fs.OSPath` convert.
- File mode bits other than read-only are not enforced on Windows. Do not
  assert `0600` after a write, and skip permission-denied tests with
  `runtime.GOOS == "windows"`. `fs.IsExecutable` replaces mode checks.
- Home and config directories use `USERPROFILE` and `APPDATA`, not `HOME`.
  `testutil` and `paths` set and resolve both.

## Files that are still open

Windows refuses to delete or rename an open file. Close files and databases
before `t.TempDir` cleanup runs, using `t.Cleanup` in reverse order of
opening, and check the `Close` error.

## Process and signals

There is no `SIGTERM`, and process groups work differently. Use
`process.Signal` and `process.SetDetached` rather than `syscall` directly.

## Line endings

Git may check out text files with CRLF. Compare golden files after
normalizing `\r\n` to `\n`, or mark them `-text` in `.gitattributes`.

## Catching these before CI

- Run the test suite with `-count=20` for any test that orders rows or
  compares times. Ties show up on Unix too when the operations are fast
  enough, and repetition exposes order assumptions that happen to pass.
- Keep a Windows job in CI from the first commit rather than adding it when
  the project is mature.
- Review any `ORDER BY` with a non-deterministic final key.
