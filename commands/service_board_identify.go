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

	"github.com/arduino/arduino-cli/commands/internal/instances"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
)

// BoardIdentify identifies the board based on the provided properties
func (s *arduinoCoreServerImpl) BoardIdentify(ctx context.Context, req *rpc.BoardIdentifyRequest) (*rpc.BoardIdentifyResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	props := properties.NewFromHashmap(req.GetProperties())
	res, err := identify(pme, props, s.settings, true)
	if err != nil {
		return nil, err
	}
	return &rpc.BoardIdentifyResponse{
		Boards: res,
	}, nil
}
