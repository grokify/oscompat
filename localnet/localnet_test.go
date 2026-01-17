package localnet_test

import (
	"io"
	"testing"
	"time"

	"github.com/grokify/oscompat/localnet"
)

func TestListenAndDial(t *testing.T) {
	name := "oscompat-test-" + time.Now().Format("20060102150405")

	// Cleanup before test (ignore error - may not exist)
	_ = localnet.Cleanup(name)

	// Create listener
	listener, err := localnet.Listen(name)
	if err != nil {
		t.Fatalf("Listen() error: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Verify name
	if listener.Name() != name {
		t.Errorf("Name() = %q, want %q", listener.Name(), name)
	}

	// Channel to receive server result
	done := make(chan error, 1)
	received := make(chan []byte, 1)

	// Start server goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer func() { _ = conn.Close() }()

		data, err := io.ReadAll(conn)
		if err != nil {
			done <- err
			return
		}
		received <- data
		done <- nil
	}()

	// Connect as client
	conn, err := localnet.Dial(name)
	if err != nil {
		t.Fatalf("Dial() error: %v", err)
	}

	// Send data
	message := []byte("hello from client")
	_, err = conn.Write(message)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	_ = conn.Close()

	// Wait for server
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Server error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server timeout")
	}

	// Verify received data
	select {
	case data := <-received:
		if string(data) != string(message) {
			t.Errorf("Received %q, want %q", data, message)
		}
	default:
		t.Error("No data received")
	}
}

func TestSocketPath(t *testing.T) {
	path := localnet.SocketPath("testapp")
	if path == "" {
		t.Error("SocketPath() returned empty string")
	}
}

func TestSocketPathEmpty(t *testing.T) {
	path := localnet.SocketPath("")
	if path != "" {
		t.Errorf("SocketPath('') = %q, want empty", path)
	}
}

func TestListenEmptyName(t *testing.T) {
	_, err := localnet.Listen("")
	if err != localnet.ErrInvalidName {
		t.Errorf("Listen('') = %v, want ErrInvalidName", err)
	}
}

func TestDialEmptyName(t *testing.T) {
	_, err := localnet.Dial("")
	if err != localnet.ErrInvalidName {
		t.Errorf("Dial('') = %v, want ErrInvalidName", err)
	}
}

func TestCleanupEmptyName(t *testing.T) {
	err := localnet.Cleanup("")
	if err != localnet.ErrInvalidName {
		t.Errorf("Cleanup('') = %v, want ErrInvalidName", err)
	}
}

func TestCleanupNonExistent(t *testing.T) {
	// Cleanup of non-existent socket should not error
	err := localnet.Cleanup("nonexistent-socket-12345")
	if err != nil {
		t.Errorf("Cleanup() of non-existent = %v, want nil", err)
	}
}

func TestListenerClose(t *testing.T) {
	name := "oscompat-close-test-" + time.Now().Format("20060102150405")

	// Cleanup before test (ignore error - may not exist)
	_ = localnet.Cleanup(name)

	listener, err := localnet.Listen(name)
	if err != nil {
		t.Fatalf("Listen() error: %v", err)
	}

	// Close should clean up
	if err := listener.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Dial should fail after close
	_, err = localnet.Dial(name)
	if err == nil {
		t.Error("Dial() after Close() should fail")
	}
}
