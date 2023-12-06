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

// SanitizeName replaces with underscores all chars that are not included
// in the ranges: 0-9, A-Z, a-z, "-" and "."
func SanitizeName(origName string) string {
	sanitized := ""
	for i, c := range origName {
		if (c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(i > 0 && c == '-') || (i > 0 && c == '.') {
			sanitized = sanitized + string(c)
		} else {
			sanitized = sanitized + "_"
		}
	}

	if len(sanitized) > 63 {
		sanitized = sanitized[:64]
	}
	return sanitized
}
