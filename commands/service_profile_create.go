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
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

// ProfileCreate creates a new project file if it does not exist. If a profile name with the associated FQBN is specified,
// it is added to the project.
func (s *arduinoCoreServerImpl) ProfileCreate(ctx context.Context, req *rpc.ProfileCreateRequest) (*rpc.ProfileCreateResponse, error) {
	if req.GetProfileName() == "" {
		return nil, &cmderrors.MissingProfileError{}
	}
	if req.GetFqbn() == "" {
		return nil, &cmderrors.MissingFQBNError{}
	}

	// Returns an error if the main file is missing from the sketch so there is no need to check if the path exists
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, err
	}

	fqbn, err := fqbn.Parse(req.GetFqbn())
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	// Check that the profile name is unique
	if profile, _ := sk.GetProfile(req.ProfileName); profile != nil {
		return nil, &cmderrors.ProfileAlreadyExitsError{Profile: req.ProfileName}
	}

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()
	if pme.Dirty() {
		return nil, &cmderrors.InstanceNeedsReinitialization{}
	}

	// Automatically detect the target platform if it is installed on the user's machine
	_, targetPlatform, _, _, _, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		if targetPlatform == nil {
			return nil, &cmderrors.PlatformNotFoundError{
				Platform: fmt.Sprintf("%s:%s", fqbn.Vendor, fqbn.Architecture),
				Cause:    errors.New(i18n.Tr("platform not installed")),
			}
		}
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	newProfile := &sketch.Profile{Name: req.GetProfileName(), FQBN: req.GetFqbn()}
	// TODO: what to do with the PlatformIndexURL?
	newProfile.Platforms = append(newProfile.Platforms, &sketch.ProfilePlatformReference{
		Packager:     targetPlatform.Platform.Package.Name,
		Architecture: targetPlatform.Platform.Architecture,
		Version:      targetPlatform.Version,
	})

	sk.Project.Profiles = append(sk.Project.Profiles, newProfile)
	if req.DefaultProfile {
		sk.Project.DefaultProfile = newProfile.Name
	}

	projectFilePath := sk.GetProjectPath()
	err = projectFilePath.WriteFile([]byte(sk.Project.AsYaml()))
	if err != nil {
		return nil, err
	}

	return &rpc.ProfileCreateResponse{}, nil
}
