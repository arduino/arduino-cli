package grab

import (
	"context"
	"io"
	"sync"
)

type transfer struct {
	mu  sync.Mutex // guards n
	ctx context.Context
	w   io.Writer
	r   io.Reader
	b   []byte
	n   int64
}

func newTransfer(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) *transfer {
	return &transfer{
		ctx: ctx,
		w:   dst,
		r:   src,
		b:   buf,
	}
}

// copy behaves similarly to io.CopyBuffer except that it checks for cancelation
// of the given context.Context and reports progress in a thread-safe manner.
func (c *transfer) copy() (written int64, err error) {
	if c.b == nil {
		c.b = make([]byte, 32*1024)
	}
	for {
		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		default:
			// keep working
		}
		nr, er := c.r.Read(c.b)
		if nr > 0 {
			nw, ew := c.w.Write(c.b[0:nr])
			if nw > 0 {
				written += int64(nw)
				c.mu.Lock()
				c.n = written
				c.mu.Unlock()
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// N returns the number of bytes transferred.
func (c *transfer) N() int64 {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
