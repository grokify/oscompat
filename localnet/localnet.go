// Package localnet provides cross-platform local network communication.
//
// This package abstracts platform differences in local IPC mechanisms:
//   - Unix/Linux/macOS: Unix domain sockets
//   - Windows: TCP on localhost (Unix domain sockets limited on older Windows)
//
// Use this package when you need reliable local inter-process communication
// that works across all platforms.
package localnet

import (
	"errors"
	"net"
)

// Common errors.
var (
	// ErrInvalidName is returned when an empty name is provided.
	ErrInvalidName = errors.New("oscompat/localnet: name cannot be empty")

	// ErrSocketExists is returned when trying to create a listener
	// but a socket file already exists (Unix only).
	ErrSocketExists = errors.New("oscompat/localnet: socket already exists")
)

// Listener wraps a net.Listener with cleanup functionality.
type Listener struct {
	net.Listener
	name    string
	cleanup func() error
}

// Close closes the listener and performs any necessary cleanup.
func (l *Listener) Close() error {
	err := l.Listener.Close()
	if l.cleanup != nil {
		cleanupErr := l.cleanup()
		if err == nil {
			err = cleanupErr
		}
	}
	return err
}

// Name returns the name used to create this listener.
func (l *Listener) Name() string {
	return l.name
}

// Listen creates a local listener for IPC.
//
// On Unix systems, this creates a Unix domain socket in a platform-appropriate
// location (e.g., /tmp/<name>.sock or $XDG_RUNTIME_DIR/<name>.sock).
//
// On Windows, this creates a TCP listener on localhost with an ephemeral port,
// storing the port in a file for clients to discover.
//
// The returned Listener's Close method will clean up any socket files.
func Listen(name string) (*Listener, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	return listen(name)
}

// Dial connects to a local IPC endpoint.
//
// On Unix systems, this connects to the Unix domain socket for the given name.
// On Windows, this reads the port file and connects via TCP to localhost.
func Dial(name string) (net.Conn, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	return dial(name)
}

// SocketPath returns the path or address that would be used for the given name.
// This is useful for debugging or documentation purposes.
func SocketPath(name string) string {
	if name == "" {
		return ""
	}
	return socketPath(name)
}

// Cleanup removes any leftover socket files or port files for the given name.
// This is useful when a previous process crashed without cleaning up.
// It's safe to call even if no socket exists.
func Cleanup(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	return cleanup(name)
}
