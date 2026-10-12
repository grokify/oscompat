package fs

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func lockPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "x.lock")
}

func skipIfUnsupported(t *testing.T) {
	t.Helper()
	unlock, err := Lock(context.Background(), lockPath(t))
	if errors.Is(err, ErrLockUnsupported) {
		t.Skip("file locking is not supported on this platform")
	}
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}

func TestLockExcludesAndReleases(t *testing.T) {
	skipIfUnsupported(t)
	path := lockPath(t)
	ctx := context.Background()

	unlock, err := Lock(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	short, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	if _, err := Lock(short, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("a second Lock while held = %v, want a deadline error", err)
	}

	unlock()
	unlock() // safe to repeat

	again, err := Lock(ctx, path)
	if err != nil {
		t.Fatalf("Lock after release: %v", err)
	}
	again()
}

func TestLockWaitsForTheHolder(t *testing.T) {
	skipIfUnsupported(t)
	path := lockPath(t)
	ctx := context.Background()
	unlock, err := Lock(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	got := make(chan error, 1)
	go func() {
		u, err := Lock(ctx, path)
		if err == nil {
			u()
		}
		got <- err
	}()

	select {
	case err := <-got:
		t.Fatalf("Lock returned while held: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	unlock()
	select {
	case err := <-got:
		if err != nil {
			t.Fatalf("Lock after release: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Lock did not acquire after the holder released")
	}
}

func TestLockCancelledContext(t *testing.T) {
	skipIfUnsupported(t)
	path := lockPath(t)
	unlock, err := Lock(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Lock(ctx, path); !errors.Is(err, context.Canceled) {
		t.Fatalf("Lock with a cancelled context = %v, want context.Canceled", err)
	}
}

func TestLockCreatesAPrivateFile(t *testing.T) {
	skipIfUnsupported(t)
	path := lockPath(t)
	unlock, err := Lock(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPrivate(info); err != nil {
		t.Fatal(err)
	}
}

func TestLockInAMissingDirectoryFails(t *testing.T) {
	_, err := Lock(context.Background(), filepath.Join(t.TempDir(), "no", "such", "x.lock"))
	if err == nil || errors.Is(err, ErrLockUnsupported) {
		t.Fatalf("Lock in a missing directory = %v", err)
	}
}

// TestLockAcrossProcesses starts this test binary as a second process that
// tries to take the lock the parent holds.
func TestLockAcrossProcesses(t *testing.T) {
	skipIfUnsupported(t)
	path := lockPath(t)
	unlock, err := Lock(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	//nolint:gosec // G204: re-executes this test binary with a fixed argument, as the standard library's own tests do
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockHelperProcess$")
	cmd.Env = append(os.Environ(), "OSCOMPAT_LOCK_HELPER="+path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill() //nolint:errcheck // the helper may already have exited
		_ = cmd.Wait()         //nolint:errcheck // reaping the helper; its status is not under test
	})

	lines := bufio.NewScanner(out)
	read := func() string {
		t.Helper()
		done := make(chan string, 1)
		go func() {
			if lines.Scan() {
				done <- lines.Text()
			} else {
				done <- "EOF"
			}
		}()
		select {
		case s := <-done:
			return s
		case <-time.After(30 * time.Second):
			t.Fatal("the helper process did not answer")
			return ""
		}
	}

	// Filter the test binary's own output ("PASS" and so on): only our markers
	// count.
	next := func() string {
		for {
			s := read()
			if strings.HasPrefix(s, "helper:") || s == "EOF" {
				return s
			}
		}
	}
	if got := next(); got != "helper:blocked" {
		t.Fatalf("while the parent held the lock, the helper said %q", got)
	}
	unlock()
	if got := next(); got != "helper:acquired" {
		t.Fatalf("after the parent released the lock, the helper said %q", got)
	}
}

func TestLockHelperProcess(t *testing.T) {
	path := os.Getenv("OSCOMPAT_LOCK_HELPER")
	if path == "" {
		t.Skip("helper process for TestLockAcrossProcesses")
	}
	short, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	_, err := Lock(short, path)
	cancel()
	if errors.Is(err, context.DeadlineExceeded) {
		os.Stdout.WriteString("helper:blocked\n") //nolint:errcheck // the parent reads this line
	} else {
		os.Stdout.WriteString("helper: unexpected " + errString(err) + "\n") //nolint:errcheck // the parent reads this line
		return
	}
	unlock, err := Lock(context.Background(), path)
	if err != nil {
		os.Stdout.WriteString("helper: failed " + err.Error() + "\n") //nolint:errcheck // the parent reads this line
		return
	}
	unlock()
	os.Stdout.WriteString("helper:acquired\n") //nolint:errcheck // the parent reads this line
}

func errString(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}

func TestSyncDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("x"), PrivateFilePerm); err != nil {
		t.Fatal(err)
	}
	if err := SyncDir(dir); err != nil {
		t.Fatalf("SyncDir: %v", err)
	}
	if runtime.GOOS == "windows" {
		// A no-op there, even for a directory that does not exist.
		if err := SyncDir(filepath.Join(dir, "missing")); err != nil {
			t.Fatalf("SyncDir on Windows = %v, want nil", err)
		}
		return
	}
	if err := SyncDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("SyncDir of a missing directory succeeded")
	}
}

func TestPermissionHelpers(t *testing.T) {
	if got, want := PermissionsEnforced(), runtime.GOOS != "windows"; got != want {
		t.Fatalf("PermissionsEnforced = %v, want %v on %s", got, want, runtime.GOOS)
	}
	dir := t.TempDir()
	private := filepath.Join(dir, "private")
	open := filepath.Join(dir, "open")
	for path, mode := range map[string]os.FileMode{private: 0o600, open: 0o644} {
		if err := os.WriteFile(path, []byte("x"), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	pi, err := os.Stat(private)
	if err != nil {
		t.Fatal(err)
	}
	oi, err := os.Stat(open)
	if err != nil {
		t.Fatal(err)
	}

	if !IsPrivate(pi) || CheckPrivate(pi) != nil {
		t.Error("an owner-only file was not private")
	}
	if !PermissionsEnforced() {
		if !IsPrivate(oi) || CheckPrivate(oi) != nil {
			t.Error("where modes are not enforced, nothing is reported as not private")
		}
		return
	}
	if IsPrivate(oi) {
		t.Error("a 0644 file was reported private")
	}
	err = CheckPrivate(oi)
	if !errors.Is(err, ErrNotPrivate) || !strings.Contains(err.Error(), "0644") {
		t.Errorf("CheckPrivate(0644) = %v", err)
	}
}

func TestChmodPrivate(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "x")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck // a temporary file in a test
	if err := ChmodPrivate(f); err != nil {
		t.Fatal(err)
	}
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPrivate(info); err != nil {
		t.Fatal(err)
	}
}
