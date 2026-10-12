//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package fs

import "os"

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		_ = d.Close() //nolint:errcheck // the sync error is the one to report
		return err
	}
	return d.Close()
}
