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
func (s *ArduinoCoreServerImpl) BoardDetails(ctx context.Context, req *rpc.BoardDetailsReq) (*rpc.BoardDetailsResponse, error) {
	return board.Details(ctx, req)
}

// BoardList FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardList(ctx context.Context, req *rpc.BoardListReq) (*rpc.BoardListResponse, error) {
	ports, err := board.List(req.GetInstance().GetId())
	if err != nil {
		return nil, err
	}

	return &rpc.BoardListResponse{
		Ports: ports,
	}, nil
}

// BoardListAll FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListAll(ctx context.Context, req *rpc.BoardListAllReq) (*rpc.BoardListAllResponse, error) {
	return board.ListAll(ctx, req)
}

// BoardSearch exposes to the gRPC interface the board search command
func (s *ArduinoCoreServerImpl) BoardSearch(ctx context.Context, req *rpc.BoardSearchReq) (*rpc.BoardSearchResponse, error) {
	return board.Search(ctx, req)
}

// BoardListWatch FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListWatch(stream rpc.ArduinoCoreService_BoardListWatchServer) error {
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
func (s *ArduinoCoreServerImpl) BoardAttach(req *rpc.BoardAttachReq, stream rpc.ArduinoCoreService_BoardAttachServer) error {

	resp, err := board.Attach(stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.BoardAttachResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// Destroy FIXMEDOC
func (s *ArduinoCoreServerImpl) Destroy(ctx context.Context, req *rpc.DestroyReq) (*rpc.DestroyResponse, error) {
	return commands.Destroy(ctx, req)
}

// Rescan FIXMEDOC
func (s *ArduinoCoreServerImpl) Rescan(ctx context.Context, req *rpc.RescanReq) (*rpc.RescanResponse, error) {
	return commands.Rescan(req.GetInstance().GetId())
}

// UpdateIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateIndex(req *rpc.UpdateIndexReq, stream rpc.ArduinoCoreService_UpdateIndexServer) error {
	resp, err := commands.UpdateIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateIndexResponse{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// UpdateLibrariesIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateLibrariesIndex(req *rpc.UpdateLibrariesIndexReq, stream rpc.ArduinoCoreService_UpdateLibrariesIndexServer) error {
	err := commands.UpdateLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateLibrariesIndexResponse{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpdateLibrariesIndexResponse{})
}

// UpdateCoreLibrariesIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateCoreLibrariesIndex(req *rpc.UpdateCoreLibrariesIndexReq, stream rpc.ArduinoCoreService_UpdateCoreLibrariesIndexServer) error {
	err := commands.UpdateCoreLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.UpdateCoreLibrariesIndexResponse{DownloadProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpdateCoreLibrariesIndexResponse{})
}

// Outdated FIXMEDOC
func (s *ArduinoCoreServerImpl) Outdated(ctx context.Context, req *rpc.OutdatedReq) (*rpc.OutdatedResponse, error) {
	return commands.Outdated(ctx, req)
}

// Upgrade FIXMEDOC
func (s *ArduinoCoreServerImpl) Upgrade(req *rpc.UpgradeReq, stream rpc.ArduinoCoreService_UpgradeServer) error {
	err := commands.Upgrade(stream.Context(), req,
		func(p *rpc.DownloadProgress) {
			stream.Send(&rpc.UpgradeResponse{
				Progress: p,
			})
		},
		func(p *rpc.TaskProgress) {
			stream.Send(&rpc.UpgradeResponse{
				TaskProgress: p,
			})
		},
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.UpgradeResponse{})
}

// Init FIXMEDOC
func (s *ArduinoCoreServerImpl) Init(req *rpc.InitReq, stream rpc.ArduinoCoreService_InitServer) error {
	resp, err := commands.Init(stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.InitResponse{DownloadProgress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.InitResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// Version FIXMEDOC
func (s *ArduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionReq) (*rpc.VersionResponse, error) {
	return &rpc.VersionResponse{Version: s.VersionString}, nil
}

// LoadSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) LoadSketch(ctx context.Context, req *rpc.LoadSketchReq) (*rpc.LoadSketchResponse, error) {
	return commands.LoadSketch(ctx, req)
}

// Compile FIXMEDOC
func (s *ArduinoCoreServerImpl) Compile(req *rpc.CompileReq, stream rpc.ArduinoCoreService_CompileServer) error {
	resp, err := compile.Compile(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.CompileResponse{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.CompileResponse{ErrStream: data}) }),
		false) // Set debug to false
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallReq, stream rpc.ArduinoCoreService_PlatformInstallServer) error {
	resp, err := core.PlatformInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformDownload FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadReq, stream rpc.ArduinoCoreService_PlatformDownloadServer) error {
	resp, err := core.PlatformDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformUninstall FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallReq, stream rpc.ArduinoCoreService_PlatformUninstallServer) error {
	resp, err := core.PlatformUninstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformUninstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformUpgrade FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeReq, stream rpc.ArduinoCoreService_PlatformUpgradeServer) error {
	resp, err := core.PlatformUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.PlatformUpgradeResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.PlatformUpgradeResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// PlatformSearch FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformSearch(ctx context.Context, req *rpc.PlatformSearchReq) (*rpc.PlatformSearchResponse, error) {
	return core.PlatformSearch(req)
}

// PlatformList FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformList(ctx context.Context, req *rpc.PlatformListReq) (*rpc.PlatformListResponse, error) {
	platforms, err := core.GetPlatforms(req)
	if err != nil {
		return nil, err
	}
	return &rpc.PlatformListResponse{InstalledPlatform: platforms}, nil
}

