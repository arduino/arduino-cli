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
	"github.com/arduino/arduino-cli/commands/lib"
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

func (s *ArduinoCoreServerImpl) UpdateIndex(req *rpc.UpdateIndexReq, stream rpc.ArduinoCore_UpdateIndexServer) error {
	resp, err := commands.UpdateIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateIndexResp{DownloadProgress: p}) },
	)
	stream.Send(resp)
	return err
}

func (s *ArduinoCoreServerImpl) Init(ctx context.Context, req *rpc.InitReq) (*rpc.InitResp, error) {
	return commands.Init(ctx, req)
}

func (s *ArduinoCoreServerImpl) Compile(req *rpc.CompileReq, stream rpc.ArduinoCore_CompileServer) error {
	resp, err := compile.Compile(
		stream.Context(), req,
		feedStream(func(data []byte) { stream.Send(&rpc.CompileResp{OutStream: data}) }),
		feedStream(func(data []byte) { stream.Send(&rpc.CompileResp{ErrStream: data}) }),
		func(p *rpc.TaskProgress) { stream.Send(&rpc.CompileResp{TaskProgress: p}) },
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.CompileResp{DownloadProgress: p}) },
	)
	stream.Send(resp)
	return err
}

func (s *ArduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallReq, stream rpc.ArduinoCore_PlatformInstallServer) error {
	resp, err := core.PlatformInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformInstallResp{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformInstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadReq, stream rpc.ArduinoCore_PlatformDownloadServer) error {
	resp, err := core.PlatformDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformDownloadResp{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallReq, stream rpc.ArduinoCore_PlatformUninstallServer) error {
	resp, err := core.PlatformUninstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformUninstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeReq, stream rpc.ArduinoCore_PlatformUpgradeServer) error {
	resp, err := core.PlatformUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformUpgradeResp{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformUpgradeResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) PlatformSearch(ctx context.Context, req *rpc.PlatformSearchReq) (*rpc.PlatformSearchResp, error) {
	return core.PlatformSearch(ctx, req)
}

func (s *ArduinoCoreServerImpl) PlatformList(ctx context.Context, req *rpc.PlatformListReq) (*rpc.PlatformListResp, error) {
	return core.PlatformList(ctx, req)
}

func (s *ArduinoCoreServerImpl) Upload(req *rpc.UploadReq, stream rpc.ArduinoCore_UploadServer) error {
	resp, err := upload.Upload(
		stream.Context(), req,
		feedStream(func(data []byte) { stream.Send(&rpc.UploadResp{OutStream: data}) }),
		feedStream(func(data []byte) { stream.Send(&rpc.UploadResp{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func feedStream(streamer func(data []byte)) io.Writer {
	r, w := io.Pipe()
	go func() {
		data := make([]byte, 1024)
		for {
			if n, err := r.Read(data); err != nil {
				return
			} else {
				streamer(data[:n])
			}
		}
	}()
	return w
}

func (s *ArduinoCoreServerImpl) LibraryDownload(req *rpc.LibraryDownloadReq, stream rpc.ArduinoCore_LibraryDownloadServer) error {
	resp, err := lib.LibraryDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryDownloadResp{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

func (s *ArduinoCoreServerImpl) LibraryInstall(req *rpc.LibraryInstallReq, stream rpc.ArduinoCore_LibraryInstallServer) error {
	err := lib.LibraryInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryInstallResp{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryInstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryInstallResp{})
}

func (s *ArduinoCoreServerImpl) LibraryUninstall(req *rpc.LibraryUninstallReq, stream rpc.ArduinoCore_LibraryUninstallServer) error {
	err := lib.LibraryUninstall(stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUninstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUninstallResp{})
}

func (s *ArduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllReq, stream rpc.ArduinoCore_LibraryUpgradeAllServer) error {
	err := lib.LibraryUpgradeAll(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryUpgradeAllResp{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUpgradeAllResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUpgradeAllResp{})
}

func (s *ArduinoCoreServerImpl) LibrarySearch(ctx context.Context, req *rpc.LibrarySearchReq) (*rpc.LibrarySearchResp, error) {
	return lib.LibrarySearch(ctx, req)
}
