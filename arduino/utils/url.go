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

package utils

import (
	"net/url"
	"runtime"
)

// URLParse parses a raw URL string and handles local files URLs depending on the platform
func URLParse(rawURL string) (*url.URL, error) {
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if URL.Scheme == "file" && runtime.GOOS == "windows" {
		// Parsed local file URLs on Windows are returned with a leading /
		// so we remove it
		URL.Path = URL.Path[1:]
	}
	return URL, nil
}
