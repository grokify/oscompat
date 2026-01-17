# oscompat Roadmap

Cross-platform OS compatibility utilities for Go.

## Current Status

### v0.1.0 (Implemented)

- [x] **id** - Cryptographically secure unique ID generation
  - Fixes Windows clock resolution issues (~15.6ms vs nanoseconds)
  - `Generate(byteLen)`, `Generate16()`, `Generate32()`

- [x] **process** - Process management utilities
  - `SetDetached(cmd)` - Process group on Unix, CREATE_NEW_PROCESS_GROUP on Windows
  - `Signal(pid)` - SIGTERM on Unix, Kill on Windows
  - `FindAndSignal(pid)` - Combined find and signal

- [x] **paths** - Configuration directory resolution
  - `UserConfig()`, `UserData()`, `UserCache()` - XDG/macOS Library/Windows AppData
  - `SystemConfig()` - /etc, /Library/Application Support, ProgramData
  - `AppConfig(name)`, `AppData(name)`, `AppCache(name)` - App-specific directories

- [x] **fs** - Filesystem utilities
  - `DefaultDirPerm()`, `DefaultFilePerm()`, `PrivateDirPerm()`, `PrivateFilePerm()`
  - `ValidatePath()` - Path traversal attack prevention
  - `NormalizePath()`, `OSPath()` - Path separator conversion

- [x] **tsync** - Timestamp synchronization utilities
  - `Equal()`, `EqualWithTolerance()` - Time comparison with tolerance
  - `Before()`, `After()` - Tolerant comparisons
  - `DefaultTolerance` - 1 second for cross-platform compatibility

- [x] **localnet** - Local IPC abstraction
  - `Listen(name)`, `Dial(name)` - Unix sockets on Unix, TCP on Windows
  - `Cleanup(name)` - Socket file cleanup
  - `SocketPath(name)` - Get the socket/address path

## Projects Using oscompat

| Project | Packages Used | Status |
|---------|---------------|--------|
| omniretrieve | id | Integrated |
| omnivault | process | Integrated |
| vaultguard | paths | Integrated |
| omnistorage | tsync | Integrated |
| omniproxy | process | Planned |
| godaemonkit | paths | Planned |
| mogo | fs, tsync | Planned |
| go-util | fs | Planned |

## Planned Enhancements

### v0.2.0 - Enhanced Filesystem & Time

Based on analysis of grokify projects, the following enhancements are planned:

#### fs - Case-Insensitive Path Comparison

**Priority:** HIGH

**Problem:** Windows filesystem is case-insensitive, requiring special handling for path comparisons.

**Affected Projects:**

| Project | File | Pattern |
|---------|------|---------|
| go-util | `fs/fs.go` | `runtime.GOOS == "windows"` with `strings.ToLower()` |

**Proposed API:**

```go
// CaseInsensitiveEqual compares two paths case-insensitively on Windows,
// case-sensitively on Unix.
func CaseInsensitiveEqual(path1, path2 string) bool

// CaseInsensitiveHasPrefix checks if path has the given prefix,
// using case-insensitive comparison on Windows.
func CaseInsensitiveHasPrefix(path, prefix string) bool

// IsCaseSensitive returns true if the current OS has case-sensitive paths.
// Returns false on Windows, true on Unix/macOS.
func IsCaseSensitive() bool
```

#### fs - File Birthtime (Creation Time)

**Priority:** MEDIUM

**Problem:** macOS provides file birth time via `Birthtimespec` in `syscall.Stat_t`, but this field doesn't exist on Linux. Need cross-platform abstraction.

**Affected Projects:**

| Project | File | Pattern |
|---------|------|---------|
| mogo | `os/osutil/osutil_darwin.go` | `//go:build darwin` with `Birthtimespec` |

**Proposed API:**

```go
// Birthtime returns the file's creation time if available.
// Returns the modification time as fallback on platforms without birthtime support.
// macOS: Uses Birthtimespec
// Windows: Uses CreationTime from Win32 API
// Linux: Falls back to ModTime (birthtime not reliably available)
func Birthtime(path string) (time.Time, error)

// BirthtimeSupported returns true if the platform supports file birthtime.
func BirthtimeSupported() bool
```

#### tsync - Syscall Time Conversion

**Priority:** MEDIUM

**Problem:** Converting `syscall.Timespec` and `syscall.Timeval` to `time.Time` is a common pattern that differs by platform.

**Affected Projects:**

| Project | File | Pattern |
|---------|------|---------|
| mogo | `time/timeutil/syscall.go` | Manual Timespec/Timeval conversion |

**Proposed API:**

```go
// FromTimespec converts a syscall.Timespec to time.Time.
// Handles platform differences in Timespec field types.
func FromTimespec(ts syscall.Timespec) time.Time

// FromTimeval converts a syscall.Timeval to time.Time.
// Handles platform differences in Timeval field types.
func FromTimeval(tv syscall.Timeval) time.Time
```

