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

package upload

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// ListProgrammersAvailableForUpload FIXMEDOC
func ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadReq) (*rpc.ListProgrammersAvailableForUploadResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" {
		return nil, fmt.Errorf("no Fully Qualified Board Name provided")
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, fmt.Errorf("incorrect FQBN: %s", err)
	}

	// Find target platforms
	_, platform, _, _, refPlatform, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, fmt.Errorf("incorrect FQBN: %s", err)
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

	return &rpc.ListProgrammersAvailableForUploadResp{
		Programmers: result,
	}, nil
}
