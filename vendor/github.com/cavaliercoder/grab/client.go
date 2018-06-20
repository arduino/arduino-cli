package grab

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// A Client is a file download client.
//
// Clients are safe for concurrent use by multiple goroutines.
type Client struct {
	// HTTPClient specifies the http.Client which will be used for communicating
	// with the remote server during the file transfer.
	HTTPClient *http.Client

	// UserAgent specifies the User-Agent string which will be set in the
	// headers of all requests made by this client.
	//
	// The user agent string may be overridden in the headers of each request.
	UserAgent string

	// BufferSize specifies the size in bytes of the buffer that is used for
	// transferring all requested files. Larger buffers may result in faster
	// throughput but will use more memory and result in less frequent updates
	// to the transfer progress statistics. The BufferSize of each request can
	// be overridden on each Request object. Default: 32KB.
	BufferSize int
}

// NewClient returns a new file download Client, using default configuration.
func NewClient() *Client {
	return &Client{
		UserAgent: "grab",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

// DefaultClient is the default client and is used by all Get convenience
// functions.
var DefaultClient = NewClient()

// Do sends a file transfer request and returns a file transfer response,
// following policy (e.g. redirects, cookies, auth) as configured on the
// client's HTTPClient.
//
// Like http.Get, Do blocks while the transfer is initiated, but returns as soon
// as the transfer has started transferring in a background goroutine, or if it
// failed early.
//
// An error is returned via Response.Err if caused by client policy (such as
// CheckRedirect), or if there was an HTTP protocol or IO error. Response.Err
// will block the caller until the transfer is completed, successfully or
// otherwise.
func (c *Client) Do(req *Request) *Response {
	// cancel will be called on all code-paths via closeResponse
	ctx, cancel := context.WithCancel(req.Context())
	resp := &Response{
		Request:    req,
		Start:      time.Now(),
		Done:       make(chan struct{}, 0),
		Filename:   req.Filename,
		ctx:        ctx,
		cancel:     cancel,
		bufferSize: req.BufferSize,
	}
	if resp.bufferSize == 0 {
		// default to Client.BufferSize
		resp.bufferSize = c.BufferSize
	}

	// Run state-machine while caller is blocked to initialize the file transfer.
	// Must never transition to the copyFile state - this happens next in another
	// goroutine.
	c.run(resp, c.statFileInfo)

	// Run copyFile in a new goroutine. copyFile will no-op if the transfer is
	// already complete or failed.
	go c.run(resp, c.copyFile)
	return resp
}

// DoChannel executes all requests sent through the given Request channel, one
// at a time, until it is closed by another goroutine. The caller is blocked
// until the Request channel is closed and all transfers have completed. All
// responses are sent through the given Response channel as soon as they are
// received from the remote servers and can be used to track the progress of
// each download.
//
// Slow Response receivers will cause a worker to block and therefore delay the
// start of the transfer for an already initiated connection - potentially
// causing a server timeout. It is the caller's responsibility to ensure a
// sufficient buffer size is used for the Response channel to prevent this.
//
// If an error occurs during any of the file transfers it will be accessible via
// the associated Response.Err function.
func (c *Client) DoChannel(reqch <-chan *Request, respch chan<- *Response) {
	// TODO: enable cancelling of batch jobs
	for req := range reqch {
		resp := c.Do(req)
		respch <- resp
		<-resp.Done
	}
}

// DoBatch executes all the given requests using the given number of concurrent
// workers. Control is passed back to the caller as soon as the workers are
// initiated.
//
// If the requested number of workers is less than one, a worker will be created
// for every request. I.e. all requests will be executed concurrently.
//
// If an error occurs during any of the file transfers it will be accessible via
// call to the associated Response.Err.
//
// The returned Response channel is closed only after all of the given Requests
// have completed, successfully or otherwise.
func (c *Client) DoBatch(workers int, requests ...*Request) <-chan *Response {
	if workers < 1 {
		workers = len(requests)
	}
	reqch := make(chan *Request, len(requests))
	respch := make(chan *Response, len(requests))
	wg := sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			c.DoChannel(reqch, respch)
			wg.Done()
		}()
	}

	// queue requests
	go func() {
		for _, req := range requests {
			reqch <- req
		}
		close(reqch)
		wg.Wait()
		close(respch)
	}()
	return respch
}

// An stateFunc is an action that mutates the state of a Response and returns
// the next stateFunc to be called.
type stateFunc func(*Response) stateFunc

