# paths: where files go

```go
import "github.com/grokify/oscompat/paths"
```

Each application directory follows the platform's convention and is created if
it does not exist.

| Function | Linux | macOS | Windows |
|----------|-------|-------|---------|
| `AppConfig(name)` | `~/.config/<name>` | `~/Library/Application Support/<name>` | `%APPDATA%\<name>` |
| `AppData(name)` | `~/.local/share/<name>` | `~/Library/Application Support/<name>` | `%LOCALAPPDATA%\<name>` |
| `AppCache(name)` | `~/.cache/<name>` | `~/Library/Caches/<name>` | `%LOCALAPPDATA%\cache\<name>` |
| `AppRuntime(name)` | `$XDG_RUNTIME_DIR/<name>` or `/tmp/runtime-<uid>/<name>` | `~/Library/Application Support/<name>` | `%LOCALAPPDATA%\run\<name>` |
| `SystemAppConfig(name)` | `/etc/<name>` | `/etc/<name>` | `%ProgramData%\<name>` |

```go
cfg, err := paths.AppConfig("myapp")
if err != nil {
    return err
}
```

- The runtime directory is created `0700`; the others `0755`. On macOS and
  Windows it sits beside the application's other files rather than in a
  temporary location, so remove stale files from it at start-up.
- Use the runtime directory for sockets and PID files, which should not survive
  a reboot.
- `Home()` returns the user's home directory or `ErrHomeNotFound`.
- An empty name is `ErrInvalidAppName`.
- `UserConfig`, `UserData`, `UserCache`, `UserRuntime` and `SystemConfig` return
  the base directories without an application name or creating anything.

In tests, set the home directory with
[`testutil.SetTempHomeDir`](testutil.md).
