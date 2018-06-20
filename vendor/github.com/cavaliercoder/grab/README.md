# grab

[![GoDoc](https://godoc.org/github.com/cavaliercoder/grab?status.svg)](https://godoc.org/github.com/cavaliercoder/grab) [![Build Status](https://travis-ci.org/cavaliercoder/grab.svg?branch=master)](https://travis-ci.org/cavaliercoder/grab) [![Go Report Card](https://goreportcard.com/badge/github.com/cavaliercoder/grab)](https://goreportcard.com/report/github.com/cavaliercoder/grab)

*Downloading the internet, one goroutine at a time!*

	$ go get github.com/cavaliercoder/grab

Grab is a Go package for downloading files from the internet with the following
rad features:

* Monitor download progress concurrently
* Auto-resume incomplete downloads
* Guess filename from content header or URL path
* Safely cancel downloads using context.Context
* Validate downloads using checksums
* Download batches of files concurrently
* Apply rate limiters

Requires Go v1.7+

## Example

The following example downloads a PDF copy of the free eBook, "An Introduction
to Programming in Go" into the current working directory.

```go
resp, err := grab.Get(".", "http://www.golang-book.com/public/pdf/gobook.pdf")
if err != nil {
	log.Fatal(err)
}

fmt.Println("Download saved to", resp.Filename)
```

The following, more complete example allows for more granular control and
periodically prints the download progress until it is complete.

The second time you run the example, it will auto-resume the previous download
and exit sooner.

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cavaliercoder/grab"
)

func main() {
	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", "http://www.golang-book.com/public/pdf/gobook.pdf")

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)

	// Output:
	// Downloading http://www.golang-book.com/public/pdf/gobook.pdf...
	//   200 OK
	//   transferred 42970 / 2893557 bytes (1.49%)
	//   transferred 1207474 / 2893557 bytes (41.73%)
	//   transferred 2758210 / 2893557 bytes (95.32%)
	// Download saved to ./gobook.pdf
}
```

## Design trade-offs

The primary use case for Grab is to concurrently downloading thousands of large
files from remote file repositories where the remote files are immutable.
Examples include operating system package repositories or ISO libraries.

Grab aims to provide robust, sane defaults. These are usually determined using
the HTTP specifications, or by mimicking the behavior of common web clients like
cURL, wget and common web browsers.

Grab aims to be stateless. The only state that exists is the remote files you
wish to download and the local copy which may be completed, partially completed
or not yet created. The advantage to this is that the local file system is not
cluttered unnecessarily with addition state files (like a `.crdownload` file).
The disadvantage of this approach is that grab must make assumptions about the
local and remote state; specifically, that they have not been modified by
another program.

If the local or remote file are modified outside of grab, and you download the
file again with resuming enabled, the local file will likely become corrupted.
In this case, you might consider making remote files immutable, or disabling
resume.

Grab aims to enable best-in-class functionality for more complex features
through extensible interfaces, rather than reimplementation. For example,
you can provide your own Hash algorithm to compute file checksums, or your
own rate limiter implementation (with all the associated trade-offs) to rate
limit downloads.
