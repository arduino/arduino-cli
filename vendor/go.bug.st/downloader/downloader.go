//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Downloader is an asynchronous downloader
type Downloader struct {
	URL           string
	Done          chan bool
	resp          *http.Response
	out           *os.File
	completed     int64
	completedLock sync.Mutex
	size          int64
	err           error
}

// DownloadOptions are optional flags that can be passed to Download function
type DownloadOptions int

const (
	// NoResume will not try to resume a partial download
	NoResume DownloadOptions = iota
)

// Close the download
func (d *Downloader) Close() error {
	err1 := d.out.Close()
	err2 := d.resp.Body.Close()
	if err1 != nil {
		return fmt.Errorf("closing output file: %s", err1)
	}
	if err2 != nil {
		return fmt.Errorf("closing input stream: %s", err2)
	}
	return nil
}

// Size return the size of the download
func (d *Downloader) Size() int64 {
	return d.size
}

// RunAndPoll starts the downloader copy-loop and calls the poll function every
// interval time to update progress.
func (d *Downloader) RunAndPoll(poll func(current int64), interval time.Duration) error {
	t := time.NewTicker(interval)
	defer t.Stop()

	go d.AsyncRun()
	for {
		select {
		case <-t.C:
			poll(d.Completed())
		case <-d.Done:
			poll(d.Completed())
			return d.Error()
		}
	}
}

// AsyncRun starts the downloader copy-loop. This function is supposed to be run
// on his own go routine because it sends a confirmation on the Done channel
func (d *Downloader) AsyncRun() {
	in := d.resp.Body
	buff := [4096]byte{}
	for {
		n, err := in.Read(buff[:])
		if n > 0 {
			d.out.Write(buff[:n])
			d.completedLock.Lock()
			d.completed += int64(n)
			d.completedLock.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			d.err = err
			break
		}
	}
	d.Close()
	d.Done <- true
}

// Run starts the downloader and waits until it completes the download.
func (d *Downloader) Run() error {
	go d.AsyncRun()
	<-d.Done
	return d.Error()
}

// Error returns the error during download or nil if no errors happened
func (d *Downloader) Error() error {
	return d.err
}

// Completed returns the bytes read so far
func (d *Downloader) Completed() int64 {
	d.completedLock.Lock()
	res := d.completed
	d.completedLock.Unlock()
	return res
}

// Download returns an asynchronous downloader that will donwload the specified url
// in the specified file. A download resume is tried if a file shorter than the requested
// url is already present.
func Download(file string, url string, options ...DownloadOptions) (*Downloader, error) {
	noResume := false
	for _, opt := range options {
		if opt == NoResume {
			noResume = true
		}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("setting up HTTP request: %s", err)
	}

	var completed int64
	if !noResume {
		if info, err := os.Stat(file); err == nil {
			completed = info.Size()
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", completed))
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// TODO: if file size == header size return nil, nil

	flags := os.O_WRONLY
	if completed == 0 {
		flags |= os.O_CREATE
	} else {
		flags |= os.O_APPEND
	}
	f, err := os.OpenFile(file, flags, 0644)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("opening %s for writing: %s", file, err)
	}

	d := &Downloader{
		URL:       url,
		Done:      make(chan bool),
		resp:      resp,
		out:       f,
		completed: completed,
		size:      resp.ContentLength + completed,
	}
	return d, nil
}
