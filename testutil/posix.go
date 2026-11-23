package testutil

import "testing"

// POSIX reports whether this platform runs programs and keeps file modes
// as POSIX systems do: a script with a #! line runs as a program, execute
// bits exist, and a path rooted at / is absolute. It is false on Windows.
func POSIX() bool { return posix }

// RequirePOSIX skips the test where POSIX does not hold, saying why the
// test needs it, such as "runs shell scripts as fake programs". Use it for
// tests that fake a Linux host or run shell scripts as programs; test
// cross-platform behavior without it.
func RequirePOSIX(t testing.TB, why string) {
	t.Helper()
	if !posix {
		t.Skipf("needs a POSIX platform: %s", why)
	}
}
