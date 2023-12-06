// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package resources

// DownloadResource has all the information to download a file
type DownloadResource struct {
	URL             string
	ArchiveFileName string
	Checksum        string
	Size            int64
	CachePath       string
}

// DownloadResult contains the result of a download
type DownloadResult struct {
	// Error is nil if the download is successful otherwise
	// it contains the reason for the download failure
	Error error

	// AlreadyDownloaded is true if a cache file is found
	// and, consequently, the download has not been executed
	AlreadyDownloaded bool
}
