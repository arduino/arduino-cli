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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// LibraryResolveDependencies FIXMEDOC
func LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesRequest) (*rpc.LibraryResolveDependenciesResponse, error) {
	lme, release, err := instances.GetLibraryManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return nil, err
	}

	return libraryResolveDependencies(ctx, lme, li, req.GetName(), req.GetVersion(), req.GetDoNotUpdateInstalledLibraries())
}

func libraryResolveDependencies(ctx context.Context, lme *librariesmanager.Explorer, li *librariesindex.Index,
	reqName, reqVersion string, noOverwrite bool) (*rpc.LibraryResolveDependenciesResponse, error) {
	version, err := commands.ParseVersion(reqVersion)
	if err != nil {
		return nil, err
	}

	// Search the requested lib
	reqLibRelease, err := li.FindRelease(reqName, version)
	if err != nil {
		return nil, err
	}

	// Extract all installed libraries
	installedLibs := map[string]*libraries.Library{}
	for _, lib := range listLibraries(lme, li, false, false) {
		installedLibs[lib.Library.Name] = lib.Library
	}

	// Resolve all dependencies...
	var overrides []*librariesindex.Release
	if noOverwrite {
		libs := lme.FindAllInstalled()
		libs = libs.FilterByVersionAndInstallLocation(nil, libraries.User)
		for _, lib := range libs {
			if release, err := li.FindRelease(lib.Name, lib.Version); err == nil {
				overrides = append(overrides, release)
			}
		}
	}
	deps := li.ResolveDependencies(reqLibRelease, overrides)

	// If no solution has been found
	if len(deps) == 0 {
		// Check if there is a problem with the first level deps
		for _, directDep := range reqLibRelease.GetDependencies() {
			if _, ok := li.Libraries[directDep.GetName()]; !ok {
				err := errors.New(tr("dependency '%s' is not available", directDep.GetName()))
				return nil, &cmderrors.LibraryDependenciesResolutionFailedError{Cause: err}
			}
		}

		// Otherwise there is no possible solution, the depends field has an invalid formula
		return nil, &cmderrors.LibraryDependenciesResolutionFailedError{}
	}

	res := []*rpc.LibraryDependencyStatus{}
	for _, dep := range deps {
		// ...and add information on currently installed versions of the libraries
		var installed *semver.Version
		required := dep.GetVersion()
		if installedLib, has := installedLibs[dep.GetName()]; has {
			installed = installedLib.Version
			if installed != nil && required != nil && installed.Equal(required) {
				// avoid situations like installed=0.53 and required=0.53.0
				required = installed
			}
		}
		res = append(res, &rpc.LibraryDependencyStatus{
			Name:             dep.GetName(),
			VersionRequired:  required.String(),
			VersionInstalled: installed.String(),
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].GetName() < res[j].GetName()
	})
	return &rpc.LibraryResolveDependenciesResponse{Dependencies: res}, nil
}
