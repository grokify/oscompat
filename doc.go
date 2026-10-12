// Package oscompat is the module that holds operating-system differences so
// that other modules do not have to.
//
// # Policy
//
// oscompat has no dependencies. It imports the Go standard library and itself,
// and nothing else: no module outside the standard library, and no
// golang.org/x/sys. That is what makes it safe for any module to import,
// including modules that keep their own dependency set small on purpose, such
// as ones that handle key material. A test in this module enforces it.
//
// In return, modules are encouraged to import oscompat instead of keeping
// platform-specific files of their own (lock_unix.go, lock_windows.go, a
// runtime.GOOS check). Code that differs by operating system belongs here once,
// tested on every platform, and a caller reads as portable code.
//
// When a module needs a difference oscompat does not cover, add it here rather
// than beside the caller, using only the standard library.
package oscompat
