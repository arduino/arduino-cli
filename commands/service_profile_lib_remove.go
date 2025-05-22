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
)

func (s *arduinoCoreServerImpl) ProfileLibRemove(ctx context.Context, req *rpc.ProfileLibRemoveRequest) (*rpc.ProfileLibRemoveResponse, error) {
	sketchPath := paths.New(req.GetSketchPath())
	projectFilePath, err := sketchPath.Join("sketch.yaml").Abs()
	if err != nil {
		return nil, err
	}

	if req.GetProfileName() == "" {
		return nil, &cmderrors.MissingProfileError{}
	}

	// Returns an error if the main file is missing from the sketch so there is no need to check if the path exists
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, err
	}

	profile, err := sk.GetProfile(req.ProfileName)
	if err != nil {
		return nil, err
	}

	lib, err := profile.GetLibrary(req.LibName, true)
	if err != nil {
		return nil, err
	}

	err = projectFilePath.WriteFile([]byte(sk.Project.AsYaml()))
	if err != nil {
		return nil, err
	}

	return &rpc.ProfileLibRemoveResponse{LibName: lib.Library, LibVersion: lib.Version.String()}, nil
}
