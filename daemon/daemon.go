//

package daemon

//go:generate protoc -I arduino --go_out=plugins=grpc:arduino arduino/arduino.proto

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/rpc"
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
	rpc.RegisterArduinoCoreServer(s, &ArduinoCoreServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("Done serving")
}

type ArduinoCoreServerImpl struct{}

func (s *ArduinoCoreServerImpl) BoardDetails(ctx context.Context, req *rpc.BoardDetailsReq) (*rpc.BoardDetailsResp, error) {
	return board.BoardDetails(ctx, req)
}

func (s *ArduinoCoreServerImpl) Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResp, error) {
	return commands.Destroy(ctx, req)
}

func (s *ArduinoCoreServerImpl) Init(ctx context.Context, req *rpc.InitReq) (*rpc.InitResp, error) {
	return commands.Init(ctx, req)
}

func (s *ArduinoCoreServerImpl) Compile(req *rpc.CompileReq, stream rpc.ArduinoCore_CompileServer) error {
	r, w := io.Pipe()
	go func() {
		data := make([]byte, 1024)
		for {
			if n, err := r.Read(data); err != nil {
				return
			} else {
				stream.Send(&rpc.CompileResp{Output: data[:n]})
			}
		}
	}()
	resp, err := compile.Compile(stream.Context(), req, w)
	stream.Send(resp)
	return err
}

func (s *ArduinoCoreServerImpl) PlatformInstall(ctx context.Context, req *rpc.PlatformInstallReq) (*rpc.PlatformInstallResp, error) {
	return core.PlatformInstall(ctx, req)
}

func (s *ArduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadReq, stream rpc.ArduinoCore_PlatformDownloadServer) error {
	resp, err := core.PlatformDownload(stream.Context(), req, func(curr *rpc.DownloadProgress) {
		stream.Send(&rpc.PlatformDownloadResp{Progress: curr})
	})
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformUninstall(ctx context.Context, req *rpc.PlatformUninstallReq) (*rpc.PlatformUninstallResp, error) {
	return core.PlatformUninstall(ctx, req)
}
