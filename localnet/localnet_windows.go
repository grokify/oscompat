//go:build windows

package localnet

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// portFileDir returns the directory for port files.
func portFileDir() string {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return filepath.Join(dir, "oscompat", "localnet")
	}
	// Fallback
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Local", "oscompat", "localnet")
}

// portFilePath returns the path to the port file.
func portFilePath(name string) string {
	return filepath.Join(portFileDir(), name+".port")
}

// socketPath returns the address description for the given name.
// On Windows, this returns the port file path since we use TCP.
func socketPath(name string) string {
	return portFilePath(name)
}

// listen creates a TCP listener on localhost and stores the port in a file.
func listen(name string) (*Listener, error) {
	portFile := portFilePath(name)

	// Ensure directory exists
	dir := filepath.Dir(portFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to create port file directory: %w", err)
	}

	// Remove existing port file if present
	os.Remove(portFile)

	// Listen on localhost with any available port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to listen: %w", err)
	}

	// Extract the port
	addr := l.Addr().(*net.TCPAddr)
	port := addr.Port

	// Write port to file
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0600); err != nil {
		l.Close()
		return nil, fmt.Errorf("oscompat/localnet: failed to write port file: %w", err)
	}

	return &Listener{
		Listener: l,
		name:     name,
		cleanup: func() error {
			err := os.Remove(portFile)
			if os.IsNotExist(err) {
				return nil // Already cleaned up
			}
			return err
		},
	}, nil
}

// dial reads the port file and connects via TCP to localhost.
func dial(name string) (net.Conn, error) {
	portFile := portFilePath(name)

	// Read port from file
	data, err := os.ReadFile(portFile)
	if err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to read port file: %w", err)
	}

	port := strings.TrimSpace(string(data))

	// Connect to localhost on the specified port
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, fmt.Errorf("oscompat/localnet: failed to connect: %w", err)
	}
	return conn, nil
}

// cleanup removes the port file.
func cleanup(name string) error {
	portFile := portFilePath(name)
	err := os.Remove(portFile)
	if os.IsNotExist(err) {
		return nil // Already cleaned up
	}
	return err
}
