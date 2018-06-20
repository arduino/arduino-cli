package grab

import (
	"context"
	"io"
	"sync/atomic"
)

type transfer struct {
	n   int64 // must be 64bit aligned on 386
	ctx context.Context
	lim RateLimiter
	w   io.Writer
	r   io.Reader
	b   []byte
}

func newTransfer(ctx context.Context, lim RateLimiter, dst io.Writer, src io.Reader, buf []byte) *transfer {
	return &transfer{
		ctx: ctx,
		lim: lim,
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
		if c.lim != nil {
			err = c.lim.WaitN(c.ctx, len(c.b))
			if err != nil {
				return
			}
		}
		nr, er := c.r.Read(c.b)
		if nr > 0 {
			nw, ew := c.w.Write(c.b[0:nr])
			if nw > 0 {
				written += int64(nw)
				atomic.StoreInt64(&c.n, written)
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
func (c *transfer) N() (n int64) {
	if c == nil {
		return 0
	}
	n = atomic.LoadInt64(&c.n)
	return
}
