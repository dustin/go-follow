// Package follow provides a way to follow (tail -f) an io.Reader.
//
// This io.Reader will never return io.EOF until after a call to
// Close(), instead it will block until more bytes are available.
package follow

import (
	"io"
	"time"
)

const (
	minDelay = 100 * time.Millisecond
	maxDelay = time.Second
)

type follower struct {
	r       io.Reader
	stopped bool
	delay   time.Duration
}

// New provides a new follower for the given Reader.
func New(r io.Reader) io.ReadCloser {
	return &follower{r: r, delay: minDelay}
}

// Close stops following the stream.
func (f *follower) Close() error {
	f.stopped = true
	return nil
}

func min(a, b time.Duration) time.Duration {
	if a > b {
		return b
	}
	return a
}

// Read into the buffer.  Block on EOF.
func (f *follower) Read(b []byte) (n int, err error) {
	for !f.stopped {
		n, err = f.r.Read(b)
		// Got data
		if err == io.EOF {
			time.Sleep(f.delay)
			f.delay = min(maxDelay, f.delay*2)
		} else {
			f.delay = minDelay
			return
		}
	}
	return 0, io.EOF
}
