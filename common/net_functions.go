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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
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
)

// DownloadIndex is a function to download a generic index.
func DownloadIndex(indexPathFunc func() (string, error), URL string) error {
	file, err := indexPathFunc()
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

// DownloadPackage downloads a package from arduino repository.
func DownloadPackage(URL string, initialData *os.File, totalSize int64, handleResultFunc func(io.Reader, *os.File, int) error) error {
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

	if handleResultFunc == nil {
		_, err = io.Copy(initialData, response.Body)
	} else {
		err = handleResultFunc(response.Body, initialData, int(initialSize))
	}
	if err != nil {
		return fmt.Errorf("Cannot read response body from %s : %s", URL, err)
	}
	return nil
}
