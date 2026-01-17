# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-17

### Added

- **id**: Cross-platform unique identifier generation using `crypto/rand`
  - `Generate(byteLen)` for custom-length hex IDs
  - `Generate16()` for 16-character hex IDs (8 random bytes)
  - `Generate32()` for 32-character hex IDs (16 random bytes)

- **process**: Cross-platform process management utilities
  - `SetDetached(cmd)` to configure commands to run detached from parent
  - `Signal(pid)` to send termination signal (SIGTERM on Unix, Kill on Windows)
  - `FindAndSignal(pid)` to find and signal a process

- **paths**: Cross-platform configuration and data directory resolution
  - `Home()` for user's home directory
  - `UserConfig()`, `UserData()`, `UserCache()`, `UserRuntime()` for user directories
  - `SystemConfig()` for system-wide config directory
  - `AppConfig(name)`, `AppData(name)`, `AppCache(name)`, `AppRuntime(name)` for app-specific directories
  - `SystemAppConfig(name)` for system-wide app configuration
  - Supports XDG Base Directory Specification on Linux, ~/Library on macOS, %APPDATA% on Windows

- **fs**: Cross-platform filesystem utilities
  - Permission constants: `DefaultDirPerm`, `DefaultFilePerm`, `PrivateDirPerm`, `PrivateFilePerm`, `ExecutablePerm`
  - `ValidatePath()`, `ValidatePathStrict()` for path traversal detection
  - `NormalizePath()`, `OSPath()` for path separator handling
  - `JoinNormalized()`, `JoinOS()`, `SafeJoin()` for path joining
  - `MkdirAll()`, `MkdirAllPrivate()` for directory creation
  - `WriteFile()`, `WriteFilePrivate()` for file writing
  - `IsCaseSensitive()`, `PathEqual()`, `PathHasPrefix()` for case-aware path comparison
  - `Birthtime()`, `BirthtimeInfo()`, `BirthtimeSupported()` for file creation time

- **tsync**: Cross-platform timestamp utilities for file synchronization
  - Tolerance constants: `DefaultTolerance`, `FAT32Tolerance`, `HighPrecisionTolerance`
  - `Tolerance()` for recommended platform tolerance
  - `Equal()`, `EqualWithTolerance()` for tolerant equality comparison
  - `Before()`, `BeforeWithTolerance()`, `After()`, `AfterWithTolerance()` for tolerant ordering
  - `Compare()`, `CompareWithTolerance()` for three-way comparison
  - `Newer()`, `Older()` for selecting timestamps
  - `Truncate()`, `TruncateToSecond()` for precision normalization
  - `FromTimespec()`, `FromTimeval()` for syscall time conversion

- **localnet**: Cross-platform local IPC (Inter-Process Communication)
  - `Listen(name)` to create local listener (Unix socket or TCP on localhost)
  - `Dial(name)` to connect to local server
  - `SocketPath(name)` to get path/address for documentation
  - `Cleanup(name)` to remove stale socket/port files

[0.1.0]: https://github.com/grokify/oscompat/releases/tag/v0.1.0
