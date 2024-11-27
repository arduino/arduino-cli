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

package commands

import (
	"context"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *arduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadRequest) (*rpc.ListProgrammersAvailableForUploadResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" {
		return nil, &cmderrors.MissingFQBNError{}
	}
	fqbn, err := fqbn.Parse(fqbnIn)
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	// Find target platforms
	_, platform, _, _, refPlatform, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		return nil, &cmderrors.UnknownFQBNError{Cause: err}
	}

	result := []*rpc.Programmer{}
	createRPCProgrammer := func(id string, programmer *cores.Programmer) *rpc.Programmer {
		return &rpc.Programmer{
			Id:       id,
			Platform: programmer.PlatformRelease.String(),
			Name:     programmer.Name,
		}
	}
	if refPlatform != platform {
		for id, programmer := range refPlatform.Programmers {
			result = append(result, createRPCProgrammer(id, programmer))
		}
	}
	for id, programmer := range platform.Programmers {
		result = append(result, createRPCProgrammer(id, programmer))
	}

	return &rpc.ListProgrammersAvailableForUploadResponse{
		Programmers: result,
	}, nil
}