### v0.3.0 - Process & Shell Enhancements

#### process - Enhanced Process Management

**Priority:** HIGH

**Problem:** omniproxy has custom `proc_windows.go` and `proc_unix.go` files with process management logic that overlaps with oscompat/process.

**Affected Projects:**

| Project | File | Pattern |
|---------|------|---------|
| omniproxy | `pkg/daemon/proc_windows.go` | Windows process handling |
| omniproxy | `pkg/daemon/proc_unix.go` | Unix process handling with Setsid |

**Proposed API Additions:**

```go
// IsRunning checks if a process with the given PID is running.
// Unix: Sends signal 0
// Windows: OpenProcess with PROCESS_QUERY_LIMITED_INFORMATION
func IsRunning(pid int) bool

// Kill forcefully terminates a process.
// Unix: SIGKILL
// Windows: TerminateProcess
func Kill(pid int) error

// Terminate gracefully terminates a process.
// Unix: SIGTERM (same as Signal)
// Windows: TerminateProcess (no graceful option)
func Terminate(pid int) error
```

#### shell - Cross-Platform Shell Execution

**Priority:** LOW

**Problem:** Executing shell commands differs between platforms (`sh -c` vs `cmd /c` vs `powershell -Command`).

**Proposed API:**

```go
package shell

// Command creates a command that runs through the system shell.
// Unix: sh -c "command"
// Windows: cmd /c "command"
func Command(command string) *exec.Cmd

// PowerShell creates a PowerShell command (Windows only, no-op on Unix).
func PowerShell(command string) *exec.Cmd

// PathListSeparator returns the separator for PATH-like environment variables.
// Unix: ":"
// Windows: ";"
func PathListSeparator() string

// SplitPathList splits a PATH-like string into components.
func SplitPathList(path string) []string
```

### v0.4.0 - System Configuration (Future)

#### proxy - System Proxy Configuration

**Priority:** LOW

**Problem:** omniproxy has platform-specific system proxy configuration that could be generalized.

**Affected Projects:**

| Project | File | Pattern |
|---------|------|---------|
| omniproxy | `pkg/system/system.go` | `runtime.GOOS` switch for darwin/windows/linux |

**Note:** This is specialized functionality. Consider keeping in omniproxy unless needed by other projects.

## Not Planned for oscompat

The following patterns were identified but are too specialized for oscompat:

| Pattern | Project | Reason |
|---------|---------|--------|
| Security inspection (biometrics, TPM, encryption) | omnitrust | Domain-specific, requires shell/WMI/syscalls |
| Service management (launchd, systemd) | omnistorage-desktop | OS-specific service APIs |
| Video device detection | marp2video | Domain-specific |
| CA certificate installation | omniproxy | Security-sensitive, keep in project |

## Analysis: OS-Specific Code in Grokify Projects

This section documents OS-specific patterns found across grokify repositories.

### Files with Build Constraints

Found 191 files with `//go:build` constraints across grokify projects:

| Project | Count | Primary Patterns |
|---------|-------|------------------|
| opentelemetry-collector-contrib | ~150 | Host metrics, Windows event logs |
| omnitrust | 16 | Security inspection (biometrics, TPM, etc.) |
| omniproxy | 4 | Daemon process, system proxy |
| mogo | 2 | macOS file stats, syscall time |
| opentelemetry-go | 4 | Host ID, OS detection |

### Files with runtime.GOOS Checks

Found 52 files with `runtime.GOOS` checks:

| Pattern | Count | Examples |
|---------|-------|----------|
| `runtime.GOOS == "windows"` | 28 | Path handling, process management |
| `runtime.GOOS == "darwin"` | 15 | macOS-specific features |
| `runtime.GOOS == "linux"` | 9 | Linux-specific features |

### Key Patterns by Category

**Process Management:**
- Signal handling (SIGTERM vs Kill)
- Process detachment (Setsid vs CREATE_NEW_PROCESS_GROUP)
- Process liveness checking (signal 0 vs OpenProcess)

**Path Handling:**
- Case sensitivity (Windows case-insensitive)
- Path separators (/ vs \)
- Config directories (XDG vs Library vs AppData)

**Time Handling:**
- Timestamp precision differences
- Birthtime availability (macOS only reliably)
- Syscall time type differences

**IPC:**
- Unix domain sockets (Unix) vs Named pipes/TCP (Windows)
- Socket file cleanup

## References

### XDG Base Directory Specification

- https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

### Windows Known Folders

- https://docs.microsoft.com/en-us/windows/win32/shell/known-folders

### macOS File System

- https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/

### Go Build Constraints

- https://pkg.go.dev/cmd/go#hdr-Build_constraints
