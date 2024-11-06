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

	"github.com/arduino/arduino-cli/internal/cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
)

// NewArduinoCoreServer returns an implementation of the ArduinoCoreService gRPC service
// that uses the provided version string.
func NewArduinoCoreServer() rpc.ArduinoCoreServiceServer {
	return &arduinoCoreServerImpl{settings: configuration.NewSettings()}
}

type arduinoCoreServerImpl struct {
	rpc.UnsafeArduinoCoreServiceServer // Force compile error for unimplemented methods

	// Settings holds configurations of the CLI and the gRPC consumers
	settings *configuration.Settings
}

// Version returns the version of the Arduino CLI
func (s *arduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionRequest) (*rpc.VersionResponse, error) {
	return &rpc.VersionResponse{Version: version.VersionInfo.VersionString}, nil
}
