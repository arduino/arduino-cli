//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package daemon

import (
	"fmt"
	"log"
	"net"

	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/proto-gen/commands/rpc_v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func runDaemonCommand(cmd *cobra.Command, args []string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	rpc_v1.RegisterArduinoCommandsServer(s, &ArduinoCommandsService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("Done serving")
}

// ArduinoCommandsService is the gRPC service implementation
type ArduinoCommandsService struct{}

// Version returns a message containing CLI version
// func (s *ArduinoCommandsService) Version(ctx context.Context, req *rpc.VersionReq) (*rpc.VersionResp, error) {
// 	return &rpc.VersionResp{Version: cli.Version}, nil
// }

// Compile sends a request for a compilation run
func (s *ArduinoCommandsService) Compile(req *rpc_v1.CompileReq, stream rpc_v1.ArduinoCommands_CompileServer) error {
	resp, err := compile.Compile(
		stream.Context(), req,
		feedStream(func(data []byte) { stream.Send(&rpc_v1.ExecResp{OutStream: data}) }),
		feedStream(func(data []byte) { stream.Send(&rpc_v1.ExecResp{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc_v1.ExecResp{ErrStream: resp.ErrStream, OutStream: resp.OutStream})
}
