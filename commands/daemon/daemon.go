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

//go:generate protoc -I arduino --go_out=plugins=grpc:arduino arduino/arduino.proto

import (
	"context"
	"io"

	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// ArduinoCoreServerImpl FIXMEDOC
type ArduinoCoreServerImpl struct {
	VersionString string
}

// BoardDetails FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardDetails(ctx context.Context, req *rpc.BoardDetailsReq) (*rpc.BoardDetailsResp, error) {
	return board.Details(ctx, req)
}

// BoardList FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardList(ctx context.Context, req *rpc.BoardListReq) (*rpc.BoardListResp, error) {
	ports, err := board.List(req.GetInstance().GetId())
	if err != nil {
		return nil, err
	}

	return &rpc.BoardListResp{
		Ports: ports,
	}, nil
}

// BoardListAll FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListAll(ctx context.Context, req *rpc.BoardListAllReq) (*rpc.BoardListAllResp, error) {
	return board.ListAll(ctx, req)
}

// BoardSearch exposes to the gRPC interface the board search command
func (s *ArduinoCoreServerImpl) BoardSearch(ctx context.Context, req *rpc.BoardSearchReq) (*rpc.BoardSearchResp, error) {
	return board.Search(ctx, req)
}

// BoardListWatch FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListWatch(stream rpc.ArduinoCore_BoardListWatchServer) error {
	msg, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	interrupt := make(chan bool)
	go func() {
		msg, err := stream.Recv()
		if err != nil {
			interrupt <- true
		}
		if msg != nil {
			interrupt <- msg.Interrupt
		}
	}()

	eventsChan, err := board.Watch(msg.Instance.Id, interrupt)
	if err != nil {
		return err
	}

	for event := range eventsChan {
		stream.Send(event)
	}

	return nil
}

// BoardAttach FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardAttach(req *rpc.BoardAttachReq, stream rpc.ArduinoCore_BoardAttachServer) error {

	resp, err := board.Attach(stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.BoardAttachResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// Destroy FIXMEDOC
func (s *ArduinoCoreServerImpl) Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResp, error) {
	return commands.Destroy(ctx, req)
}

// Rescan FIXMEDOC
func (s *ArduinoCoreServerImpl) Rescan(ctx context.Context, req *rpc.RescanReq) (*rpc.RescanResp, error) {
	return commands.Rescan(req.GetInstance().GetId())
}

// UpdateIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateIndex(req *rpc.UpdateIndexReq, stream rpc.ArduinoCore_UpdateIndexServer) error {
	resp, err := commands.UpdateIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateIndexResp{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// UpdateLibrariesIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateLibrariesIndex(req *rpc.UpdateLibrariesIndexReq, stream rpc.ArduinoCore_UpdateLibrariesIndexServer) error {
	err := commands.UpdateLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateLibrariesIndexResp{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpdateLibrariesIndexResp{})
}

// UpdateCoreLibrariesIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateCoreLibrariesIndex(req *rpc.UpdateCoreLibrariesIndexReq, stream rpc.ArduinoCore_UpdateCoreLibrariesIndexServer) error {
	err := commands.UpdateCoreLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateCoreLibrariesIndexResp{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpdateCoreLibrariesIndexResp{})
}

// Outdated FIXMEDOC
func (s *ArduinoCoreServerImpl) Outdated(ctx context.Context, req *rpc.OutdatedReq) (*rpc.OutdatedResp, error) {
	return commands.Outdated(ctx, req)
}

// Upgrade FIXMEDOC
func (s *ArduinoCoreServerImpl) Upgrade(req *rpc.UpgradeReq, stream rpc.ArduinoCore_UpgradeServer) error {
	err := commands.Upgrade(stream.Context(), req,
		func(p *rpc.DownloadProgress) {
			stream.Send(&rpc.UpgradeResp{
				Progress: p,
			})
		},
		func(p *rpc.TaskProgress) {
			stream.Send(&rpc.UpgradeResp{
				TaskProgress: p,
			})
		},
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpgradeResp{})
}

