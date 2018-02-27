/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package common

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/bcmi-labs/arduino-cli/pathutils"
)

// DownloadIndex is a function to download a generic index.
func DownloadIndex(indexPath pathutils.Path, URL string) error {
	file, err := indexPath.Get()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// HandleResultFunc defines a function able to handle the content of the
// download stream of the package (DownloadPackage), filling the File with the content
// of the Reader, starting from the initial position
type HandleDownloadPackageResultFunc func(io.Reader, *os.File, int) error

// DefaultDownloadHandlerFunc is the default HandleDownloadPackageResultFunc, which
// simply copies the content of the Reader into the File, starting from the initialSize
func DefaultDownloadHandlerFunc(source io.Reader, initialData *os.File, initialSize int) error {
	// Copy the file content
	_, err := io.Copy(initialData, source)
	return err
}

// DownloadPackageProgressChangedHandler defines a function able to handle the update
// of the progress of the current download
type DownloadPackageProgressChangedHandler func(fileSize int64, downloadedSoFar int64)

// DownloadPackage downloads a package from Arduino repository.
// Besides the download information (URL, initialData and totalSize), two external handlers are available for:
//  - (handleResultFunc) handling the result of the download (i.e. decide how to copy the download to the file
// 	  or do something weird during the operation)
//  - (progressChangedHandler) handling the download progress change (and perhaps display it somehow)
// None of the handlers is mandatory; they won't be used if nil.
func DownloadPackage(URL string, initialData *os.File, totalSize int64, handleResultFunc HandleDownloadPackageResultFunc,
	progressChangedHandler DownloadPackageProgressChangedHandler) error {

	if initialData == nil {
		return errors.New("Cannot fill a nil file pointer")
	}

	client := &http.Client{}

	var initialSize int64
	stats, err := initialData.Stat()
	if err != nil {
		initialSize = 0
	} else {
		fileSize := stats.Size()
		if fileSize >= totalSize {
			initialSize = 0
		} else {
			initialSize = fileSize
		}
	}

	client.Timeout = time.Duration(totalSize-initialSize) / 57344 * time.Second // size to download divided by 56KB/s (56 * 1024 = 57344)

	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return fmt.Errorf("Cannot create HTTP to URL %s request: %s", URL, err)
	}

	if initialSize > 0 {
		request.Header.Add("Range", fmt.Sprintf("bytes=%d-", initialSize))
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Cannot fetch %s Response creation error: %s",
			URL, err)
	} else if response.StatusCode != 200 &&
		response.StatusCode != 206 &&
		response.StatusCode != 416 {
		response.Body.Close()
		return fmt.Errorf("Cannot fetch %s Source responded with code %d",
			URL, response.StatusCode)
	}
	defer response.Body.Close()

	// Handle the progress update handler, by creating a ProgressProxyReader;
	// only if it's needed (i.e. we actually have an external progressChangedHandler)
	progressProxyReader := response.Body
	downloadedSoFar := initialSize
	if progressChangedHandler != nil {
		progressProxyReader = ProgressProxyReader{response.Body, func(progressDelta int64) {
			// WARNING: This is using a closure on downloadedSoFar!
			downloadedSoFar += progressDelta
			progressChangedHandler(totalSize, downloadedSoFar)
		},
		}
	}

	// Use the external handleResultFunc, if available, or the default one otherwise
	if handleResultFunc == nil {
		handleResultFunc = DefaultDownloadHandlerFunc
	}

	err = handleResultFunc(progressProxyReader, initialData, int(initialSize))
	if err != nil {
		return fmt.Errorf("Cannot read response body from %s : %s", URL, err)
	}
	return nil
}

// FIXME: Move outside? perhaps in commons?
// HandleProgressUpdateFunc defines a function able to handle the progressDelta, in bytes
type HandleProgressUpdateFunc func(progressDelta int64)

// It's proxy reader, intercepting reads to post progress updates, implement io.Reader
type ProgressProxyReader struct {
	io.Reader
	handleProgressUpdateFunc HandleProgressUpdateFunc
}

func (r ProgressProxyReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.handleProgressUpdateFunc(int64(n))
	return
}

// Close the reader when it implements io.Closer
func (r ProgressProxyReader) Close() (err error) {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return
}