// Upload FIXMEDOC
func (s *ArduinoCoreServerImpl) Upload(req *rpc.UploadReq, stream rpc.ArduinoCoreService_UploadServer) error {
	resp, err := upload.Upload(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadResponse{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadResponse{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// UploadUsingProgrammer FIXMEDOC
func (s *ArduinoCoreServerImpl) UploadUsingProgrammer(req *rpc.UploadUsingProgrammerReq, stream rpc.ArduinoCoreService_UploadUsingProgrammerServer) error {
	resp, err := upload.UsingProgrammer(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadUsingProgrammerResponse{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.UploadUsingProgrammerResponse{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// BurnBootloader FIXMEDOC
func (s *ArduinoCoreServerImpl) BurnBootloader(req *rpc.BurnBootloaderReq, stream rpc.ArduinoCoreService_BurnBootloaderServer) error {
	resp, err := upload.BurnBootloader(
		stream.Context(), req,
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.BurnBootloaderResponse{OutStream: data}) }),
		utils.FeedStreamTo(func(data []byte) { stream.Send(&rpc.BurnBootloaderResponse{ErrStream: data}) }),
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *ArduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadReq) (*rpc.ListProgrammersAvailableForUploadResponse, error) {
	return upload.ListProgrammersAvailableForUpload(ctx, req)
}

// LibraryDownload FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryDownload(req *rpc.LibraryDownloadReq, stream rpc.ArduinoCoreService_LibraryDownloadServer) error {
	resp, err := lib.LibraryDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// LibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryInstall(req *rpc.LibraryInstallReq, stream rpc.ArduinoCoreService_LibraryInstallServer) error {
	err := lib.LibraryInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryInstallResponse{})
}

// LibraryUninstall FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUninstall(req *rpc.LibraryUninstallReq, stream rpc.ArduinoCoreService_LibraryUninstallServer) error {
	err := lib.LibraryUninstall(stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUninstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUninstallResponse{})
}

// LibraryUpgradeAll FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllReq, stream rpc.ArduinoCoreService_LibraryUpgradeAllServer) error {
	err := lib.LibraryUpgradeAll(req.GetInstance().GetId(),
		func(p *rpc.DownloadProgress) { stream.Send(&rpc.LibraryUpgradeAllResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { stream.Send(&rpc.LibraryUpgradeAllResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.LibraryUpgradeAllResponse{})
}

// LibraryResolveDependencies FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesReq) (*rpc.LibraryResolveDependenciesResponse, error) {
	return lib.LibraryResolveDependencies(ctx, req)
}

// LibrarySearch FIXMEDOC
func (s *ArduinoCoreServerImpl) LibrarySearch(ctx context.Context, req *rpc.LibrarySearchReq) (*rpc.LibrarySearchResponse, error) {
	return lib.LibrarySearch(ctx, req)
}

// LibraryList FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryList(ctx context.Context, req *rpc.LibraryListReq) (*rpc.LibraryListResponse, error) {
	return lib.LibraryList(ctx, req)
}

// ArchiveSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchReq) (*rpc.ArchiveSketchResponse, error) {
	return sketch.ArchiveSketch(ctx, req)
}

//ZipLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) ZipLibraryInstall(req *rpc.ZipLibraryInstallReq, stream rpc.ArduinoCoreService_ZipLibraryInstallServer) error {
	err := lib.ZipLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.ZipLibraryInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.ZipLibraryInstallResponse{})
}

//GitLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) GitLibraryInstall(req *rpc.GitLibraryInstallReq, stream rpc.ArduinoCoreService_GitLibraryInstallServer) error {
	err := lib.GitLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { stream.Send(&rpc.GitLibraryInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return stream.Send(&rpc.GitLibraryInstallResponse{})
}
