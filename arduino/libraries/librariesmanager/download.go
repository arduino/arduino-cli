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

package librariesmanager

import (
	"net/url"
)

// LibraryIndexURL is the URL where to get the library index.
var LibraryIndexURL, _ = url.Parse("https://downloads.arduino.cc/libraries/library_index.json")

// LibraryIndexGZURL is the URL where to get the gzipped library index.
var LibraryIndexGZURL, _ = url.Parse("https://downloads.arduino.cc/libraries/library_index.json.gz")

// LibraryIndexSignature is the URL where to get the library index signature.
var LibraryIndexSignature, _ = url.Parse("https://downloads.arduino.cc/libraries/library_index.json.sig")

// LibraryIndexWithSignatureArchiveURL is the URL where to get the library index.
var LibraryIndexWithSignatureArchiveURL, _ = url.Parse("https://downloads.arduino.cc/libraries/library_index.tar.bz2")
