package process

import (
	"runtime"
	"slices"
	"testing"
)

// The opener for each platform, so the test checks the exact command without
// starting a browser.
var wantOpener = map[string]struct {
	name string
	pre  []string
}{
	"darwin":  {"open", nil},
	"windows": {"rundll32", []string{"url.dll,FileProtocolHandler"}},
}

func TestOpenCommandIsThePlatformsOpener(t *testing.T) {
	want, ok := wantOpener[runtime.GOOS]
	if !ok {
		want.name = "xdg-open"
	}
	for _, target := range []string{
		"https://example.com/a?b=1&c=2",
		"/tmp/report with spaces.html",
		`C:\Users\me\report.html`,
	} {
		name, args, err := openCommand(target)
		if err != nil {
			t.Fatalf("%q: %v", target, err)
		}
		if name != want.name {
			t.Errorf("%q: program = %q, want %q", target, name, want.name)
		}
		// The target is passed as one argument, unchanged and last.
		if wantArgs := append(slices.Clone(want.pre), target); !slices.Equal(args, wantArgs) {
			t.Errorf("%q: args = %q, want %q", target, args, wantArgs)
		}
	}
}

func TestOpenDefaultRefusesWhatItCannotOpenSafely(t *testing.T) {
	for _, target := range []string{"", "   ", "-a", "--help"} {
		if err := OpenDefault(target); err == nil {
			t.Errorf("OpenDefault(%q) was accepted", target)
		}
	}
}
