/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
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
 */

package librariesmanager

import (
	"fmt"
	"net/url"

	"github.com/bcmi-labs/arduino-cli/pathutils"

	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/cavaliercoder/grab"
)

// LibraryIndexURL is the URL where to get library index.
var LibraryIndexURL, _ = url.Parse("http://downloads.arduino.cc/libraries/library_index.json")

// IndexPath returns the path of the library_index.json file.
func IndexPath() pathutils.Path {
	return configs.IndexPath("library_index.json")
}

// DownloadLibrariesFile downloads the libraries index file from Arduino repository.
func DownloadLibrariesFile() (*grab.Response, error) {
	path, err := IndexPath().Get()
	if err != nil {
		return nil, fmt.Errorf("getting library_index.json path: %s", err)
	}
	req, err := grab.NewRequest(path, LibraryIndexURL.String())
	req.NoResume = true
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %s", err)
	}
	client := grab.NewClient()
	return client.Do(req), nil

	// TODO: Download from gzipped URL index
}
