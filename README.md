# oscompat

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

Cross-platform OS compatibility utilities for Go.

This package provides abstractions for OS-specific behaviors that differ between Windows and Unix-like systems (Linux, macOS).

## Installation

```bash
go get github.com/grokify/oscompat
```

## Packages

### id

Cross-platform unique identifier generation using `crypto/rand`.

**Why this exists:** Time-based ID generation (e.g., `time.Now().String()`) is unreliable on Windows due to the system clock's coarse resolution (~15.6ms vs nanoseconds on Unix). This causes ID collisions when generating multiple IDs in rapid succession.

```go
import "github.com/grokify/oscompat/id"

// Generate a 16-character hex ID (8 random bytes)
spanID := id.Generate16()

// Generate a 32-character hex ID (16 random bytes)
traceID := id.Generate32()

// Generate custom length (N bytes = 2N hex characters)
customID := id.Generate(4) // 8-character hex string
```

### process

Cross-platform process management utilities.

**Why this exists:** Unix and Windows handle process signals differently:

- Unix has `SIGTERM` for graceful termination; Windows does not
- Unix uses `Setpgid` for process group management; Windows does not

```go
import "github.com/grokify/oscompat/process"

// Configure a command to run detached from parent
cmd := exec.Command("myapp")
process.SetDetached(cmd)

// Send termination signal (SIGTERM on Unix, Kill on Windows)
err := process.Signal(pid)

// Find and signal a process
err := process.FindAndSignal(pid)
```

### paths

Cross-platform configuration and data directory resolution.

**Why this exists:** Applications need platform-appropriate directories:

- Unix: XDG Base Directory Specification (~/.config, ~/.local/share)
- macOS: ~/Library/Application Support, ~/Library/Caches
- Windows: %APPDATA%, %LOCALAPPDATA%, %ProgramData%

```go
import "github.com/grokify/oscompat/paths"

// Get app-specific config directory (creates if needed)
configDir, err := paths.AppConfig("myapp")
// Unix:    ~/.config/myapp
// macOS:   ~/Library/Application Support/myapp
// Windows: %APPDATA%\myapp

// Get app-specific data directory
dataDir, err := paths.AppData("myapp")

// Get app-specific cache directory
cacheDir, err := paths.AppCache("myapp")

// Get system-wide config directory
sysConfig, err := paths.SystemConfig()
// Unix:    /etc
// Windows: %ProgramData%
```

### fs

Cross-platform filesystem utilities.

**Why this exists:**

- File permissions (0755, 0644) are Unix-centric; Windows ignores them
- Path separators differ (/ vs \)
- Path traversal validation needs to handle both separators

```go
import "github.com/grokify/oscompat/fs"

// Validate path for traversal attacks (handles both / and \)
err := fs.ValidatePath("foo/../../../etc/passwd") // returns ErrPathTraversal

// Normalize path to forward slashes (for storage)
normalized := fs.NormalizePath(`foo\bar\baz`) // "foo/bar/baz"

// Convert to OS-specific path
osPath := fs.OSPath("foo/bar/baz") // "foo\bar\baz" on Windows

// Safe path joining with traversal protection
fullPath, err := fs.SafeJoin("/base", "relative/path")

// Create directory with default permissions
err := fs.MkdirAll("/path/to/dir", 0) // uses DefaultDirPerm (0755)

// Create private directory (owner-only)
err := fs.MkdirAllPrivate("/path/to/private")

// Write file with default permissions
err := fs.WriteFile("file.txt", data, 0) // uses DefaultFilePerm (0644)

// Write private file (owner-only)
err := fs.WriteFilePrivate("secret.txt", data)
```

### tsync

Cross-platform timestamp utilities for file synchronization.

**Why this exists:** Filesystem timestamp precision varies:

- Windows NTFS: ~100ns precision, sometimes rounded to 2 seconds
- Linux ext4/XFS: nanosecond precision
- FAT32: 2-second precision
- Network drives: varies widely

```go
import "github.com/grokify/oscompat/tsync"

// Compare timestamps with platform-appropriate tolerance (1 second)
if tsync.Equal(file1.ModTime(), file2.ModTime()) {
    // Files have same modification time (within tolerance)
}

// Check if file1 is newer than file2 (beyond tolerance)
if tsync.After(file1.ModTime(), file2.ModTime()) {
    // file1 is definitively newer
}

// Compare with custom tolerance
if tsync.EqualWithTolerance(t1, t2, tsync.FAT32Tolerance) {
    // Equal within 2-second tolerance
}

// Truncate timestamp for cross-platform storage
normalized := tsync.TruncateToSecond(time.Now())
```

### localnet

Cross-platform local network communication (IPC).

**Why this exists:** Unix domain sockets don't exist on Windows (or have limited support).

- Unix: Uses Unix domain sockets in /tmp or $XDG_RUNTIME_DIR
- Windows: Uses TCP on localhost with port discovery via file

```go
import "github.com/grokify/oscompat/localnet"

// Server: Create local listener
listener, err := localnet.Listen("myapp")
if err != nil {
    return err
}
defer listener.Close()

for {
    conn, err := listener.Accept()
    // handle connection...
}

// Client: Connect to local server
conn, err := localnet.Dial("myapp")
if err != nil {
    return err
}
defer conn.Close()
// use conn...

// Cleanup stale socket (e.g., after crash)
localnet.Cleanup("myapp")
```

## Platform Support

All packages are tested on:

- Linux (ubuntu-latest)
- macOS (macos-latest)
- Windows (windows-latest)

## License

MIT

 [build-status-svg]: https://github.com/grokify/oscompat/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/grokify/oscompat/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/grokify/oscompat/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/grokify/oscompat/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/oscompat
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/oscompat
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/oscompat
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/oscompat
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Foscompat
 [loc-svg]: https://tokei.rs/b1/github/grokify/oscompat
 [repo-url]: https://github.com/grokify/oscompat
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/oscompat/blob/master/LICENSE