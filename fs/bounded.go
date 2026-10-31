package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrNotRegular is returned for a file that is not a regular file, such as a
// directory or a device.
var ErrNotRegular = errors.New("oscompat/fs: not a regular file")

// ErrTooLarge is returned for a file, or a stream, that is larger than its
// limit.
var ErrTooLarge = errors.New("oscompat/fs: larger than the limit")

// OpenRegularInRoot opens name inside root as OpenInRoot does, and returns the
// file only if the handle it holds is a regular file of at most maxBytes. The
// checks are made on the open handle, not on the path, so they describe the
// file that will be read. On any failure the file is closed and the error
// wraps ErrOutsideRoot, os.ErrNotExist, ErrNotRegular or ErrTooLarge, as the
// case may be.
//
// The size is a snapshot, and a file can grow while it is read. Wrap the file
// with CapReader to enforce the limit on what is actually read:
//
//	f, err := fs.OpenRegularInRoot(root, req.Path, max)
//	if err != nil {
//	    return err
//	}
//	defer f.Close()
//	body := fs.CapReader(f, max)
//
// A negative maxBytes is treated as zero.
func OpenRegularInRoot(root, name string, maxBytes int64) (*os.File, error) {
	maxBytes = max(maxBytes, 0)
	f, err := OpenInRoot(root, name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, closeWith(f, fmt.Errorf("oscompat/fs: %q: %w", name, err))
	}
	switch {
	case !info.Mode().IsRegular():
		return nil, closeWith(f, fmt.Errorf("%w: %q", ErrNotRegular, name))
	case info.Size() > maxBytes:
		return nil, closeWith(f, fmt.Errorf("%w: %q is %d bytes; the limit is %d", ErrTooLarge, name, info.Size(), maxBytes))
	}
	return f, nil
}

// closeWith closes f and returns err, joined with the close error if there is
// one, so neither is lost.
func closeWith(f *os.File, err error) error {
	if cerr := f.Close(); cerr != nil {
		return errors.Join(err, cerr)
	}
	return err
}

// CapReader returns a reader that yields the bytes of r up to maxBytes and then
// reports ErrTooLarge if r holds more. It differs from io.LimitReader, which
// stops silently at the limit and so turns an oversized stream into a
// truncated one that looks complete. Bytes within the limit are delivered
// before the error, and the error is repeated on every later read.
//
// A negative maxBytes is treated as zero.
func CapReader(r io.Reader, maxBytes int64) io.Reader {
	return &capReader{r: r, max: max(maxBytes, 0), left: max(maxBytes, 0)}
}

type capReader struct {
	r    io.Reader
	max  int64
	left int64
	err  error
}

func (c *capReader) Read(p []byte) (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	// Ask for one byte more than is left, so a stream that is exactly at the
	// limit ends with io.EOF and one that is over it is noticed.
	if int64(len(p)) > c.left+1 {
		p = p[:c.left+1]
	}
	n, err := c.r.Read(p)
	if int64(n) > c.left {
		n = int(c.left)
		c.left = 0
		c.err = fmt.Errorf("%w: more than %d bytes", ErrTooLarge, c.max)
		return n, c.err
	}
	c.left -= int64(n)
	return n, err
}
