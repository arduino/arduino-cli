package iohelpers

import "io"

// ReleaseReadCloser implements releasable ownership of a ReadCloser.  Once
// initialized, this object assumes "ownership" of its backing ReadCloser; the
// backing ReadCloser will be closed if its Close method is called while
// ownership is retained.  When Release is called, the backing ReadCloser gives
// up ownership of the ReadCloser, which is returned, at which point Close
// becomes a nop.
//
// Example usage:
//
//   func FooBar() (io.ReadCloser, error) {
//   	r, err := GetSomeReadCloser()
//   	if err != nil {
//   		return nil, err
//   	}
//   	rr := &iohelpers.ReleaseReadCloser{r}
//   	defer rr.Close()
//
//   	...  // Code including early error returns (without rr.Release())
//
//   	return rr.Release(), nil // Transfer ownership to caller.
//   }
type ReleaseReadCloser struct {
	io.ReadCloser
}

// Close closes the backing io.ReadCloser, if it has not been released yet.
func (r *ReleaseReadCloser) Close() error {
	if r.ReadCloser != nil {
		return r.ReadCloser.Close()
	}
	return nil
}

// Release returns r's backing io.ReadCloser and transfers ownership to the
// caller.  Release should be called at most once; subsequent calls after the
// first will return nil.
func (r *ReleaseReadCloser) Release() io.ReadCloser {
	result := r.ReadCloser
	r.ReadCloser = nil
	return result
}
