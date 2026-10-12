# oscompat

Cross-platform operating-system compatibility utilities for Go. It holds the
code that differs between Windows and Unix-like systems (Linux, macOS), so that
other modules do not carry platform-specific files of their own.

```bash
go get github.com/grokify/oscompat
```

## Packages

| Package | Solves | Guide |
|---------|--------|-------|
| `fs` | Permissions, paths, executables, file locks, durable renames | [fs](guides/fs.md) |
| `process` | Signals, process groups, liveness, start time, exec | [process](guides/process.md) |
| `paths` | Config, data, cache and runtime directories | [paths](guides/paths.md) |
| `id` | Unique identifiers that do not collide on Windows | [id](guides/id.md) |
| `tsync` | Timestamp comparison across filesystems | [tsync](guides/tsync.md) |
| `localnet` | Local inter-process sockets | [localnet](guides/localnet.md) |
| `testutil` | Test helpers that behave the same everywhere | [testutil](guides/testutil.md) |

New here? Read [Getting started](guides/getting-started.md). Writing tests that
fail only on Windows? Read [Writing tests that pass on
Windows](guides/portable-tests.md).

## Policy

oscompat has **no dependencies**. It imports the Go standard library and
itself, and nothing else: no module outside the standard library, and not even
`golang.org/x/sys`. A test in the module enforces it. That is what makes it safe
for any module to import, including modules that keep their own dependency set
small on purpose, such as ones that handle key material.

In return, modules are encouraged to import oscompat instead of keeping
`_windows.go` and `_unix.go` files, build tags or `runtime.GOOS` checks of their
own. Code that differs by operating system belongs here once, tested on every
platform, so that callers read as portable code.

If a module needs a difference oscompat does not cover, add it here rather than
beside the caller, using only the standard library.

## Platform support

All packages are tested on Linux, macOS and Windows.

## Releases

See the [release notes](releases/v0.7.0.md) and the
[changelog](https://github.com/grokify/oscompat/blob/main/CHANGELOG.md).
