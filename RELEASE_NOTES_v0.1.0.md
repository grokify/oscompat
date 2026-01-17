# Release Notes v0.1.0

This is the initial release of **oscompat**, a cross-platform OS compatibility library for Go.

## Overview

oscompat provides abstractions for OS-specific behaviors that differ between Windows and Unix-like systems (Linux, macOS). Rather than scattering `runtime.GOOS` checks throughout application code, oscompat encapsulates these differences in reusable, well-tested packages.

## New Packages

### id - Unique Identifier Generation

Cryptographically secure ID generation using `crypto/rand`.

**Problem solved:** Time-based ID generation (e.g., `time.Now().UnixNano()`) is unreliable on Windows due to the system clock's coarse resolution (~15.6ms vs nanoseconds on Unix), causing ID collisions when generating multiple IDs rapidly.

**Key functions:**

- `Generate(byteLen int)` - Generate custom-length hex ID
- `Generate16()` - Generate 16-character hex ID (8 random bytes)
- `Generate32()` - Generate 32-character hex ID (16 random bytes)

### process - Process Management

Cross-platform process signal and detachment handling.

**Problem solved:** Unix has `SIGTERM` for graceful termination and `Setpgid` for process groups; Windows has neither.

**Key functions:**

- `SetDetached(cmd)` - Configure command to run detached from parent
- `Signal(pid)` - Send termination signal (SIGTERM on Unix, Kill on Windows)
- `FindAndSignal(pid)` - Find process by PID and send termination signal

### paths - Directory Resolution

Platform-appropriate configuration and data directory resolution.

**Problem solved:** Applications need standard directories that vary by platform:

| Directory Type | Linux (XDG) | macOS | Windows |
|---------------|-------------|-------|---------|
| User Config | ~/.config | ~/Library/Application Support | %APPDATA% |
| User Data | ~/.local/share | ~/Library/Application Support | %LOCALAPPDATA% |
| User Cache | ~/.cache | ~/Library/Caches | %LOCALAPPDATA%\cache |
| System Config | /etc | /etc | %ProgramData% |

**Key functions:**

- `Home()` - User's home directory
- `AppConfig(name)`, `AppData(name)`, `AppCache(name)` - App-specific directories (auto-creates)
- `UserConfig()`, `UserData()`, `UserCache()`, `UserRuntime()` - Base user directories
- `SystemConfig()`, `SystemAppConfig(name)` - System-wide directories

### fs - Filesystem Utilities

Cross-platform path handling, permissions, and validation.

**Problems solved:**

- File permissions (0755, 0644) are Unix-centric; Windows ignores them
- Path separators differ (/ vs \)
- Path traversal attacks must handle both separator types
- Case sensitivity differs (Windows is case-insensitive)

**Key functions:**

- `ValidatePath()`, `ValidatePathStrict()` - Detect directory traversal attacks
- `NormalizePath()`, `OSPath()` - Convert between normalized and OS-specific paths
- `SafeJoin()` - Path joining with traversal protection
- `MkdirAll()`, `WriteFile()` - Directory/file creation with sensible defaults
- `IsCaseSensitive()`, `PathEqual()`, `PathHasPrefix()` - Case-aware path comparison
- `Birthtime()`, `BirthtimeInfo()` - File creation time (where supported)

**Permission constants:** `DefaultDirPerm` (0755), `DefaultFilePerm` (0644), `PrivateDirPerm` (0700), `PrivateFilePerm` (0600)

### tsync - Timestamp Utilities

Tolerant timestamp comparison for file synchronization.

**Problem solved:** Filesystem timestamp precision varies dramatically:

| Filesystem | Precision |
|------------|-----------|
| Linux ext4/XFS | nanosecond |
| Windows NTFS | ~100ns (sometimes 2s) |
| FAT32 | 2 seconds |
| Network drives | varies widely |

**Key functions:**

- `Equal()`, `Before()`, `After()` - Compare with default 1-second tolerance
- `EqualWithTolerance()`, `BeforeWithTolerance()`, `AfterWithTolerance()` - Custom tolerance
- `Compare()`, `CompareWithTolerance()` - Three-way comparison (-1/0/+1)
- `Truncate()`, `TruncateToSecond()` - Normalize precision for storage
- `FromTimespec()`, `FromTimeval()` - Convert syscall time structures

**Tolerance constants:** `DefaultTolerance` (1s), `FAT32Tolerance` (2s), `HighPrecisionTolerance` (100ms)

### localnet - Local IPC

Cross-platform inter-process communication.

**Problem solved:** Unix domain sockets don't exist on Windows (or have limited support).

| Platform | Implementation |
|----------|----------------|
| Unix/Linux/macOS | Unix domain sockets in /tmp or $XDG_RUNTIME_DIR |
| Windows | TCP on localhost with port discovery via file |

**Key functions:**

- `Listen(name)` - Create local listener for IPC
- `Dial(name)` - Connect to local server
- `SocketPath(name)` - Get path/address for documentation
- `Cleanup(name)` - Remove stale socket/port files after crash

## Platform Support

All packages are tested on:

- Linux (ubuntu-latest)
- macOS (macos-latest)
- Windows (windows-latest)

CI runs on Go 1.24.x and 1.25.x.

## Dependencies

**Zero external dependencies.** oscompat uses only the Go standard library.

## Test Coverage

| Package | Coverage |
|---------|----------|
| fs | 90.5% |
| id | 83.3% |
| localnet | 83.3% |
| paths | 80.3% |
| process | 83.3% |
| tsync | 100.0% |

## Installation

```bash
go get github.com/grokify/oscompat
```

## License

MIT
