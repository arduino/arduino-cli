# grab

[![GoDoc](https://godoc.org/github.com/cavaliercoder/grab?status.svg)](https://godoc.org/github.com/cavaliercoder/grab) [![Build Status](https://travis-ci.org/cavaliercoder/grab.svg)](https://travis-ci.org/cavaliercoder/grab) [![Go Report Card](https://goreportcard.com/badge/github.com/cavaliercoder/grab)](https://goreportcard.com/report/github.com/cavaliercoder/grab)

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

Requires Go v1.7+

## Older versions

If you are using an older version of Go, or require previous versions of the
Grab API, you can import older version of this package, thanks to gpkg.in.
Please see all GitHub tags for available versions.

	$ go get gopkg.in/cavaliercoder/grab.v1


## Example

The following example downloads a PDF copy of the free eBook, "An Introduction
to Programming in Go" and periodically prints the download progress until it is
complete.

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
