//go:build windows

package testutil

// posix: Windows runs no #! script as a program and keeps no execute bits.
const posix = false
