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
	"net/url"
	"slices"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"go.bug.st/f"
)

// ProfileLibAdd adds a library to the specified profile or to the default profile.
func (s *arduinoCoreServerImpl) ProfileLibAdd(ctx context.Context, req *rpc.ProfileLibAddRequest) (*rpc.ProfileLibAddResponse, error) {
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

	var addedLibs []*sketch.ProfileLibraryReference
	var skippedLibs []*sketch.ProfileLibraryReference
	if reqLocalLib := req.GetLibrary().GetLocalLibrary(); reqLocalLib != nil {
		// Add a local library
		path := paths.New(reqLocalLib.GetPath())
		if path == nil {
			return nil, &cmderrors.InvalidArgumentError{Message: "invalid library path"}
		}
		addedLib := &sketch.ProfileLibraryReference{InstallDir: path}
		profile.Libraries = append(profile.Libraries, addedLib)
		addedLibs = append(addedLibs, addedLib)
	} else if reqGitLib := req.GetLibrary().GetGitLibrary(); reqGitLib != nil {
		// Add a git library
		gitURL, err := url.Parse(reqGitLib.GetUrl())
		if err != nil {
			return nil, &cmderrors.InvalidURLError{Cause: err}
		}
		addedLib := &sketch.ProfileLibraryReference{GitURL: gitURL}
		profile.Libraries = append(profile.Libraries, addedLib)
		addedLibs = append(addedLibs, addedLib)
	} else if reqIndexLib := req.GetLibrary().GetIndexLibrary(); reqIndexLib != nil {
		// Obtain the library index from the manager
		li, err := instances.GetLibrariesIndex(req.GetInstance())
		if err != nil {
			return nil, err
		}
		version, err := parseVersion(reqIndexLib.GetVersion())
		if err != nil {
			return nil, err
		}
		libRelease, err := li.FindRelease(reqIndexLib.GetName(), version)
		if err != nil {
			return nil, err
		}

		add := func(libReleaseToAdd *librariesindex.Release, isDep bool) {
			libRefToAdd := &sketch.ProfileLibraryReference{
				Library:      libReleaseToAdd.GetName(),
				Version:      libReleaseToAdd.GetVersion(),
				IsDependency: isDep,
			}
			existingLibRef, _ := profile.GetLibrary(libReleaseToAdd.GetName())
			if existingLibRef == nil {
				profile.Libraries = append(profile.Libraries, libRefToAdd)
				addedLibs = append(addedLibs, libRefToAdd)
				return
			}

			// The library is already present in the profile.

			// If the existing library was a dependency, and we are adding a non-dependency,
			// update the flag to indicate that it's not a dependency anymore.
			if !isDep && existingLibRef.IsDependency {
				existingLibRef.IsDependency = false
			}

			// If no-overwrite is specified, skip updating the library version.
			// If the same version of the library has been already added to the profile, skip it.
			if req.GetNoOverwrite() || existingLibRef.Version.Equal(libReleaseToAdd.GetVersion()) {
				skippedLibs = append(skippedLibs, libRefToAdd)
				return
			}

			// otherwise update the version of the library
			existingLibRef.Version = libReleaseToAdd.GetVersion()
			addedLibs = append(addedLibs, existingLibRef)
		}

		if req.GetAddDependencies() {
			libWithDeps, err := libraryResolveDependencies(li, libRelease.GetName(), libRelease.GetVersion().String(), nil)
			// deps contains the main library as well, so we skip it when adding dependencies
			if err != nil {
				return nil, err
			}
			// sort to make the output order deterministic
			slices.SortFunc(libWithDeps, librariesindex.ReleaseCompare)
			for _, lib := range libWithDeps {
				isDep := libRelease.GetName() != lib.GetName()
				add(lib, isDep)
			}
		} else {
			add(libRelease, false)
		}
	} else {
		return nil, &cmderrors.InvalidArgumentError{Message: "library must be specified"}
	}

	projectFilePath := sk.GetProjectPath()
	if err = projectFilePath.WriteFile([]byte(sk.Project.AsYaml())); err != nil {
		return nil, err
	}

	return &rpc.ProfileLibAddResponse{
		AddedLibraries:   f.Map(addedLibs, (*sketch.ProfileLibraryReference).ToRpc),
		SkippedLibraries: f.Map(skippedLibs, (*sketch.ProfileLibraryReference).ToRpc),
		ProfileName:      profileName,
	}, nil
}
