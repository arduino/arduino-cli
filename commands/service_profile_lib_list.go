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
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"go.bug.st/f"
)

// ProfileLibList lists the libraries in the build profile.
func (s *arduinoCoreServerImpl) ProfileLibList(ctx context.Context, req *rpc.ProfileLibListRequest) (*rpc.ProfileLibListResponse, error) {
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, err
	}

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
		return nil, &cmderrors.UnknownProfileError{Profile: profileName}
	}

	return &rpc.ProfileLibListResponse{
		Libraries:   f.Map(profile.Libraries, (*sketch.ProfileLibraryReference).ToRpc),
		ProfileName: profileName,
	}, nil
}
