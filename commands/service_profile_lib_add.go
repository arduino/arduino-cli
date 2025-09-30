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
	"context"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

// ProfileLibAdd adds a library to the specified profile or to the default profile.
func (s *arduinoCoreServerImpl) ProfileLibAdd(ctx context.Context, req *rpc.ProfileLibAddRequest) (*rpc.ProfileLibAddResponse, error) {
	// Returns an error if the main file is missing from the sketch so there is no need to check if the path exists
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, err
	}
	projectFilePath := sk.GetProjectPath()

	// If no profile is specified, try to use the default one
	profileName := sk.Project.DefaultProfile
	if req.GetProfileName() != "" {
		profileName = req.GetProfileName()
	}
	if profileName == "" {
		return nil, &cmderrors.MissingProfileError{}
	}

	profile, err := sk.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	var addedLib *sketch.ProfileLibraryReference
	if reqLocalLib := req.GetLibrary().GetLocalLibrary(); reqLocalLib != nil {
		// Add a local library
		path := paths.New(reqLocalLib.GetPath())
		if path == nil {
			return nil, &cmderrors.InvalidArgumentError{Message: "invalid library path"}
		}
		addedLib = &sketch.ProfileLibraryReference{InstallDir: path}
		profile.Libraries = append(profile.Libraries, addedLib)
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
		// If the library has been already added to the profile, just update the version
		if lib, _ := profile.GetLibrary(reqIndexLib.GetName(), false); lib != nil {
			lib.Version = libRelease.GetVersion()
			addedLib = lib
		} else {
			addedLib = &sketch.ProfileLibraryReference{
				Library: reqIndexLib.GetName(),
				Version: libRelease.GetVersion(),
			}
			profile.Libraries = append(profile.Libraries, addedLib)
		}
	} else {
		return nil, &cmderrors.InvalidArgumentError{Message: "library must be specified"}
	}

	err = projectFilePath.WriteFile([]byte(sk.Project.AsYaml()))
	if err != nil {
		return nil, err
	}

	return &rpc.ProfileLibAddResponse{Library: addedLib.ToRpc(), ProfileName: profileName}, nil
}
