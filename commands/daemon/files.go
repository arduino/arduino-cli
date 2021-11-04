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

package daemon

import (
	"context"
	"os"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/files/v1"
)

// FilesService implements the `Files` service
type FilesService struct {
	rpc.UnimplementedFilesServiceServer
}

// LoadFile returns a requested file content or an error.
func (s *FilesService) LoadFile(ctx context.Context, req *rpc.LoadFileRequest) (*rpc.LoadFileResponse, error) {
	content, err := os.ReadFile(req.Path)
	// TODO: support req.Type other than RAW or remove Type from .proto
	if err == nil {
		return &rpc.LoadFileResponse{
			Content: content,
			Type:    rpc.ContentType_CONTENT_TYPE_RAW_UNSPECIFIED,
		}, nil
	}

	return nil, err
}
