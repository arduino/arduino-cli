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

package cores

import (
	"regexp"
	"strings"
)

// CoreIDTuple represents a tuple to identify a Core
type CoreIDTuple struct {
	Package     string // The package where this core belongs to.
	CoreName    string // The core name.
	CoreVersion string // The version of the core, to get the release.
}

var coreTupleRegexp = regexp.MustCompile("[a-zA-Z0-9]+:[a-zA-Z0-9]+(=([0-9]|[0-9].)*[0-9]+)?")

func ParseArgs(args []string) []CoreIDTuple {
	ret := make([]CoreIDTuple, 0, 5)

	for _, arg := range args {
		if coreTupleRegexp.MatchString(arg) {
			// splits the string according to regexp into its components.
			split := strings.FieldsFunc(arg, func(r rune) bool {
				return r == '=' || r == ':'
			})
			if len(split) < 3 {
				split = append(split, "latest")
			}
			ret = append(ret, CoreIDTuple{
				Package:     split[0],
				CoreName:    split[1],
				CoreVersion: split[2],
			})
		}
	}
	return ret
}
