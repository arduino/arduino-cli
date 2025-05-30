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
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

// ProfileDump dumps the content of the project file.
func (s *arduinoCoreServerImpl) ProfileDump(ctx context.Context, req *rpc.ProfileDumpRequest) (*rpc.ProfileDumpResponse, error) {
	sk, err := sketch.New(paths.New(req.GetSketchPath()))
	if err != nil {
		return nil, err
	}
	switch req.GetDumpFormat() {
	case "yaml":
		return &rpc.ProfileDumpResponse{EncodedProfile: sk.Project.AsYaml()}, nil
	case "", "json":
		data, err := json.MarshalIndent(sk.Project, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling settings: %v", err)
		}
		return &rpc.ProfileDumpResponse{EncodedProfile: string(data)}, nil
	default:
		return nil, &cmderrors.InvalidArgumentError{Message: fmt.Sprintf("unsupported format: %s", req.GetDumpFormat())}
	}
}
