// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"cmp"
	"context"
	"slices"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

// ProfileLibRemove removes a library from the specified profile or from the default profile.
func (s *arduinoCoreServerImpl) ProfileLibRemove(ctx context.Context, req *rpc.ProfileLibRemoveRequest) (*rpc.ProfileLibRemoveResponse, error) {
	// Returns an error if the main file is missing from the sketch so there is no need to check if the path exists
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, err
	}

	// If no profile is specified, try to use the default one
	profileName := cmp.Or(req.GetProfileName(), sk.Project.DefaultProfile)
	if profileName == "" {
		return nil, &cmderrors.MissingProfileError{}
	}

	profile, err := sk.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	var removedLibraries []*rpc.ProfileLibraryReference
	remove := func(libraryToRemove *sketch.ProfileLibraryReference) error {
		removedLibrary, err := profile.RemoveLibrary(libraryToRemove)
		if err != nil {
			return &cmderrors.InvalidArgumentError{Message: "could not remove library", Cause: err}
		}
		removedLibraries = append(removedLibraries, removedLibrary.ToRpc())
		return nil
	}

	libToRemove, err := sketch.FromRpcProfileLibraryReference(req.GetLibrary())
	if err != nil {
		return nil, &cmderrors.InvalidArgumentError{Message: "invalid library reference", Cause: err}
	}
	if err := remove(libToRemove); err != nil {
		return nil, err
	}

	// Get the dependencies of the libraries to see if any of them could be removed as well
	if req.GetRemoveDependencies() {
		if req.GetLibrary().GetIndexLibrary() == nil {
			// No dependencies to remove
			return nil, &cmderrors.InvalidArgumentError{Message: "automatic dependency removal is supported only for IndexLibraries"}
		}
		// Obtain the library index from the manager
		li, err := instances.GetLibrariesIndex(req.GetInstance())
		if err != nil {
			return nil, err
		}

		// Get all the dependencies required by the profile excluding the removed library
		requiredDeps := map[string]bool{}
		for _, profLib := range profile.Libraries {
			if profLib.IsDependency {
				continue
			}
			if profLib.Library == "" {
				continue
			}
			deps, err := libraryResolveDependencies(li, profLib.Library, profLib.Version.String(), nil)
			if err != nil {
				return nil, &cmderrors.InvalidArgumentError{Cause: err, Message: "cannot resolve dependencies for installed libraries"}
			}
			for _, dep := range deps {
				requiredDeps[dep.Library.Name] = true
			}
		}

		candidateDepsToRemove, err := libraryResolveDependencies(li, libToRemove.Library, libToRemove.Version.String(), nil)
		if err != nil {
			return nil, &cmderrors.InvalidArgumentError{Cause: err, Message: "cannot resolve dependencies for installed libraries"}
		}
		// sort to make the output order deterministic
		slices.SortFunc(candidateDepsToRemove, librariesindex.ReleaseCompare)
		for _, depToRemove := range candidateDepsToRemove {
			if requiredDeps[depToRemove.Library.Name] {
				continue
			}
			_ = remove(&sketch.ProfileLibraryReference{Library: depToRemove.Library.Name, Version: depToRemove.Version})
		}
	}

	projectFilePath := sk.GetProjectPath()
	if err = projectFilePath.WriteFile([]byte(sk.Project.AsYaml())); err != nil {
		return nil, &cmderrors.CantUpdateSketchError{Cause: err}
	}

	return &rpc.ProfileLibRemoveResponse{
		RemovedLibraries: removedLibraries,
		ProfileName:      profileName,
	}, nil
}