// run calls the given stateFunc function and all subsequent returned stateFuncs
// until a stateFunc returns nil or the Response.ctx is canceled. Each stateFunc
// should mutate the state of the given Response until it has completed
// downloading or failed.
func (c *Client) run(resp *Response, f stateFunc) {
	for {
		select {
		case <-resp.ctx.Done():
			if resp.IsComplete() {
				return
			}
			resp.err = resp.ctx.Err()
			f = c.closeResponse

		default:
			// keep working
		}
		if f = f(resp); f == nil {
			return
		}
	}
}

// statFileInfo retrieves FileInfo for any local file matching
// Response.Filename.
//
// If the file does not exist, is a directory, or its name is unknown the next
// stateFunc is headRequest.
//
// If the file exists, Response.fi is set and the next stateFunc is
// validateLocal.
//
// If an error occurs, the next stateFunc is closeResponse.
func (c *Client) statFileInfo(resp *Response) stateFunc {
	if resp.Filename == "" {
		return c.headRequest
	}
	fi, err := os.Stat(resp.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return c.headRequest
		}
		resp.err = err
		return c.closeResponse
	}
	if fi.IsDir() {
		resp.Filename = ""
		return c.headRequest
	}
	resp.fi = fi
	return c.validateLocal
}

// validateLocal compares a local copy of the downloaded file to the remote
// file.
//
// An error is returned if the local file is larger than the remote file, or
// Request.SkipExisting is true.
//
// If the existing file matches the length of the remote file, the next
// stateFunc is checksumFile.
//
// If the local file is smaller than the remote file and the remote server is
// known to support ranged requests, the next stateFunc is getRequest.
func (c *Client) validateLocal(resp *Response) stateFunc {
	if resp.Request.SkipExisting {
		resp.err = ErrFileExists
		return c.closeResponse
	}

	// determine expected file size
	size := resp.Request.Size
	if size == 0 && resp.HTTPResponse != nil {
		size = resp.HTTPResponse.ContentLength
	}
	if size == 0 {
		return c.headRequest
	}

	if size == resp.fi.Size() {
		resp.DidResume = true
		resp.bytesResumed = resp.fi.Size()
		return c.checksumFile
	}

	if resp.Request.NoResume {
		return c.getRequest
	}

	if size < resp.fi.Size() {
		resp.err = ErrBadLength
		return c.closeResponse
	}

	if resp.CanResume {
		resp.Request.HTTPRequest.Header.Set(
			"Range",
			fmt.Sprintf("bytes=%d-", resp.fi.Size()))
		resp.DidResume = true
		resp.bytesResumed = resp.fi.Size()
		return c.getRequest
	}
	return c.headRequest
}

func (c *Client) checksumFile(resp *Response) stateFunc {
	if resp.Request.hash == nil {
		return c.closeResponse
	}

	if resp.Filename == "" {
		panic("filename not set")
	}

	// open downloaded file
	f, err := os.Open(resp.Filename)
	if err != nil {
		resp.err = err
		return c.closeResponse
	}
	defer f.Close()

	// hash file
	t := newTransfer(resp.Request.Context(), nil, resp.Request.hash, f, nil)
	if nc, err := t.copy(); err != nil {
		resp.err = err
		return c.closeResponse
	} else if nc != resp.Size {
		resp.err = ErrBadLength
		return c.closeResponse
	}

	// compare checksum
	sum := resp.Request.hash.Sum(nil)
	if !bytes.Equal(sum, resp.Request.checksum) {
		if resp.Request.deleteOnError {
			os.Remove(resp.Filename)
		}
		resp.err = ErrBadChecksum
	}
	return c.closeResponse
}

// doHTTPRequest sends a HTTP Request and returns the response
func (c *Client) doHTTPRequest(req *http.Request) (*http.Response, error) {
	if c.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return c.HTTPClient.Do(req)
}

func (c *Client) headRequest(resp *Response) stateFunc {
	if resp.optionsKnown {
		return c.getRequest
	}
	resp.optionsKnown = true

	if resp.Request.NoResume {
		return c.getRequest
	}

	if resp.Filename != "" && resp.fi == nil {
		// destination path is already known and does not exist
		return c.getRequest
	}

	hreq := new(http.Request)
	*hreq = *resp.Request.HTTPRequest
	hreq.Method = "HEAD"

	resp.HTTPResponse, resp.err = c.doHTTPRequest(hreq)
	if resp.err != nil {
		return c.closeResponse
	}
	return c.readResponse
}

