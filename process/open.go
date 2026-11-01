package process

import (
	"fmt"
	"os/exec"
	"strings"
)

// OpenDefault opens target, a URL or a file path, with the user's default
// application: the browser for a web address, the registered program for a
// file. It returns once the opener has started and does not wait for the
// application.
//
// It runs open on macOS, xdg-open on Linux and the BSDs, and the URL protocol
// handler of rundll32 on Windows. A target that is empty, or that starts with
// "-" and so could be read as an option of the opener, is refused.
func OpenDefault(target string) error {
	name, args, err := openCommand(target)
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...) // #nosec G204 -- the program is fixed per platform and target is checked above
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("process: open %q: %w", target, err)
	}
	// Reap the opener so it does not linger as a zombie. Its exit status says
	// nothing useful: an opener can hand the target off and exit before the
	// application starts, or report a failure in a window of its own.
	go func() { _ = cmd.Wait() }()
	return nil
}

// openCommand returns the program and arguments that open target on this
// platform, after checking target.
func openCommand(target string) (string, []string, error) {
	if strings.TrimSpace(target) == "" {
		return "", nil, fmt.Errorf("process: open: empty target")
	}
	if strings.HasPrefix(target, "-") {
		return "", nil, fmt.Errorf("process: open %q: a target starting with - would be read as an option", target)
	}
	name, args := opener(target)
	return name, args, nil
}
