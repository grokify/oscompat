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

## Fixtures and golden files

Git may check out text files with CRLF on Windows, so a fixture read from disk
can differ from the same bytes generated in the test.

```go
want := testutil.ReadFileLF(t, "testdata/input.json") // CRLF becomes LF

var update = flag.Bool("update", false, "rewrite the golden files")

testutil.AssertGolden(t, "testdata/out.json", got, *update)
```

`AssertGolden` compares `got` with the file ignoring CRLF versus LF, and
writes the file instead when `update` is set. Produce `got` with LF endings.
For a string or bytes already in memory, use `fs.NormalizeLineEndings` or
`fs.NormalizeLineEndingsBytes`.

See also [Writing tests that pass on Windows](portable-tests.md).
