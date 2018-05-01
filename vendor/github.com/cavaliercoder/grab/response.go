package grab

import (
	"context"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Response represents the response to a completed or in-progress download
// request.
//
// A response may be returned as soon a HTTP response is received from a remote
// server, but before the body content has started transferring.
//
// All Response method calls are thread-safe.
type Response struct {
	// The Request that was submitted to obtain this Response.
	Request *Request

	// HTTPResponse represents the HTTP response received from an HTTP request.
	//
	// The response Body should not be used as it will be consumed and closed by
	// grab.
	HTTPResponse *http.Response

	// Filename specifies the path where the file transfer is stored in local
	// storage.
	Filename string

	// Size specifies the total expected size of the file transfer.
	Size int64

	// Start specifies the time at which the file transfer started.
	Start time.Time

	// End specifies the time at which the file transfer completed.
	//
	// This will return zero until the transfer has completed.
	End time.Time

	// CanResume specifies that the remote server advertised that it can resume
	// previous downloads, as the 'Accept-Ranges: bytes' header is set.
	CanResume bool

	// DidResume specifies that the file transfer resumed a previously incomplete
	// transfer.
	DidResume bool

	// Done is closed once the transfer is finalized, either successfully or with
	// errors. Errors are available via Response.Err
	Done chan struct{}

	// ctx is a Context that controls cancelation of an inprogress transfer
	ctx context.Context

	// cancel is a cancel func that can be used to cancel the context of this
	// Response.
	cancel context.CancelFunc

	// fi is the FileInfo for the destination file if it already existed before
	// transfer started.
	fi os.FileInfo

	// optionsKnown indicates that a HEAD request has been completed and the
	// capabilities of the remote server are known.
	optionsKnown bool

	// writer is the file handle used to write the downloaded file to local
	// storage
	writer     io.WriteCloser
	writeFlags int

	// bytesCompleted specifies the number of bytes which were already
	// transferred before this transfer began.
	bytesResumed int64

	// transfer is responsible for copying data from the remote server to a local
	// file, tracking progress and allowing for cancelation.
	transfer *transfer

	// bytesPerSecond specifies the number of bytes that have been transferred in
	// the last 1-second window.
	bytesPerSecond   float64
	bytesPerSecondMu sync.Mutex

	// bufferSize specifies the size in bytes of the transfer buffer.
	bufferSize int

	// Error contains any error that may have occurred during the file transfer.
	// This should not be read until IsComplete returns true.
	err error
}

// IsComplete returns true if the download has completed. If an error occurred
// during the download, it can be returned via Err.
func (c *Response) IsComplete() bool {
	select {
	case <-c.Done:
		return true
	default:
		return false
	}
}

// Cancel cancels the file transfer by cancelling the underlying Context for
// this Response. Cancel blocks until the transfer is closed and returns any
// error - typically context.Canceled.
func (c *Response) Cancel() error {
	c.cancel()
	return c.Err()
}

// Wait blocks until the download is completed.
func (c *Response) Wait() {
	<-c.Done
}

// Err blocks the calling goroutine until the underlying file transfer is
// completed and returns any error that may have occurred. If the download is
// already completed, Err returns immediately.
func (c *Response) Err() error {
	<-c.Done
	return c.err
}

// BytesComplete returns the total number of bytes which have been copied to
// the destination, including any bytes that were resumed from a previous
// download.
func (c *Response) BytesComplete() int64 {
	return c.bytesResumed + c.transfer.N()
}

// BytesPerSecond returns the number of bytes transferred in the last second. If
// the download is already complete, the average bytes/sec for the life of the
// download is returned.
func (c *Response) BytesPerSecond() float64 {
	if c.IsComplete() {
		return float64(c.transfer.N()) / c.Duration().Seconds()
	}
	c.bytesPerSecondMu.Lock()
	defer c.bytesPerSecondMu.Unlock()
	return c.bytesPerSecond
}

// Progress returns the ratio of total bytes that have been downloaded. Multiply
// the returned value by 100 to return the percentage completed.
func (c *Response) Progress() float64 {
	if c.Size == 0 {
		return 0
	}
	return float64(c.BytesComplete()) / float64(c.Size)
}

// Duration returns the duration of a file transfer. If the transfer is in
// process, the duration will be between now and the start of the transfer. If
// the transfer is complete, the duration will be between the start and end of
// the completed transfer process.
func (c *Response) Duration() time.Duration {
	if c.IsComplete() {
		return c.End.Sub(c.Start)
	}

	return time.Now().Sub(c.Start)
}

// ETA returns the estimated time at which the the download will complete, given
// the current BytesPerSecond. If the transfer has already completed, the actual
// end time will be returned.
func (c *Response) ETA() time.Time {
	if c.IsComplete() {
		return c.End
	}
	bt := c.BytesComplete()
	bps := c.BytesPerSecond()
	if bps == 0 {
		return time.Time{}
	}
	secs := float64(c.Size-bt) / bps
	return time.Now().Add(time.Duration(secs) * time.Second)
}

// watchBps watches the progress of a transfer and maintains statistics.
func (c *Response) watchBps() {
	var prev int64
	then := c.Start

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-c.Done:
			return

		case now := <-t.C:
			d := now.Sub(then)
			then = now

			cur := c.transfer.N()
			bs := cur - prev
			prev = cur

			c.bytesPerSecondMu.Lock()
			c.bytesPerSecond = float64(bs) / d.Seconds()
			c.bytesPerSecondMu.Unlock()
		}
	}
}
