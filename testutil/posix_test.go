package testutil

import (
	"runtime"
	"testing"
)

func TestPOSIX(t *testing.T) {
	if POSIX() == (runtime.GOOS == "windows") {
		t.Errorf("POSIX() = %v on %s", POSIX(), runtime.GOOS)
	}
}

func TestRequirePOSIX(t *testing.T) {
	skipped := true
	t.Run("inner", func(t *testing.T) {
		RequirePOSIX(t, "runs shell scripts as fake programs")
		skipped = false
	})
	if skipped == POSIX() {
		t.Errorf("RequirePOSIX skipped %v on a platform where POSIX() is %v", skipped, POSIX())
	}
}
