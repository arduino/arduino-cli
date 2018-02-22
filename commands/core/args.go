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
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/cores"
)

// downloadItem represents a core or tool download
// TODO: can be greatly simplified by usign directly a pointer to the Platform or Tool
type downloadItem struct {
	Package string
	releases.DownloadItem
}

// platformReference represents a tuple to identify a Platform
type platformReference struct {
	Package     string // The package where this core belongs to.
	CoreName    string // The core name.
	CoreVersion string // The version of the core, to get the release.
}

var coreTupleRegexp = regexp.MustCompile("[a-zA-Z0-9]+:[a-zA-Z0-9]+(=([0-9]|[0-9].)*[0-9]+)?")

// parsePlatformReferenceArgs parses a sequence of "packager:name=version" tokens and returns a CoreIDTuple slice.
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
				Package:     split[0],
				CoreName:    split[1],
				CoreVersion: split[2],
			})
		} else {
			ret = append(ret, platformReference{
				Package:  "invalid-arg",
				CoreName: arg,
			})
		}
	}
	return ret
}

// findDownloadItems takes a set of platformReference and returns a set of items to download and
// a set of outputs for non existing platforms.
func findDownloadItems(sc *cores.StatusContext, items []platformReference) ([]downloadItem, []downloadItem, []output.ProcessResult) {
	itemC := len(items)
	retCores := make([]downloadItem, 0, itemC)
	retTools := make([]downloadItem, 0, itemC)
	fails := make([]output.ProcessResult, 0, itemC)

	// value is not used, this map is only to check if an item is inside (set implementation)
	// see https://stackoverflow.com/questions/34018908/golang-why-dont-we-have-a-set-datastructure
	presenceMap := make(map[string]bool, itemC)

	for _, item := range items {
		if item.Package == "invalid-arg" {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    "Invalid item (not PACKAGER:CORE[=VERSION])",
			})
			continue
		}
		pkg, exists := sc.Packages[item.Package]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Package %s not found", item.Package),
			})
			continue
		}
		core, exists := pkg.Plaftorms[item.CoreName]
		if !exists {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    "Core not found",
			})
			continue
		}

		_, exists = presenceMap[item.CoreName]
		if exists { //skip
			continue
		}

		release := core.GetVersion(item.CoreVersion)
		if release == nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Version %s Not Found", item.CoreVersion),
			})
			continue
		}

		// replaces "latest" with latest version too
		deps, err := sc.GetDepsOfPlatformRelease(release)
		if err != nil {
			fails = append(fails, output.ProcessResult{
				ItemName: item.CoreName,
				Error:    fmt.Sprintf("Cannot get tool dependencies of %s core: %s", core.Name, err.Error()),
			})
			continue
		}

		retCores = append(retCores, downloadItem{
			Package: pkg.Name,
			DownloadItem: releases.DownloadItem{
				Name:    core.Architecture,
				Release: release,
			},
		})

		presenceMap[core.Name] = true
		for _, tool := range deps {
			_, exists = presenceMap[tool.ToolName]
			if exists { //skip
				continue
			}

			presenceMap[tool.ToolName] = true
			retTools = append(retTools, downloadItem{
				Package: pkg.Name,
				DownloadItem: releases.DownloadItem{
					Name:    tool.ToolName,
					Release: tool.Release,
				},
			})
		}
	}
	return retCores, retTools, fails
}
