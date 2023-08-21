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
	"errors"
	"fmt"
	"io"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/monitor"
	"github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

// ArduinoCoreServerImpl FIXMEDOC
type ArduinoCoreServerImpl struct {
	// Force compile error for unimplemented methods
	rpc.UnsafeArduinoCoreServiceServer

	VersionString string
}

var tr = i18n.Tr

func convertErrorToRPCStatus(err error) error {
	if err == nil {
		return nil
	}
	if cmdErr, ok := err.(cmderrors.CommandError); ok {
		return cmdErr.ToRPCStatus().Err()
	}
	return err
}

// BoardDetails FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardDetails(ctx context.Context, req *rpc.BoardDetailsRequest) (*rpc.BoardDetailsResponse, error) {
	resp, err := board.Details(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// BoardList FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardList(ctx context.Context, req *rpc.BoardListRequest) (*rpc.BoardListResponse, error) {
	ports, _, err := board.List(req)
	if err != nil {
		return nil, convertErrorToRPCStatus(err)
	}
	return &rpc.BoardListResponse{
		Ports: ports,
	}, nil
}

// BoardListAll FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListAll(ctx context.Context, req *rpc.BoardListAllRequest) (*rpc.BoardListAllResponse, error) {
	resp, err := board.ListAll(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// BoardSearch exposes to the gRPC interface the board search command
func (s *ArduinoCoreServerImpl) BoardSearch(ctx context.Context, req *rpc.BoardSearchRequest) (*rpc.BoardSearchResponse, error) {
	resp, err := board.Search(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// BoardListWatch FIXMEDOC
func (s *ArduinoCoreServerImpl) BoardListWatch(req *rpc.BoardListWatchRequest, stream rpc.ArduinoCoreService_BoardListWatchServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	if req.GetInstance() == nil {
		err := fmt.Errorf(tr("no instance specified"))
		syncSend.Send(&rpc.BoardListWatchResponse{
			EventType: "error",
			Error:     err.Error(),
		})
		return err
	}

	eventsChan, err := board.Watch(stream.Context(), req)
	if err != nil {
		return convertErrorToRPCStatus(err)
	}

	for event := range eventsChan {
		if err := syncSend.Send(event); err != nil {
			logrus.Infof("sending board watch message: %v", err)
		}
	}

	return nil
}

// Destroy FIXMEDOC
func (s *ArduinoCoreServerImpl) Destroy(ctx context.Context, req *rpc.DestroyRequest) (*rpc.DestroyResponse, error) {
	resp, err := commands.Destroy(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// UpdateIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateIndex(req *rpc.UpdateIndexRequest, stream rpc.ArduinoCoreService_UpdateIndexServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := commands.UpdateIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.UpdateIndexResponse{DownloadProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// UpdateLibrariesIndex FIXMEDOC
func (s *ArduinoCoreServerImpl) UpdateLibrariesIndex(req *rpc.UpdateLibrariesIndexRequest, stream rpc.ArduinoCoreService_UpdateLibrariesIndexServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := commands.UpdateLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.UpdateLibrariesIndexResponse{DownloadProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// Create FIXMEDOC
func (s *ArduinoCoreServerImpl) Create(ctx context.Context, req *rpc.CreateRequest) (*rpc.CreateResponse, error) {
	var userAgent []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userAgent = md.Get("user-agent")
	}
	if len(userAgent) == 0 {
		userAgent = []string{"gRPCClientUnknown/0.0.0"}
	}
	res, err := commands.Create(req, userAgent...)
	return res, convertErrorToRPCStatus(err)
}

// Init FIXMEDOC
func (s *ArduinoCoreServerImpl) Init(req *rpc.InitRequest, stream rpc.ArduinoCoreService_InitServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := commands.Init(req, func(message *rpc.InitResponse) { syncSend.Send(message) })
	return convertErrorToRPCStatus(err)
}

// Version FIXMEDOC
func (s *ArduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionRequest) (*rpc.VersionResponse, error) {
	return &rpc.VersionResponse{Version: s.VersionString}, nil
}

// NewSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	resp, err := sketch.NewSketch(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// LoadSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) LoadSketch(ctx context.Context, req *rpc.LoadSketchRequest) (*rpc.LoadSketchResponse, error) {
	resp, err := sketch.LoadSketch(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// SetSketchDefaults FIXMEDOC
func (s *ArduinoCoreServerImpl) SetSketchDefaults(ctx context.Context, req *rpc.SetSketchDefaultsRequest) (*rpc.SetSketchDefaultsResponse, error) {
	resp, err := sketch.SetSketchDefaults(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// Compile FIXMEDOC
func (s *ArduinoCoreServerImpl) Compile(req *rpc.CompileRequest, stream rpc.ArduinoCoreService_CompileServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_OutStream{OutStream: data},
		})
	})
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_ErrStream{ErrStream: data},
		})
	})
	progressStream := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_Progress{Progress: p},
		})
	}
	compileRes, compileErr := compile.Compile(stream.Context(), req, outStream, errStream, progressStream)
	outStream.Close()
	errStream.Close()
	var compileRespSendErr error
	if compileRes != nil {
		compileRespSendErr = syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_Result{
				Result: compileRes,
			},
		})
	}
	if compileErr != nil {
		return convertErrorToRPCStatus(compileErr)
	}
	return compileRespSendErr
}

// PlatformInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallRequest, stream rpc.ArduinoCoreService_PlatformInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := core.PlatformInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return syncSend.Send(resp)
}

// PlatformDownload FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadRequest, stream rpc.ArduinoCoreService_PlatformDownloadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := core.PlatformDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return syncSend.Send(resp)
}

// PlatformUninstall FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallRequest, stream rpc.ArduinoCoreService_PlatformUninstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := core.PlatformUninstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformUninstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return syncSend.Send(resp)
}

// PlatformUpgrade FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeRequest, stream rpc.ArduinoCoreService_PlatformUpgradeServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := core.PlatformUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformUpgradeResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformUpgradeResponse{TaskProgress: p}) },
	)
	if err2 := syncSend.Send(resp); err2 != nil {
		return err2
	}
	return convertErrorToRPCStatus(err)
}

// PlatformSearch FIXMEDOC
func (s *ArduinoCoreServerImpl) PlatformSearch(ctx context.Context, req *rpc.PlatformSearchRequest) (*rpc.PlatformSearchResponse, error) {
	resp, err := core.PlatformSearch(req)
	return resp, convertErrorToRPCStatus(err)
}

// Upload FIXMEDOC
func (s *ArduinoCoreServerImpl) Upload(req *rpc.UploadRequest, stream rpc.ArduinoCoreService_UploadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_OutStream{OutStream: data},
		})
	})
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_ErrStream{ErrStream: data},
		})
	})
	res, err := upload.Upload(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if res != nil {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_Result{
				Result: res,
			},
		})
	}
	return convertErrorToRPCStatus(err)
}

// UploadUsingProgrammer FIXMEDOC
func (s *ArduinoCoreServerImpl) UploadUsingProgrammer(req *rpc.UploadUsingProgrammerRequest, stream rpc.ArduinoCoreService_UploadUsingProgrammerServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadUsingProgrammerResponse{
			Message: &rpc.UploadUsingProgrammerResponse_OutStream{
				OutStream: data,
			},
		})
	})
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadUsingProgrammerResponse{
			Message: &rpc.UploadUsingProgrammerResponse_ErrStream{
				ErrStream: data,
			},
		})
	})
	err := upload.UsingProgrammer(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return nil
}

// SupportedUserFields FIXMEDOC
func (s *ArduinoCoreServerImpl) SupportedUserFields(ctx context.Context, req *rpc.SupportedUserFieldsRequest) (*rpc.SupportedUserFieldsResponse, error) {
	res, err := upload.SupportedUserFields(ctx, req)
	return res, convertErrorToRPCStatus(err)
}

// BurnBootloader FIXMEDOC
func (s *ArduinoCoreServerImpl) BurnBootloader(req *rpc.BurnBootloaderRequest, stream rpc.ArduinoCoreService_BurnBootloaderServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.BurnBootloaderResponse{
			Message: &rpc.BurnBootloaderResponse_OutStream{
				OutStream: data,
			},
		})
	})
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.BurnBootloaderResponse{
			Message: &rpc.BurnBootloaderResponse_ErrStream{
				ErrStream: data,
			},
		})
	})
	resp, err := upload.BurnBootloader(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return syncSend.Send(resp)
}

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *ArduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadRequest) (*rpc.ListProgrammersAvailableForUploadResponse, error) {
	resp, err := upload.ListProgrammersAvailableForUpload(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// LibraryDownload FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryDownload(req *rpc.LibraryDownloadRequest, stream rpc.ArduinoCoreService_LibraryDownloadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := lib.LibraryDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return convertErrorToRPCStatus(err)
	}
	return syncSend.Send(resp)
}

// LibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryInstall(req *rpc.LibraryInstallRequest, stream rpc.ArduinoCoreService_LibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.LibraryInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryInstallResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// LibraryUpgrade FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUpgrade(req *rpc.LibraryUpgradeRequest, stream rpc.ArduinoCoreService_LibraryUpgradeServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.LibraryUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// LibraryUninstall FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUninstall(req *rpc.LibraryUninstallRequest, stream rpc.ArduinoCoreService_LibraryUninstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.LibraryUninstall(stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUninstallResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// LibraryUpgradeAll FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllRequest, stream rpc.ArduinoCoreService_LibraryUpgradeAllServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.LibraryUpgradeAll(req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// LibraryResolveDependencies FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesRequest) (*rpc.LibraryResolveDependenciesResponse, error) {
	resp, err := lib.LibraryResolveDependencies(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// LibrarySearch FIXMEDOC
func (s *ArduinoCoreServerImpl) LibrarySearch(ctx context.Context, req *rpc.LibrarySearchRequest) (*rpc.LibrarySearchResponse, error) {
	resp, err := lib.LibrarySearch(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// LibraryList FIXMEDOC
func (s *ArduinoCoreServerImpl) LibraryList(ctx context.Context, req *rpc.LibraryListRequest) (*rpc.LibraryListResponse, error) {
	resp, err := lib.LibraryList(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// ArchiveSketch FIXMEDOC
func (s *ArduinoCoreServerImpl) ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchRequest) (*rpc.ArchiveSketchResponse, error) {
	resp, err := sketch.ArchiveSketch(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// ZipLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) ZipLibraryInstall(req *rpc.ZipLibraryInstallRequest, stream rpc.ArduinoCoreService_ZipLibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.ZipLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.ZipLibraryInstallResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// GitLibraryInstall FIXMEDOC
func (s *ArduinoCoreServerImpl) GitLibraryInstall(req *rpc.GitLibraryInstallRequest, stream rpc.ArduinoCoreService_GitLibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	err := lib.GitLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.GitLibraryInstallResponse{TaskProgress: p}) },
	)
	return convertErrorToRPCStatus(err)
}

// EnumerateMonitorPortSettings FIXMEDOC
func (s *ArduinoCoreServerImpl) EnumerateMonitorPortSettings(ctx context.Context, req *rpc.EnumerateMonitorPortSettingsRequest) (*rpc.EnumerateMonitorPortSettingsResponse, error) {
	resp, err := monitor.EnumerateMonitorPortSettings(ctx, req)
	return resp, convertErrorToRPCStatus(err)
}

// Monitor FIXMEDOC
func (s *ArduinoCoreServerImpl) Monitor(stream rpc.ArduinoCoreService_MonitorServer) error {
	syncSend := NewSynchronizedSend(stream.Send)

	// The configuration must be sent on the first message
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	openReq := req.GetOpenRequest()
	if openReq == nil {
		return &cmderrors.InvalidInstanceError{}
	}
	portProxy, _, err := monitor.Monitor(stream.Context(), openReq)
	if err != nil {
		return err
	}

	// Send a message with Success set to true to notify the caller of the port being now active
	_ = syncSend.Send(&rpc.MonitorResponse{Success: true})

	cancelCtx, cancel := context.WithCancel(stream.Context())
	go func() {
		defer cancel()
		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				syncSend.Send(&rpc.MonitorResponse{Error: err.Error()})
				return
			}
			if conf := msg.GetUpdatedConfiguration(); conf != nil {
				for _, c := range conf.GetSettings() {
					if err := portProxy.Config(c.GetSettingId(), c.GetValue()); err != nil {
						syncSend.Send(&rpc.MonitorResponse{Error: err.Error()})
					}
				}
			}
			if close := msg.GetClose(); close {
				if err := portProxy.Close(); err != nil {
					logrus.WithError(err).Debug("Error closing monitor port")
				}
			}
			tx := msg.GetTxData()
			for len(tx) > 0 {
				n, err := portProxy.Write(tx)
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					syncSend.Send(&rpc.MonitorResponse{Error: err.Error()})
					return
				}
				tx = tx[n:]
			}
		}
	}()

	go func() {
		defer cancel()
		buff := make([]byte, 4096)
		for {
			n, err := portProxy.Read(buff)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				syncSend.Send(&rpc.MonitorResponse{Error: err.Error()})
				break
			}
			if err := syncSend.Send(&rpc.MonitorResponse{RxData: buff[:n]}); err != nil {
				break
			}
		}
	}()

	<-cancelCtx.Done()
	portProxy.Close()
	return nil
}
