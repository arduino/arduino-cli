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

package resources

import (
	"fmt"
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/cavaliercoder/grab"
)

// ArchivePath returns the path of the Archive of the specified DownloadResource relative
// to the specified downloadDir
func (r *DownloadResource) ArchivePath(downloadDir *paths.Path) (*paths.Path, error) {
	staging := downloadDir.Join(r.CachePath)
	if err := staging.MkdirAll(); err != nil {
		return nil, err
	}
	return staging.Join(r.ArchiveFileName), nil
}

// IsCached returns true if the specified DownloadResource has already been downloaded
func (r *DownloadResource) IsCached(downloadDir *paths.Path) (bool, error) {
	archivePath, err := r.ArchivePath(downloadDir)
	if err != nil {
		return false, fmt.Errorf("getting archive path: %s", err)
	}

	if exist, err := archivePath.Exist(); err != nil {
		return false, fmt.Errorf("checking archive existence: %s", err)
	} else {
		return exist, nil
	}
}

// Download a DownloadResource.
func (r *DownloadResource) Download(downloadDir *paths.Path) (*grab.Response, error) {
	cached, err := r.TestLocalArchiveIntegrity(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("testing local archive integrity: %s", err)
	}
	if cached {
		// File is cached, nothing to do here
		return nil, nil
	}

	path, err := r.ArchivePath(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("getting archive path: %s", err)
	}

	if stats, err := path.Stat(); os.IsNotExist(err) {
		// normal download
	} else if err == nil && stats.Size() > r.Size {
		// file is bigger than expected, retry download...
		if err := path.Remove(); err != nil {
			return nil, fmt.Errorf("removing corrupted archive file: %s", err)
		}
	} else if err == nil {
		// resume download
	} else {
		return nil, fmt.Errorf("getting archive file info: %s", err)
	}

	req, err := grab.NewRequest(path.String(), r.URL)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %s", err)
	}
	client := grab.NewClient()
	return client.Do(req), nil
}
