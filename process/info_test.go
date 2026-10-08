package process

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

// startHelper starts a copy of the test binary that blocks until killed.
func startHelper(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperBlock") //nolint:gosec // re-runs this test binary
	cmd.Env = append(os.Environ(), "OSCOMPAT_HELPER=block")
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill() //nolint:errcheck // already exited is fine
		_ = cmd.Wait()         //nolint:errcheck // exit status is irrelevant after a kill
	})
	return cmd
}

// TestHelperBlock is not a test: it is the body of the helper process.
func TestHelperBlock(t *testing.T) {
	if os.Getenv("OSCOMPAT_HELPER") != "block" {
		t.Skip("helper process only")
	}
	time.Sleep(time.Minute)
}

func TestExistsCurrentProcess(t *testing.T) {
	ok, err := Exists(os.Getpid())
	if err != nil || !ok {
		t.Fatalf("Exists(self) = %v, %v; want true, nil", ok, err)
	}
}

func TestExistsAfterExit(t *testing.T) {
	cmd := startHelper(t)
	pid := cmd.Process.Pid
	if ok, err := Exists(pid); err != nil || !ok {
		t.Fatalf("Exists(running helper) = %v, %v; want true, nil", ok, err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait() //nolint:errcheck // killed on purpose
	if ok, err := Exists(pid); err != nil || ok {
		t.Fatalf("Exists(exited helper) = %v, %v; want false, nil", ok, err)
	}
}

func TestInvalidPIDs(t *testing.T) {
	for _, pid := range []int{0, -1, -12345} {
		if _, err := Exists(pid); err == nil {
			t.Errorf("Exists(%d): want error", pid)
		}
		if _, err := StartTime(pid); err == nil {
			t.Errorf("StartTime(%d): want error", pid)
		}
	}
}

func TestStartTime(t *testing.T) {
	before := time.Now()
	cmd := startHelper(t)
	got, err := StartTime(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	// ps reports whole seconds, so allow a little slack on both sides.
	if got.Before(before.Add(-3*time.Second)) || got.After(time.Now().Add(3*time.Second)) {
		t.Fatalf("StartTime = %v, want about %v", got, before)
	}
}

func TestStartTimeOfSelfIsInThePast(t *testing.T) {
	got, err := StartTime(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if got.After(time.Now().Add(2 * time.Second)) {
		t.Fatalf("StartTime(self) = %v is in the future", got)
	}
	if time.Since(got) > 24*time.Hour {
		t.Fatalf("StartTime(self) = %v is implausibly old for a test process", got)
	}
}

func TestStartTimeNotFound(t *testing.T) {
	cmd := startHelper(t)
	pid := cmd.Process.Pid
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait() //nolint:errcheck // killed on purpose
	if _, err := StartTime(pid); !errors.Is(err, ErrNotFound) {
		t.Fatalf("StartTime(exited) err = %v, want ErrNotFound", err)
	}
}
