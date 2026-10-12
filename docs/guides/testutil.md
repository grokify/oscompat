# testutil: tests that behave the same everywhere

```go
import "github.com/grokify/oscompat/testutil"
```

## Home directory

```go
func TestConfig(t *testing.T) {
    home := testutil.SetTempHomeDir(t)
    // os.UserHomeDir() now returns home on every platform
}
```

It sets both `HOME` and `USERPROFILE`, because Windows reads the second, and
removes the directory when the test ends.

## Durations

The Windows clock advances in steps of about 15.6 ms, so a fast operation can
measure as zero.

```go
testutil.AssertDurationNonNegative(t, elapsed)
if testutil.DurationNonNegative(elapsed) { /* ... */ }
```

Assert that a duration is not negative, never that it is positive.

## Capturing output

```go
out, err := testutil.CaptureStdout(func() { fmt.Println("hi") })
stdout, stderr, err := testutil.CaptureOutput(func() { /* ... */ })
```

The pipe is read concurrently, which avoids the deadlock a large write causes on
Windows when a test reads only after the function returns.

See also [Writing tests that pass on Windows](portable-tests.md).
