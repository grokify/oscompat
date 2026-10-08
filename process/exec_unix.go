//go:build !windows

package process

import (
	"os"
	"syscall"
)

func execProgram(path string, argv []string) error {
	return syscall.Exec(path, argv, os.Environ()) //nolint:gosec // the caller chooses the program
}
