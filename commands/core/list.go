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

package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// GetPlatforms returns a list of installed platforms, optionally filtered by
// those requiring an update.
func GetPlatforms(req *rpc.PlatformListRequest) ([]*rpc.Platform, error) {
	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	res := []*rpc.Platform{}
	for _, targetPackage := range pme.GetPackages() {
		for _, platform := range targetPackage.Platforms {
			platformRelease := pme.GetInstalledPlatformRelease(platform)

			// The All flags adds to the list of installed platforms the installable platforms (from the indexes)
			// If both All and UpdatableOnly are set All takes precedence
			if req.All {
				installedVersion := ""
				if platformRelease == nil { // if the platform is not installed
					platformRelease = platform.GetLatestRelease()
				} else {
					installedVersion = platformRelease.Version.String()
				}
				// it could happen, especially with indexes not perfectly compliant with specs that a platform do not contain a valid release
				if platformRelease != nil {
					rpcPlatform := commands.PlatformReleaseToRPC(platformRelease)
					rpcPlatform.Installed = installedVersion
					res = append(res, rpcPlatform)
					continue
				}
			}

			if platformRelease != nil {
				latest := platform.GetLatestRelease()
				if latest == nil {
					return nil, &arduino.PlatformNotFoundError{Platform: platform.String(), Cause: fmt.Errorf("the platform has no releases")}
				}

				// show only the updatable platforms
				if req.UpdatableOnly && latest == platformRelease {
					continue
				}

				rpcPlatform := commands.PlatformReleaseToRPC(platformRelease)
				rpcPlatform.Installed = platformRelease.Version.String()
				rpcPlatform.Latest = latest.Version.String()
				res = append(res, rpcPlatform)
			}
		}
	}
	// Sort result alphabetically and put deprecated platforms at the bottom
	sort.Slice(res, func(i, j int) bool {
		return strings.ToLower(res[i].Name) < strings.ToLower(res[j].Name)
	})
	sort.SliceStable(res, func(i, j int) bool {
		if !res[i].Deprecated && res[j].Deprecated {
			return true
		}
		return false
	})
	return res, nil
}
