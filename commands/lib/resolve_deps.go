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

package lib

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// LibraryResolveDependencies FIXMEDOC
func LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesReq) (*rpc.LibraryResolveDependenciesResp, error) {
	lm := commands.GetLibraryManager(req.GetInstance().GetId())

	// Search the requested lib
	reqLibRelease, err := findLibraryIndexRelease(lm, req)
	if err != nil {
		return nil, fmt.Errorf("looking for library: %s", err)
	}

	// Extract all installed libraries
	installedLibs := map[string]*libraries.Library{}
	for _, lib := range listLibraries(lm, false, false) {
		installedLibs[lib.Library.Name] = lib.Library
	}

	// Resolve all dependencies...
	deps := lm.Index.ResolveDependencies(reqLibRelease)

	// If no solution has been found
	if len(deps) == 0 {
		// Check if there is a problem with the first level deps
		for _, directDep := range reqLibRelease.GetDependencies() {
			if _, ok := lm.Index.Libraries[directDep.GetName()]; !ok {
				return nil, fmt.Errorf("dependency '%s' is not available", directDep.GetName())
			}
		}

		// Otherwise there is no possible solution, the depends field has an invalid formula
		return nil, fmt.Errorf("no valid solution found")
	}

	res := []*rpc.LibraryDependencyStatus{}
	for _, dep := range deps {
		// ...and add information on currently installed versions of the libraries
		installed := ""
		if installedLib, has := installedLibs[dep.GetName()]; has {
			installed = installedLib.Version.String()
		}
		res = append(res, &rpc.LibraryDependencyStatus{
			Name:             dep.GetName(),
			VersionRequired:  dep.GetVersion().String(),
			VersionInstalled: installed,
		})
	}
	return &rpc.LibraryResolveDependenciesResp{Dependencies: res}, nil
}
