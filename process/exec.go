package process

import "fmt"

// Exec runs the program at path with the given arguments, where argv[0] is
// the program name as it should see it, using the current environment and
// working directory.
//
// On Unix it replaces the current process, so it returns only on failure. On
// Windows, which cannot replace a running process, it runs the program as a
// child with the terminal attached and exits with the child's exit status; it
// also returns only on failure. Callers should treat a nil return as
// impossible, and look the program up (exec.LookPath) first for a clear error.
func Exec(path string, argv []string) error {
	if path == "" {
		return fmt.Errorf("process: exec: empty path")
	}
	if len(argv) == 0 {
		return fmt.Errorf("process: exec: empty argv")
	}
	return execProgram(path, argv)
}
