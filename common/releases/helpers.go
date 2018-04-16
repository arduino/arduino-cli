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

package releases

import (
	"fmt"
	"os"

	"github.com/cavaliercoder/grab"
)

// Download a DownloadResource.
func (r *DownloadResource) Download() (*grab.Response, error) {
	cached, err := r.TestLocalArchiveIntegrity()
	if err != nil {
		return nil, fmt.Errorf("testing local archive integrity: %s", err)
	}
	if cached {
		// File is cached, nothing to do here
		return nil, nil
	}

	path, err := r.ArchivePath()
	if err != nil {
		return nil, fmt.Errorf("getting archive path: %s", err)
	}

	if stats, err := os.Stat(path); os.IsNotExist(err) {
		// normal download
	} else if err == nil && stats.Size() >= r.Size {
		// file is bigger than expected, retry download...
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("removing corrupted archive file: %s", err)
		}
	} else if err == nil {
		// resume download
	} else {
		return nil, fmt.Errorf("getting archive file info: %s", err)
	}

	req, err := grab.NewRequest(path, r.URL)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %s", err)
	}
	client := grab.NewClient()
	return client.Do(req), nil
}
