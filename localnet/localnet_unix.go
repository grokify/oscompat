//go:build !windows

package localnet

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// socketDir returns the directory for socket files.
func socketDir() string {
	// Prefer XDG_RUNTIME_DIR if available (more secure, auto-cleaned)
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	// Fallback to /tmp
	return "/tmp"
}

// socketPath returns the full path to the socket file.
func socketPath(name string) string {
	return filepath.Join(socketDir(), name+".sock")
}

// listen creates a Unix domain socket listener.
func listen(name string) (*Listener, error) {
	path := socketPath(name)

	// Remove existing socket if present
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("oscompat/localnet: failed to remove existing socket: %w", err)
	}

	// Create the listener
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to listen: %w", err)
	}

	// Set permissions to owner-only for security
	if err := os.Chmod(path, 0700); err != nil {
		_ = l.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("oscompat/localnet: failed to set socket permissions: %w", err)
	}

	return &Listener{
		Listener: l,
		name:     name,
		cleanup: func() error {
			err := os.Remove(path)
			if os.IsNotExist(err) {
				return nil // Already cleaned up
			}
			return err
		},
	}, nil
}

// dial connects to a Unix domain socket.
func dial(name string) (net.Conn, error) {
	path := socketPath(name)
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to connect: %w", err)
	}
	return conn, nil
}

// cleanup removes the socket file.
func cleanup(name string) error {
	path := socketPath(name)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // Already cleaned up
	}
	return err
}
