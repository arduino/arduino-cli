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

	"github.com/arduino/arduino-cli/arduino/utils"
	cmd "github.com/arduino/arduino-cli/commands/debug"
	dbg "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	"github.com/pkg/errors"
)

// DebugService implements the `Debug` service
type DebugService struct {
	dbg.UnimplementedDebugServiceServer
}

// Debug returns a stream response that can be used to fetch data from the
// target. The first message passed through the `Debug` request must
// contain DebugRequest configuration params, not data.
func (s *DebugService) Debug(stream dbg.DebugService_DebugServer) error {

	// Grab the first message
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// Ensure it's a config message and not data
	req := msg.GetDebugRequest()
	if req == nil {
		return errors.Errorf("First message must contain debug request, not data")
	}

	// Launch debug recipe attaching stdin and out to grpc streaming
	signalChan := make(chan os.Signal)
	defer close(signalChan)
	resp, err := cmd.Debug(stream.Context(), req,
		utils.ConsumeStreamFrom(func() ([]byte, error) {
			command, err := stream.Recv()
			if command.GetSendInterrupt() {
				signalChan <- os.Interrupt
			}
			return command.GetData(), err
		}),
		utils.FeedStreamTo(func(data []byte) {
			stream.Send(&dbg.DebugResponse{Data: data})
		}),
		signalChan)
	if err != nil {
		return (err)
	}
	return stream.Send(resp)
}

// GetDebugConfig return metadata about a debug session
func (s *DebugService) GetDebugConfig(ctx context.Context, req *dbg.DebugConfigRequest) (*dbg.GetDebugConfigResponse, error) {
	return cmd.GetDebugConfig(ctx, req)
}
