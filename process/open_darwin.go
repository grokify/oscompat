//go:build darwin

package process

// opener returns the command that opens target on macOS.
func opener(target string) (string, []string) {
	return "open", []string{target}
}
