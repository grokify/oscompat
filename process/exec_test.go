package process

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestHelperExec is not a test: it is the body of a helper process that
// calls Exec, which replaces the process on Unix.
func TestHelperExec(t *testing.T) {
	if os.Getenv("OSCOMPAT_HELPER") != "exec" {
		t.Skip("helper process only")
	}
	err := Exec(os.Args[0], []string{os.Args[0], "-test.run=TestHelperTarget"})
	t.Fatalf("Exec returned: %v", err) // reaching here means it failed
}

// TestHelperTarget is not a test: it is the program Exec starts.
func TestHelperTarget(t *testing.T) {
	if os.Getenv("OSCOMPAT_HELPER") != "exec" {
		t.Skip("helper process only")
	}
	os.Stdout.WriteString("exec-target-ran\n") //nolint:errcheck // best effort in a helper
	os.Exit(7)
}

func TestExecRunsProgramAndPropagatesStatus(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperExec") //nolint:gosec // re-runs this test binary
	cmd.Env = append(os.Environ(), "OSCOMPAT_HELPER=exec")
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
		t.Fatalf("err = %v, want exit status 7", err)
	}
	if !strings.Contains(string(out), "exec-target-ran") {
		t.Fatalf("output = %q, want the target's output", out)
	}
}

func TestExecRejectsBadArguments(t *testing.T) {
	if err := Exec("", []string{"x"}); err == nil {
		t.Error("empty path: want error")
	}
	if err := Exec("x", nil); err == nil {
		t.Error("empty argv: want error")
	}
}

func TestExecMissingProgram(t *testing.T) {
	if err := Exec("/nonexistent/oscompat-test-binary", []string{"x"}); err == nil {
		t.Fatal("missing program: want error")
	}
}