func (c *Client) getRequest(resp *Response) stateFunc {
	resp.HTTPResponse, resp.err = c.doHTTPRequest(resp.Request.HTTPRequest)
	if resp.err != nil {
		return c.closeResponse
	}
	return c.readResponse
}

func (c *Client) readResponse(resp *Response) stateFunc {
	if resp.HTTPResponse == nil {
		panic("Response.HTTPResponse is not ready")
	}

	// check status code
	if !resp.Request.IgnoreBadStatusCodes {
		if resp.HTTPResponse.StatusCode < 200 || resp.HTTPResponse.StatusCode > 299 {
			resp.err = ErrBadStatusCode
			return c.closeResponse
		}
	}

	// check expected size
	resp.Size = resp.bytesResumed + resp.HTTPResponse.ContentLength
	if resp.HTTPResponse.ContentLength > 0 && resp.Request.Size > 0 {
		if resp.Request.Size != resp.Size {
			resp.err = ErrBadLength
			return c.closeResponse
		}
	}

	// check filename
	if resp.Filename == "" {
		filename, err := guessFilename(resp.HTTPResponse)
		if err != nil {
			resp.err = err
			return c.closeResponse
		}
		// Request.Filename will be empty or a directory
		resp.Filename = filepath.Join(resp.Request.Filename, filename)
		return c.statFileInfo
	}

	if resp.HTTPResponse.Header.Get("Accept-Ranges") == "bytes" {
		resp.CanResume = true
	}

	if resp.HTTPResponse.Request.Method == "HEAD" {
		return c.statFileInfo
	}
	return c.openWriter
}

// openWriter opens the destination file for writing and seeks to the location
// from whence the file transfer will resume.
//
// Requires that Response.Filename and resp.DidResume are already be set.
func (c *Client) openWriter(resp *Response) stateFunc {
	if !resp.Request.NoCreateDirectories {
		resp.err = mkdirp(resp.Filename)
		if resp.err != nil {
			return c.closeResponse
		}
	}

	// compute write flags
	flag := os.O_CREATE | os.O_WRONLY
	if resp.fi != nil {
		if resp.DidResume {
			flag = os.O_APPEND | os.O_WRONLY
		} else {
			flag = os.O_TRUNC | os.O_WRONLY
		}
	}

	// open file
	f, err := os.OpenFile(resp.Filename, flag, 0644)
	if err != nil {
		resp.err = err
		return c.closeResponse
	}
	resp.writer = f

	// seek to start or end
	whence := os.SEEK_SET
	if resp.bytesResumed > 0 {
		whence = os.SEEK_END
	}
	_, resp.err = f.Seek(0, whence)
	if resp.err != nil {
		return c.closeResponse
	}

	// init transfer
	if resp.bufferSize < 1 {
		resp.bufferSize = 32 * 1024
	}
	b := make([]byte, resp.bufferSize)
	resp.transfer = newTransfer(
		resp.Request.Context(),
		resp.Request.RateLimiter,
		resp.writer,
		resp.HTTPResponse.Body,
		b)

	// next step is copyFile, but this will be called later in another goroutine
	return nil
}

// copy transfers content for a HTTP connection established via Client.do()
func (c *Client) copyFile(resp *Response) stateFunc {
	if resp.IsComplete() {
		return nil
	}

	// run BeforeCopy hook
	if resp.Request.BeforeCopy != nil {
		resp.err = resp.Request.BeforeCopy(resp)
		if resp.err != nil {
			return c.closeResponse
		}
	}

	if resp.transfer == nil {
		panic("developer error: Response.transfer is not initialized")
	}
	go resp.watchBps()
	if _, resp.err = resp.transfer.copy(); resp.err != nil {
		return c.closeResponse
	}

	// set timestamp
	if !resp.Request.IgnoreRemoteTime {
		if resp.err = setLastModified(resp.HTTPResponse, resp.Filename); resp.err != nil {
			return c.closeResponse
		}
	}
	return c.checksumFile
}

// close finalizes the Response
func (c *Client) closeResponse(resp *Response) stateFunc {
	if resp.IsComplete() {
		panic("response already closed")
	}

	resp.fi = nil
	if resp.writer != nil {
		resp.writer.Close()
		resp.writer = nil
	}
	if resp.HTTPResponse != nil && resp.HTTPResponse.Body != nil {
		resp.HTTPResponse.Body.Close()
	}

	resp.End = time.Now()
	close(resp.Done)
	if resp.cancel != nil {
		resp.cancel()
	}
	return nil
}
