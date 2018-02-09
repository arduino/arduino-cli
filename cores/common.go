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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common"
)

// CoreIDTuple represents a tuple to identify a Core
type CoreIDTuple struct {
	Package     string // The package where this core belongs to.
	CoreName    string // The core name.
	CoreVersion string // The version of the core, to get the release.
}

var coreTupleRegexp = regexp.MustCompile("[a-zA-Z0-9]+:[a-zA-Z0-9]+(=([0-9]|[0-9].)*[0-9]+)?")

// ParseArgs parses a sequence of "packager:name=version" tokens and returns a CoreIDTuple slice.
//
// If version is not present it is assumed as "latest" version.
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
		} else {
			ret = append(ret, CoreIDTuple{
				Package:  "invalid-arg",
				CoreName: arg,
			})
		}
	}
	return ret
}

// IsCoreInstalled detects if a core has been installed.
func IsCoreInstalled(packageName string, name string) (bool, error) {
	location, err := common.CoresFolder(packageName).Get()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filepath.Join(location, name))
	if !os.IsNotExist(err) {
		return true, nil
	}
	return false, nil
}

// IsToolInstalled detects if a tool has been installed.
func IsToolInstalled(packageName string, name string) (bool, error) {
	location, err := common.ToolsFolder(packageName).Get()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filepath.Join(location, name))
	if !os.IsNotExist(err) {
		return true, nil
	}
	return false, nil
}

// IsToolVersionInstalled detects if a specific version of a tool has been installed.
func IsToolVersionInstalled(packageName string, name string, version string) (bool, error) {
	location, err := common.ToolsFolder(packageName).Get()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filepath.Join(location, name, version))
	if !os.IsNotExist(err) {
		return true, nil
	}
	return false, nil
}

// GetLatestInstalledCoreVersion returns the latest version of an installed core.
func GetLatestInstalledCoreVersion(packageName string, name string) (string, error) {
	location, err := common.CoresFolder(packageName).Get()
	if err != nil {
		return "", err
	}
	return getLatestInstalledVersion(location, name)
}

// GetLatestInstalledToolVersion returns the latest version of an installed tool.
func GetLatestInstalledToolVersion(packageName string, name string) (string, error) {
	location, err := common.ToolsFolder(packageName).Get()
	if err != nil {
		return "", err
	}
	return getLatestInstalledVersion(location, name)
}

func getLatestInstalledVersion(location string, name string) (string, error) {
	var versions []string
	root := filepath.Join(location, name)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == root {
			return nil
		}
		versions = append(versions, info.Name())
		return filepath.SkipDir
	})
	if err != nil {
		return "", err
	}

	if len(versions) > 0 {
		sort.Strings(versions)
		return versions[len(versions)-1], nil
	}
	return "", errors.New("no versions found")
}
