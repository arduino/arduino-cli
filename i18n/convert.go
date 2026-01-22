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

package i18n

import (
	"regexp"
	"strconv"
	"strings"
)

var javaFormatPlaceholderRegexp = regexp.MustCompile(`{(\d)}`)

// FromJavaToGoSyntax convert a translation string made for Java to a one suitable for golang (printf-style).
// The conversion transforms java placeholders like "{0}","{1}","{2}",etc... with the equivalent for golang
// "%[1]v","%[2]v","%[3]v",etc...
// The double single-quote "”" is translated into a single single-quote "'".
func FromJavaToGoSyntax(s string) string {
	// Replace "{x}" => "%[x+1]v"
	for _, submatch := range javaFormatPlaceholderRegexp.FindAllStringSubmatch(s, -1) {
		idx, err := strconv.Atoi(submatch[1])
		if err != nil {
			panic(err)
		}
		s = strings.Replace(s, submatch[0], "%["+strconv.Itoa(idx+1)+"]v", -1)
	}

	// Replace "''" => "'"
	s = strings.Replace(s, "''", "'", -1)

	return s
}
