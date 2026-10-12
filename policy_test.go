package oscompat

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/grokify/oscompat"

// TestHasNoDependencies enforces the module's policy: it imports the standard
// library and itself, and requires no other module.
func TestHasNoDependencies(t *testing.T) {
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(mod), "\n") {
		f := strings.Fields(strings.SplitN(line, "//", 2)[0])
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "require", "replace", "tool":
			t.Errorf("go.mod:%d: %q: oscompat takes no dependencies", i+1, strings.TrimSpace(line))
		}
	}
	if _, err := os.Stat("go.sum"); err == nil {
		t.Error("go.sum exists: oscompat takes no dependencies")
	}

	fset := token.NewFileSet()
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != "." && (strings.HasPrefix(name, ".") || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if p == modulePath || strings.HasPrefix(p, modulePath+"/") {
				continue
			}
			// A standard library path has no dot in its first element.
			if first, _, _ := strings.Cut(p, "/"); strings.Contains(first, ".") {
				t.Errorf("%s imports %s: oscompat imports only the standard library and itself", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
