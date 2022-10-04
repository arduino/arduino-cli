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

package commands

import "strings"

// SearchTermsFromQueryString returns the terms inside the query string.
// All non alphanumeric characters (expect ':') are considered separators.
// All search terms are converted to lowercase.
func SearchTermsFromQueryString(query string) []string {
	// Split on anything but 0-9, a-z or :
	return strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || r == ':')
	})
}
