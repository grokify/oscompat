# process: signals, liveness, exec

```go
import "github.com/grokify/oscompat/process"
```

## Stopping a process

Windows has no `SIGTERM`.

```go
err := process.Signal(pid)         // SIGTERM on Unix, Kill on Windows
err = process.FindAndSignal(pid)   // finds the process first
```

## Detached children

```go
cmd := exec.Command("myapp", "serve")
process.SetDetached(cmd) // its own process group on Unix
err := cmd.Start()
```

## Is it still the same process?

A file that records a PID goes stale when the process dies, and the system can
give the same PID to an unrelated process. Compare the start time.

```go
ok, err := process.Exists(rec.PID)

started, err := process.StartTime(rec.PID)
switch {
case errors.Is(err, process.ErrNotFound):
    // gone
case err == nil && started.Sub(rec.StartedAt).Abs() > 2*time.Second:
    // the PID was reused by a different process
}
```

| Function | Behavior |
|----------|----------|
| `Exists(pid)` | Whether the process is running. A process owned by another user counts as running |
| `StartTime(pid)` | The start time, or `ErrNotFound` |

On Unix the start time comes from `ps` and has one-second resolution, so allow a
tolerance when comparing. On Windows it is the creation time. A PID of zero or
less is rejected, because on Unix it addresses a process group.

## Replacing the current process

```go
err := process.Exec(path, []string{"claude", "--resume", id})
```

Unix replaces the process. Windows cannot, so it runs the program as a child with
the terminal attached and exits with its status. It returns only on failure.

## Opening a URL or a file

Each platform opens things in the user's default application with a different
program.

```go
err := process.OpenDefault("http://localhost:9401") // the default browser
err = process.OpenDefault("/tmp/report.html")        // the program registered for .html
```

It runs `open` on macOS, `xdg-open` on Linux and the BSDs, and the URL protocol
handler of `rundll32` on Windows, which avoids `cmd.exe` and its handling of
characters such as `&`. The target is passed as one argument. It returns once the
opener has started and does not wait for the application.

An empty target, or one starting with `-` that the opener could read as an
option, is refused. Name a file that starts with `-` as `./-name`.

