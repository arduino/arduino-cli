/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/commands"

	"github.com/arduino/arduino-cli/common/formatter"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"go.bug.st/relaxed-semver"
)

// parsePlatformReferenceArgs parses a sequence of "packager:arch@version" tokens and returns a platformReference slice.
func parsePlatformReferenceArgs(args []string) []*packagemanager.PlatformReference {
	ret := []*packagemanager.PlatformReference{}
	for _, arg := range args {
		reference, err := parsePlatformReferenceArg(arg)
		if err != nil {
			formatter.PrintError(err, "Invalid item "+arg)
			os.Exit(commands.ErrBadArgument)
		}
		ret = append(ret, reference)
	}
	return ret
}

func parsePlatformReferenceArg(arg string) (*packagemanager.PlatformReference, error) {
	split := strings.SplitN(arg, "@", 2)
	arg = split[0]
	var version *semver.Version
	if len(split) > 1 {
		if ver, err := semver.Parse(split[1]); err == nil {
			version = ver
		} else {
			return nil, fmt.Errorf("invalid version: %s", err)
		}
	}
	split = strings.Split(arg, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid item %s", arg)
	}
	return &packagemanager.PlatformReference{
		Package:              split[0],
		PlatformArchitecture: split[1],
		PlatformVersion:      version,
	}, nil
}
