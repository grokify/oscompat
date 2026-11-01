//go:build !darwin && !windows

package process

// opener returns the command that opens target on Linux and the BSDs.
func opener(target string) (string, []string) {
	return "xdg-open", []string{target}
}
