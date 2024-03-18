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

	"github.com/arduino/arduino-cli/commands/cache"
	"github.com/arduino/arduino-cli/commands/updatecheck"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// NewArduinoCoreServer returns an implementation of the ArduinoCoreService gRPC service
// that uses the provided version string.
func NewArduinoCoreServer(version string) rpc.ArduinoCoreServiceServer {
	return &arduinoCoreServerImpl{
		versionString: version,
	}
}

type arduinoCoreServerImpl struct {
	rpc.UnsafeArduinoCoreServiceServer // Force compile error for unimplemented methods

	versionString string
}

// Version FIXMEDOC
func (s *arduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionRequest) (*rpc.VersionResponse, error) {
	return &rpc.VersionResponse{Version: s.versionString}, nil
}

// SupportedUserFields FIXMEDOC
func (s *arduinoCoreServerImpl) SupportedUserFields(ctx context.Context, req *rpc.SupportedUserFieldsRequest) (*rpc.SupportedUserFieldsResponse, error) {
	return SupportedUserFields(ctx, req)
}

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *arduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadRequest) (*rpc.ListProgrammersAvailableForUploadResponse, error) {
	return ListProgrammersAvailableForUpload(ctx, req)
}

// EnumerateMonitorPortSettings FIXMEDOC
func (s *arduinoCoreServerImpl) EnumerateMonitorPortSettings(ctx context.Context, req *rpc.EnumerateMonitorPortSettingsRequest) (*rpc.EnumerateMonitorPortSettingsResponse, error) {
	return EnumerateMonitorPortSettings(ctx, req)
}

// CheckForArduinoCLIUpdates FIXMEDOC
func (s *arduinoCoreServerImpl) CheckForArduinoCLIUpdates(ctx context.Context, req *rpc.CheckForArduinoCLIUpdatesRequest) (*rpc.CheckForArduinoCLIUpdatesResponse, error) {
	return updatecheck.CheckForArduinoCLIUpdates(ctx, req)
}

// CleanDownloadCacheDirectory FIXMEDOC
func (s *arduinoCoreServerImpl) CleanDownloadCacheDirectory(ctx context.Context, req *rpc.CleanDownloadCacheDirectoryRequest) (*rpc.CleanDownloadCacheDirectoryResponse, error) {
	return cache.CleanDownloadCacheDirectory(ctx, req)
}
