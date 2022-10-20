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
	"errors"
	"sort"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// LibraryResolveDependencies FIXMEDOC
func LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesRequest) (*rpc.LibraryResolveDependenciesResponse, error) {
	lm := commands.GetLibraryManager(req)
	if lm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	// Search the requested lib
	reqLibRelease, err := findLibraryIndexRelease(lm, req)
	if err != nil {
		return nil, err
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
				err := errors.New(tr("dependency '%s' is not available", directDep.GetName()))
				return nil, &arduino.LibraryDependenciesResolutionFailedError{Cause: err}
			}
		}

		// Otherwise there is no possible solution, the depends field has an invalid formula
		return nil, &arduino.LibraryDependenciesResolutionFailedError{}
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
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return &rpc.LibraryResolveDependenciesResponse{Dependencies: res}, nil
}
