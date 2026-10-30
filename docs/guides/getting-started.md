# Getting started

```bash
go get github.com/grokify/oscompat
```

Import the package you need. There is nothing to configure and nothing else to
install, because the module has no dependencies.

## Which package for which problem

| You need to | Use |
|-------------|-----|
| Keep two processes from writing a file at once | `fs.Lock` |
| Make a rename durable | `fs.SyncDir` |
| Refuse a file other users can read, or write one only the owner can | `fs.CheckPrivate`, `fs.WriteFilePrivate`, `fs.ChmodPrivate` |
| Know whether mode bits mean anything here | `fs.PermissionsEnforced` |
| Check a path from outside for `..` | `fs.ValidatePath`, `fs.SafeJoin` |
| Check a forward-slash path, such as one inside a repository or an archive | `fs.ValidateSlashPath` |
| Tell whether a path is absolute on any platform | `fs.IsAbsAny` |
| Confine a caller's path to a directory, following links | `fs.RealDir`, `fs.ResolveInRoot` |
| Find a config, data, cache or runtime directory | `paths.AppConfig`, `paths.AppData`, `paths.AppCache`, `paths.AppRuntime` |
| Stop a child process, or run one detached | `process.Signal`, `process.SetDetached` |
| Tell whether a recorded PID is still the same process | `process.Exists`, `process.StartTime` |
| Hand the terminal to another program | `process.Exec` |
| Generate an identifier | `id.Generate16`, `id.Generate32` |
| Compare file modification times | `tsync.Equal`, `tsync.After` |
| Talk between local processes | `localnet.Listen`, `localnet.Dial` |
| Write tests that pass on Windows | [the guide](portable-tests.md), and `testutil` |

## A first example

Replace a file so that a crash leaves the old one or the new one, and so that two
processes do not do it at once:

```go
import (
    "context"
    "os"
    "path/filepath"

    "github.com/grokify/oscompat/fs"
)

func replace(ctx context.Context, path string, data []byte) error {
    unlock, err := fs.Lock(ctx, path+".lock")
    if err != nil {
        return err
    }
    defer unlock()

    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
    if err != nil {
        return err
    }
    defer os.Remove(tmp.Name()) // a no-op once renamed

    if err := fs.ChmodPrivate(tmp); err != nil {
        return err
    }
    if _, err := tmp.Write(data); err != nil {
        return err
    }
    if err := tmp.Sync(); err != nil {
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    if err := os.Rename(tmp.Name(), path); err != nil {
        return err
    }
    return fs.SyncDir(dir)
}
```

None of it checks the operating system.

## Contributing a difference

If you find yourself writing `runtime.GOOS` or a build tag, add it here instead,
with a test, using only the standard library. See the [policy](../index.md#policy).
