package extract

import (
	"context"
	"errors"
	"io"
)

func copyCancel(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, newCancelableReader(ctx, src))
}

type cancelableReader struct {
	ctx context.Context
	src io.Reader
}

func (r *cancelableReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, errors.New("interrupted")
	default:
		return r.src.Read(p)
	}
}

func newCancelableReader(ctx context.Context, src io.Reader) *cancelableReader {
	return &cancelableReader{
		ctx: ctx,
		src: src,
	}
}
