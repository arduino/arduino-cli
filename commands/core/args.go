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

	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/cores"
)

// platformReference represents a tuple to identify a Platform
type platformReference struct {
	Package              string // The package where this Platform belongs to.
	PlatformArchitecture string
	PlatformVersion      string
}

var coreTupleRegexp = regexp.MustCompile("[a-zA-Z0-9]+:[a-zA-Z0-9]+(=([0-9]|[0-9].)*[0-9]+)?")

// parsePlatformReferenceArgs parses a sequence of "packager:arch=version" tokens and returns a platformReference slice.
//
// If version is not present it is assumed as "latest" version.
func parsePlatformReferenceArgs(args []string) []platformReference {
	ret := []platformReference{}

	for _, arg := range args {
		if coreTupleRegexp.MatchString(arg) {
			// splits the string according to regexp into its components.
			split := strings.FieldsFunc(arg, func(r rune) bool {
				return r == '=' || r == ':'
			})
			if len(split) < 3 {
				split = append(split, "latest")
			}
			ret = append(ret, platformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
				PlatformVersion:      split[2],
			})
		} else {
			// TODO: handle errors properly
			ret = append(ret, platformReference{
				Package:              "invalid-arg",
				PlatformArchitecture: arg,
			})
		}
	}
	return ret
}

// findItemsToDownload takes a set of platformReference and returns a set of items to download and
// a set of outputs for non existing platforms.
func findItemsToDownload(sc *cores.Packages, items []platformReference) ([]*cores.PlatformRelease, []*cores.ToolRelease, []output.ProcessResult) {
	itemC := len(items)
	retPlatforms := []*cores.PlatformRelease{}
	retTools := []*cores.ToolRelease{}
	fails := make([]output.ProcessResult, 0, itemC)

	// value is not used, this map is only to check if an item is inside (set implementation)
	// see https://stackoverflow.com/questions/34018908/golang-why-dont-we-have-a-set-datastructure
	presenceMap := make(map[string]bool, itemC)

	for _, item := range items {
		if item.Package == "invalid-arg" {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    "Invalid item (not PACKAGER:ARCH[=VERSION])",
			})
			continue
		}
		pkg, exists := sc.Packages[item.Package]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    fmt.Sprintf("Package %s not found", item.Package),
			})
			continue
		}
		platform, exists := pkg.Platforms[item.PlatformArchitecture]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    "Platform not found",
			})
			continue
		}

		_, exists = presenceMap[item.PlatformArchitecture]
		if exists { //skip
			continue
		}

		release := platform.GetVersion(item.PlatformVersion)
		if release == nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    fmt.Sprintf("Version %s Not Found", item.PlatformVersion),
			})
			continue
		}

		// replaces "latest" with latest version too
		toolDeps, err := sc.GetDepsOfPlatformRelease(release)
		if err != nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.PlatformArchitecture,
				Error:    fmt.Sprintf("Cannot get tool dependencies of plafotmr %s: %s", platform.Name, err.Error()),
			})
			continue
		}

		retPlatforms = append(retPlatforms, release)

		presenceMap[platform.Name] = true
		for _, tool := range toolDeps {
			if presenceMap[tool.Tool.Name] {
				continue
			}
			presenceMap[tool.Tool.Name] = true
			retTools = append(retTools, tool)
		}
	}
	return retPlatforms, retTools, fails
}
