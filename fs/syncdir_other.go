//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd

package fs

func syncDir(string) error { return nil }
