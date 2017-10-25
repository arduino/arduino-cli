package iohelpers

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/fluxio/iohelpers/counting"
)

// errorCloser implements ReadCloser over an underlying io.Reader with a trivial
// Close that just returns err.
type errorCloser struct {
	io.Reader
	err error
}

func (c *errorCloser) Close() error {
	return c.err
}

func TestReleaseReadCloser(t *testing.T) {
	// Close, then Release case
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		rr := &ReleaseReadCloser{c}
		err := rr.Close()
		if err != nil {
			t.Errorf("Unexpected error on Close: %v", err)
		}
		r := rr.Release()
		if c.Count != 1 {
			t.Errorf("Close, then Release should close backing reader; got %d closes",
				c.Count)
		}
		if r != c {
			t.Errorf("Released ReadCloser should be counting.Closer but %v != %v", r, c)
		}
	}

	// Release, then Close case
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		rr := &ReleaseReadCloser{c}
		r := rr.Release()
		if r != c {
			t.Errorf("Released ReadCloser should be counting.Closer but %v != %v", r, c)
		}
		err := rr.Close()
		if err != nil {
			t.Errorf("Unexpected error on Close: %v", err)
		}
		if c.Count != 0 {
			t.Errorf("Release, then Close should be nop; close count == %d", c.Count)
		}
	}

	// Repeated Release returns nil after first
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		rr := &ReleaseReadCloser{c}
		r1 := rr.Release()
		if r1 != c {
			t.Errorf("Released ReadCloser should be counting.Closer but %v != %v", r1, c)
		}
		r2 := rr.Release()
		if r2 != nil {
			t.Errorf("Repeated Release should return nil after first; got %v", r2)
		}
	}

	// Repeated Close calls are passed through directly.
	// This is often a programming error, but we want to behave exactly
	// as the underlying ReadCloser would.
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		rr := &ReleaseReadCloser{c}
		err1 := rr.Close()
		if c.Count != 1 {
			t.Errorf("Expected 1 close, but got %d", c.Count)
		}
		if err1 != nil {
			t.Errorf("Unexpected error on 1st close: %v", err1)
		}
		err2 := rr.Close()
		if c.Count != 2 {
			t.Errorf("Expected 2 closes, but got %d", c.Count)
		}
		if err2 != nil {
			t.Errorf("Unexpected error on 2nd close: %v", err2)
		}
	}

	// Errors are passed through from Close calls.
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &errorCloser{b, fmt.Errorf("expected error")}
		rr := &ReleaseReadCloser{c}
		err := rr.Close()
		if err != c.err {
			t.Errorf("Expected error: %#v but got: %#v", c.err, err)
		}
	}

	// Test that a deferred Close with Release (the first
	// envisioned idiomatic usage) works as expected.
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		rc := func() io.ReadCloser {
			rr := &ReleaseReadCloser{c}
			defer rr.Close()
			return rr.Release()
		}()
		if c.Count != 0 {
			t.Errorf("Deferred Close should be nop, but close count = %d",
				c.Count)
		}
		if rc != c {
			t.Errorf("Released ReadCloser should be counting.Closer but %v != %v", rc, c)
		}
	}

	// Test that a deferred Close without a Release (the second
	// envisioned idiomatic usage) works as expected.
	{
		b := bytes.NewReader([]byte{1, 2, 3, 4})
		c := &counting.Closer{Reader: b}
		func() {
			rr := &ReleaseReadCloser{c}
			defer rr.Close()
		}()
		if c.Count != 1 {
			t.Errorf("Deferred Close should be executed, but close count = %d", c.Count)
		}
	}
}
