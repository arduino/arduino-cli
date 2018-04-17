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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package core

import (
	"fmt"
	"regexp"
	"strings"

	"os"

	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
)

var coreTupleRegexp = regexp.MustCompile("[a-zA-Z0-9]+:[a-zA-Z0-9]+(=([0-9]|[0-9].)*[0-9]+)?")

// parsePlatformReferenceArgs parses a sequence of "packager:arch=version" tokens and returns a platformReference slice.
//
// If version is not present it is assumed as "latest" version.
func parsePlatformReferenceArgs(args []string) []packagemanager.PlatformReference {
	ret := []packagemanager.PlatformReference{}

	for _, arg := range args {
		if coreTupleRegexp.MatchString(arg) {
			// splits the string according to regexp into its components.
			split := strings.FieldsFunc(arg, func(r rune) bool {
				return r == '=' || r == ':'
			})
			if len(split) < 3 {
				split = append(split, "latest")
			}
			ret = append(ret, packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
				PlatformVersion:      split[2],
			})
		} else {
			// Why even bother to return an error; just fail and let the user know right away...
			formatter.PrintError(nil,
				fmt.Sprintf("'%s' is an invalid item (does not match the syntax 'PACKAGER:ARCH[=VERSION]')", arg))
			os.Exit(commands.ErrBadArgument)
		}
	}
	return ret
}
