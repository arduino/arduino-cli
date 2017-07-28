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

package releases

import "os"

// Release represents a generic release.
type Release interface {
	// OpenLocalArchiveForDownload opens the local archive file
	// in append mode if it exists, otherwise it creates it. Returns
	// the file pointer.
	OpenLocalArchiveForDownload() (*os.File, error)
	// ArchivePath returns the fullPath of the Archive of this release.
	ArchivePath() (string, error)
	// ExpectedChecksum returns the expected checksum for this release.
	ExpectedChecksum() string
	// ArchiveName returns the archive file name (not the path).
	ArchiveName() string
	// ArchiveURL returns the archive URL.
	ArchiveURL() string
	// ArchiveSize returns the archive size.
	ArchiveSize() int64
	// GetDownloadCacheFolder returns the cache folder of this release.
	// Mostly this is based on the type of release (library, core, tool)
	GetDownloadCacheFolder() (string, error)
}