// Init FIXMEDOC
func (s *ArduinoCoreServerImpl) Init(req *rpc.InitReq, stream rpc.ArduinoCore_InitServer) error {
	resp, err := commands.Init(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.InitResp{DownloadProgress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.InitResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// Version FIXMEDOC
func (s *ArduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionReq) (*rpc.VersionResp, error) {
	return &rpc.VersionResp{Version: s.VersionString}, nil
}

// LoadSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) LoadSketch(ctx context.Context, req *rpc.LoadSketchReq) (*rpc.LoadSketchResp, error) {
	return commands.LoadSketch(ctx, req)
}

// Compile FIXMEDOC
func (s *ArduinoCoreServerImpl) Compile(req *rpc.CompileReq, stream rpc.ArduinoCore_CompileServer) error {
	resp, err := compile.Compile(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.CompileResp{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.CompileResp{ErrStream: data}) }),
		false) // Set debug to false
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformInstall FIXMEDOC
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

// PlatformDownload FIXMEDOC
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

// PlatformUninstall FIXMEDOC
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

// PlatformUpgrade FIXMEDOC
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

// PlatformSearch FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformSearch(ctx context.Context, req *rpc.PlatformSearchReq) (*rpc.PlatformSearchResp, error) {
	return core.PlatformSearch(req)
}

// PlatformList FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformList(ctx context.Context, req *rpc.PlatformListReq) (*rpc.PlatformListResp, error) {
	platforms, err := core.GetPlatforms(req)
	if err != nil {
		return nil, err
	}
	return &rpc.PlatformListResp{InstalledPlatform: platforms}, nil
}

// Upload FIXMEDOC
func (s *ArduinoCoreServerImpl) Upload(req *rpc.UploadReq, stream rpc.ArduinoCore_UploadServer) error {
	resp, err := upload.Upload(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadResp{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadResp{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// UploadUsingProgrammer FIXMEDOC
func (s *ArduinoCoreServerImpl) UploadUsingProgrammer(req *rpc.UploadUsingProgrammerReq, stream rpc.ArduinoCore_UploadUsingProgrammerServer) error {
	resp, err := upload.UsingProgrammer(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadUsingProgrammerResp{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadUsingProgrammerResp{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// BurnBootloader FIXMEDOC
func (s *ArduinoCoreServerImpl) BurnBootloader(req *rpc.BurnBootloaderReq, stream rpc.ArduinoCore_BurnBootloaderServer) error {
	resp, err := upload.BurnBootloader(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.BurnBootloaderResp{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.BurnBootloaderResp{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *ArduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadReq) (*rpc.ListProgrammersAvailableForUploadResp, error) {
	return upload.ListProgrammersAvailableForUpload(ctx, req)
}

// LibraryDownload FIXMEDOC
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

// LibraryInstall FIXMEDOC
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

// LibraryUninstall FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUninstall(req *rpc.LibraryUninstallReq, stream rpc.ArduinoCore_LibraryUninstallServer) error {
	err := lib.LibraryUninstall(stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUninstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUninstallResp{})
}

// LibraryUpgradeAll FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllReq, stream rpc.ArduinoCore_LibraryUpgradeAllServer) error {
	err := lib.LibraryUpgradeAll(req.GetInstance().GetId(),
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryUpgradeAllResp{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUpgradeAllResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUpgradeAllResp{})
}

// LibraryResolveDependencies FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesReq) (*rpc.LibraryResolveDependenciesResp, error) {
	return lib.LibraryResolveDependencies(ctx, req)
}

// LibrarySearch FIXMEDOC
func (s *ArduinoCoreServerImpl) LibrarySearch(ctx context.Context, req *rpc.LibrarySearchReq) (*rpc.LibrarySearchResp, error) {
	return lib.LibrarySearch(ctx, req)
}

// LibraryList FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryList(ctx context.Context, req *rpc.LibraryListReq) (*rpc.LibraryListResp, error) {
	return lib.LibraryList(ctx, req)
}

// ArchiveSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchReq) (*rpc.ArchiveSketchResp, error) {
	return sketch.ArchiveSketch(ctx, req)
}

//ZipLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) ZipLibraryInstall(req *rpc.ZipLibraryInstallReq, stream rpc.ArduinoCore_ZipLibraryInstallServer) error {
	err := lib.ZipLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.ZipLibraryInstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.ZipLibraryInstallResp{})
}

//GitLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) GitLibraryInstall(req *rpc.GitLibraryInstallReq, stream rpc.ArduinoCore_GitLibraryInstallServer) error {
	err := lib.GitLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.GitLibraryInstallResp{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.GitLibraryInstallResp{})
}
