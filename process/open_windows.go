//go:build windows

package process

// opener returns the command that opens target on Windows. The URL protocol
// handler opens a web address in the default browser and a file with the
// program registered for its type, without going through cmd.exe, which would
// interpret characters such as & in the target.
func opener(target string) (string, []string) {
	return "rundll32", []string{"url.dll,FileProtocolHandler", target}
}
