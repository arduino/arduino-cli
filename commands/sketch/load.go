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

package sketch

import (
	"context"

	"github.com/arduino/arduino-cli/internal/arduino"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

// LoadSketch collects and returns all files composing a sketch
func LoadSketch(ctx context.Context, req *rpc.LoadSketchRequest) (*rpc.LoadSketchResponse, error) {
	// TODO: This should be a ToRpc function for the Sketch struct
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	otherSketchFiles := make([]string, sk.OtherSketchFiles.Len())
	for i, file := range sk.OtherSketchFiles {
		otherSketchFiles[i] = file.String()
	}

	additionalFiles := make([]string, sk.AdditionalFiles.Len())
	for i, file := range sk.AdditionalFiles {
		additionalFiles[i] = file.String()
	}

	rootFolderFiles := make([]string, sk.RootFolderFiles.Len())
	for i, file := range sk.RootFolderFiles {
		rootFolderFiles[i] = file.String()
	}

	defaultPort, defaultProtocol := sk.GetDefaultPortAddressAndProtocol()

	profiles := make([](*rpc.SketchProfile), len(sk.Project.Profiles))
	for i, profile := range sk.Project.Profiles {
		profiles[i] = &rpc.SketchProfile{
			Name: profile.Name,
			Fqbn: profile.FQBN,
		}
	}

	defaultProfileResp := &rpc.SketchProfile{}
	defaultProfile := sk.GetProfile(sk.Project.DefaultProfile)
	if defaultProfile != nil {
		defaultProfileResp.Name = defaultProfile.Name
		defaultProfileResp.Fqbn = defaultProfile.FQBN
	}

	return &rpc.LoadSketchResponse{
		MainFile:         sk.MainFile.String(),
		LocationPath:     sk.FullPath.String(),
		OtherSketchFiles: otherSketchFiles,
		AdditionalFiles:  additionalFiles,
		RootFolderFiles:  rootFolderFiles,
		DefaultFqbn:      sk.GetDefaultFQBN(),
		DefaultPort:      defaultPort,
		DefaultProtocol:  defaultProtocol,
		Profiles:         profiles,
		DefaultProfile:   defaultProfileResp,
	}, nil
}
