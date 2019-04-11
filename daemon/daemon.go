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
	"github.com/arduino/arduino-cli/commands/upload"
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

func (s *ArduinoCoreServerImpl) Rescan(ctx context.Context, req *rpc.RescanReq) (*rpc.RescanResp, error) {
	return commands.Rescan(ctx, req)
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
	resp, err := compile.Compile(stream.Context(), req, w, func(taskProgress *rpc.TaskProgress) {
		stream.Send(&rpc.CompileResp{TaskProgress: taskProgress})
	}, func(downloadProgress *rpc.DownloadProgress) {
		stream.Send(&rpc.CompileResp{DownloadProgress: downloadProgress})
	})
	stream.Send(resp)
	return err
}

func (s *ArduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallReq, stream rpc.ArduinoCore_PlatformInstallServer) error {
	resp, err := core.PlatformInstall(stream.Context(), req, func(progress *rpc.DownloadProgress) {
		stream.Send(&rpc.PlatformInstallResp{Progress: progress})
	}, func(taskProgress *rpc.TaskProgress) {
		stream.Send(&rpc.PlatformInstallResp{TaskProgress: taskProgress})
	})
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadReq, stream rpc.ArduinoCore_PlatformDownloadServer) error {
	resp, err := core.PlatformDownload(stream.Context(), req, func(progress *rpc.DownloadProgress) {
		stream.Send(&rpc.PlatformDownloadResp{Progress: progress})
	})
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallReq, stream rpc.ArduinoCore_PlatformUninstallServer) error {
	resp, err := core.PlatformUninstall(stream.Context(), req, func(taskProgress *rpc.TaskProgress) {
		stream.Send(&rpc.PlatformUninstallResp{TaskProgress: taskProgress})
	})
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeReq, stream rpc.ArduinoCore_PlatformUpgradeServer) error {
	resp, err := core.PlatformUpgrade(stream.Context(), req, func(progress *rpc.DownloadProgress) {
		stream.Send(&rpc.PlatformUpgradeResp{Progress: progress})
	}, func(taskProgress *rpc.TaskProgress) {
		stream.Send(&rpc.PlatformUpgradeResp{TaskProgress: taskProgress})
	})
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) Upload(req *rpc.UploadReq, stream rpc.ArduinoCore_UploadServer) error {
	r, w := io.Pipe()
	r2, w2 := io.Pipe()

	feedStream(r, func(data []byte) { stream.Send(&rpc.UploadResp{OutStream: data}) })
	feedStream(r2, func(data []byte) { stream.Send(&rpc.UploadResp{ErrStream: data}) })

	resp, err := upload.Upload(stream.Context(), req, w, w2)
	stream.Send(resp)
	return err
}

func feedStream(out io.Reader, streamer func(data []byte)) {
	go func() {
		data := make([]byte, 1024)
		for {
			if n, err := out.Read(data); err != nil {
				return
			} else {
				streamer(data[:n])
			}
		}
	}()
}
