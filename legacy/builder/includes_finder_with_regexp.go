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

package builder

import (
	"regexp"
	"strings"
)

var INCLUDE_REGEXP = regexp.MustCompile("(?ms)^\\s*#[ \t]*include\\s*[<\"](\\S+)[\">]")

func IncludesFinderWithRegExp(source string) string {
	match := INCLUDE_REGEXP.FindStringSubmatch(source)
	if match != nil {
		return strings.TrimSpace(match[1])
	}
	return findIncludeForOldCompilers(source)
}

func findIncludeForOldCompilers(source string) string {
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		splittedLine := strings.Split(line, ":")
		for i := range splittedLine {
			if strings.Contains(splittedLine[i], "fatal error") {
				return strings.TrimSpace(splittedLine[i+1])
			}
		}
	}
	return ""
}
