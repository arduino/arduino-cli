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
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/arduino/arduino-cli/commands/cache"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/updatecheck"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
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

// BoardSearch exposes to the gRPC interface the board search command
func (s *arduinoCoreServerImpl) BoardSearch(ctx context.Context, req *rpc.BoardSearchRequest) (*rpc.BoardSearchResponse, error) {
	return BoardSearch(ctx, req)
}

// BoardListWatch FIXMEDOC
func (s *arduinoCoreServerImpl) BoardListWatch(req *rpc.BoardListWatchRequest, stream rpc.ArduinoCoreService_BoardListWatchServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	if req.GetInstance() == nil {
		err := fmt.Errorf(tr("no instance specified"))
		syncSend.Send(&rpc.BoardListWatchResponse{
			EventType: "error",
			Error:     err.Error(),
		})
		return err
	}

	eventsChan, err := BoardListWatch(stream.Context(), req)
	if err != nil {
		return err
	}

	for event := range eventsChan {
		if err := syncSend.Send(event); err != nil {
			logrus.Infof("sending board watch message: %v", err)
		}
	}

	return nil
}

// Destroy FIXMEDOC
func (s *arduinoCoreServerImpl) Destroy(ctx context.Context, req *rpc.DestroyRequest) (*rpc.DestroyResponse, error) {
	return Destroy(ctx, req)
}

// UpdateIndex FIXMEDOC
func (s *arduinoCoreServerImpl) UpdateIndex(req *rpc.UpdateIndexRequest, stream rpc.ArduinoCoreService_UpdateIndexServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	res, err := UpdateIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) {
			syncSend.Send(&rpc.UpdateIndexResponse{
				Message: &rpc.UpdateIndexResponse_DownloadProgress{DownloadProgress: p},
			})
		},
	)
	if res != nil {
		syncSend.Send(&rpc.UpdateIndexResponse{
			Message: &rpc.UpdateIndexResponse_Result_{Result: res},
		})
	}
	return err
}

// UpdateLibrariesIndex FIXMEDOC
func (s *arduinoCoreServerImpl) UpdateLibrariesIndex(req *rpc.UpdateLibrariesIndexRequest, stream rpc.ArduinoCoreService_UpdateLibrariesIndexServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	res, err := UpdateLibrariesIndex(stream.Context(), req,
		func(p *rpc.DownloadProgress) {
			syncSend.Send(&rpc.UpdateLibrariesIndexResponse{
				Message: &rpc.UpdateLibrariesIndexResponse_DownloadProgress{DownloadProgress: p},
			})
		},
	)
	if res != nil {
		syncSend.Send(&rpc.UpdateLibrariesIndexResponse{
			Message: &rpc.UpdateLibrariesIndexResponse_Result_{Result: res},
		})
	}
	return err
}

// Create FIXMEDOC
func (s *arduinoCoreServerImpl) Create(ctx context.Context, req *rpc.CreateRequest) (*rpc.CreateResponse, error) {
	var userAgent []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userAgent = md.Get("user-agent")
	}
	if len(userAgent) == 0 {
		userAgent = []string{"gRPCClientUnknown/0.0.0"}
	}
	return Create(req, userAgent...)
}

// Init FIXMEDOC
func (s *arduinoCoreServerImpl) Init(req *rpc.InitRequest, stream rpc.ArduinoCoreService_InitServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return Init(req, func(message *rpc.InitResponse) { syncSend.Send(message) })
}

// Version FIXMEDOC
func (s *arduinoCoreServerImpl) Version(ctx context.Context, req *rpc.VersionRequest) (*rpc.VersionResponse, error) {
	return &rpc.VersionResponse{Version: s.versionString}, nil
}

// NewSketch FIXMEDOC
func (s *arduinoCoreServerImpl) NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	return NewSketch(ctx, req)
}

// LoadSketch FIXMEDOC
func (s *arduinoCoreServerImpl) LoadSketch(ctx context.Context, req *rpc.LoadSketchRequest) (*rpc.LoadSketchResponse, error) {
	resp, err := LoadSketch(ctx, req)
	return &rpc.LoadSketchResponse{Sketch: resp}, err
}

// SetSketchDefaults FIXMEDOC
func (s *arduinoCoreServerImpl) SetSketchDefaults(ctx context.Context, req *rpc.SetSketchDefaultsRequest) (*rpc.SetSketchDefaultsResponse, error) {
	return SetSketchDefaults(ctx, req)
}

// Compile FIXMEDOC
func (s *arduinoCoreServerImpl) Compile(req *rpc.CompileRequest, stream rpc.ArduinoCoreService_CompileServer) error {
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
	compileRes, compileErr := Compile(stream.Context(), req, outStream, errStream, progressStream)
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
		return compileErr
	}
	return compileRespSendErr
}

// PlatformInstall FIXMEDOC
func (s *arduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallRequest, stream rpc.ArduinoCoreService_PlatformInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := PlatformInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformInstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return syncSend.Send(resp)
}

// PlatformDownload FIXMEDOC
func (s *arduinoCoreServerImpl) PlatformDownload(req *rpc.PlatformDownloadRequest, stream rpc.ArduinoCoreService_PlatformDownloadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := PlatformDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return syncSend.Send(resp)
}

// PlatformUninstall FIXMEDOC
func (s *arduinoCoreServerImpl) PlatformUninstall(req *rpc.PlatformUninstallRequest, stream rpc.ArduinoCoreService_PlatformUninstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := PlatformUninstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformUninstallResponse{TaskProgress: p}) },
	)
	if err != nil {
		return err
	}
	return syncSend.Send(resp)
}

// PlatformUpgrade FIXMEDOC
func (s *arduinoCoreServerImpl) PlatformUpgrade(req *rpc.PlatformUpgradeRequest, stream rpc.ArduinoCoreService_PlatformUpgradeServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := PlatformUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.PlatformUpgradeResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.PlatformUpgradeResponse{TaskProgress: p}) },
	)
	if err2 := syncSend.Send(resp); err2 != nil {
		return err2
	}
	return err
}

// PlatformSearch FIXMEDOC
func (s *arduinoCoreServerImpl) PlatformSearch(ctx context.Context, req *rpc.PlatformSearchRequest) (*rpc.PlatformSearchResponse, error) {
	return PlatformSearch(req)
}

// Upload FIXMEDOC
func (s *arduinoCoreServerImpl) Upload(req *rpc.UploadRequest, stream rpc.ArduinoCoreService_UploadServer) error {
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
	res, err := Upload(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if res != nil {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_Result{
				Result: res,
			},
		})
	}
	return err
}

// UploadUsingProgrammer FIXMEDOC
func (s *arduinoCoreServerImpl) UploadUsingProgrammer(req *rpc.UploadUsingProgrammerRequest, stream rpc.ArduinoCoreService_UploadUsingProgrammerServer) error {
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
	err := UploadUsingProgrammer(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if err != nil {
		return err
	}
	return nil
}

// SupportedUserFields FIXMEDOC
func (s *arduinoCoreServerImpl) SupportedUserFields(ctx context.Context, req *rpc.SupportedUserFieldsRequest) (*rpc.SupportedUserFieldsResponse, error) {
	return SupportedUserFields(ctx, req)
}

// BurnBootloader FIXMEDOC
func (s *arduinoCoreServerImpl) BurnBootloader(req *rpc.BurnBootloaderRequest, stream rpc.ArduinoCoreService_BurnBootloaderServer) error {
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
	resp, err := BurnBootloader(stream.Context(), req, outStream, errStream)
	outStream.Close()
	errStream.Close()
	if err != nil {
		return err
	}
	return syncSend.Send(resp)
}

// ListProgrammersAvailableForUpload FIXMEDOC
func (s *arduinoCoreServerImpl) ListProgrammersAvailableForUpload(ctx context.Context, req *rpc.ListProgrammersAvailableForUploadRequest) (*rpc.ListProgrammersAvailableForUploadResponse, error) {
	return ListProgrammersAvailableForUpload(ctx, req)
}

// LibraryDownload FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryDownload(req *rpc.LibraryDownloadRequest, stream rpc.ArduinoCoreService_LibraryDownloadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	resp, err := LibraryDownload(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryDownloadResponse{Progress: p}) },
	)
	if err != nil {
		return err
	}
	return syncSend.Send(resp)
}

// LibraryInstall FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryInstall(req *rpc.LibraryInstallRequest, stream rpc.ArduinoCoreService_LibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return LibraryInstall(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryInstallResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryInstallResponse{TaskProgress: p}) },
	)
}

// LibraryUpgrade FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryUpgrade(req *rpc.LibraryUpgradeRequest, stream rpc.ArduinoCoreService_LibraryUpgradeServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return LibraryUpgrade(
		stream.Context(), req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeResponse{TaskProgress: p}) },
	)
}

// LibraryUninstall FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryUninstall(req *rpc.LibraryUninstallRequest, stream rpc.ArduinoCoreService_LibraryUninstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return LibraryUninstall(stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUninstallResponse{TaskProgress: p}) },
	)
}

// LibraryUpgradeAll FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryUpgradeAll(req *rpc.LibraryUpgradeAllRequest, stream rpc.ArduinoCoreService_LibraryUpgradeAllServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return LibraryUpgradeAll(req,
		func(p *rpc.DownloadProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{Progress: p}) },
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.LibraryUpgradeAllResponse{TaskProgress: p}) },
	)
}

// LibraryResolveDependencies FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryResolveDependencies(ctx context.Context, req *rpc.LibraryResolveDependenciesRequest) (*rpc.LibraryResolveDependenciesResponse, error) {
	return LibraryResolveDependencies(ctx, req)
}

// LibrarySearch FIXMEDOC
func (s *arduinoCoreServerImpl) LibrarySearch(ctx context.Context, req *rpc.LibrarySearchRequest) (*rpc.LibrarySearchResponse, error) {
	return LibrarySearch(ctx, req)
}

// LibraryList FIXMEDOC
func (s *arduinoCoreServerImpl) LibraryList(ctx context.Context, req *rpc.LibraryListRequest) (*rpc.LibraryListResponse, error) {
	return LibraryList(ctx, req)
}

// ArchiveSketch FIXMEDOC
func (s *arduinoCoreServerImpl) ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchRequest) (*rpc.ArchiveSketchResponse, error) {
	return ArchiveSketch(ctx, req)
}

// ZipLibraryInstall FIXMEDOC
func (s *arduinoCoreServerImpl) ZipLibraryInstall(req *rpc.ZipLibraryInstallRequest, stream rpc.ArduinoCoreService_ZipLibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return ZipLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.ZipLibraryInstallResponse{TaskProgress: p}) },
	)
}

// GitLibraryInstall FIXMEDOC
func (s *arduinoCoreServerImpl) GitLibraryInstall(req *rpc.GitLibraryInstallRequest, stream rpc.ArduinoCoreService_GitLibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	return GitLibraryInstall(
		stream.Context(), req,
		func(p *rpc.TaskProgress) { syncSend.Send(&rpc.GitLibraryInstallResponse{TaskProgress: p}) },
	)
}

// EnumerateMonitorPortSettings FIXMEDOC
func (s *arduinoCoreServerImpl) EnumerateMonitorPortSettings(ctx context.Context, req *rpc.EnumerateMonitorPortSettingsRequest) (*rpc.EnumerateMonitorPortSettingsResponse, error) {
	return EnumerateMonitorPortSettings(ctx, req)
}

// Monitor FIXMEDOC
func (s *arduinoCoreServerImpl) Monitor(stream rpc.ArduinoCoreService_MonitorServer) error {
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
	portProxy, _, err := Monitor(stream.Context(), openReq)
	if err != nil {
		return err
	}

	// Send a message with Success set to true to notify the caller of the port being now active
	_ = syncSend.Send(&rpc.MonitorResponse{Success: true})

	cancelCtx, cancel := context.WithCancel(stream.Context())
	gracefulCloseInitiated := &atomic.Bool{}
	gracefuleCloseCtx, gracefulCloseCancel := context.WithCancel(context.Background())

	// gRPC stream receiver (gRPC data -> monitor, config, close)
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
			if closeMsg := msg.GetClose(); closeMsg {
				gracefulCloseInitiated.Store(true)
				if err := portProxy.Close(); err != nil {
					logrus.WithError(err).Debug("Error closing monitor port")
				}
				gracefulCloseCancel()
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

	// gRPC stream sender (monitor -> gRPC)
	go func() {
		defer cancel() // unlock the receiver
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
	if gracefulCloseInitiated.Load() {
		// Port closing has been initiated in the receiver
		<-gracefuleCloseCtx.Done()
	} else {
		portProxy.Close()
	}
	return nil
}

// CheckForArduinoCLIUpdates FIXMEDOC
func (s *arduinoCoreServerImpl) CheckForArduinoCLIUpdates(ctx context.Context, req *rpc.CheckForArduinoCLIUpdatesRequest) (*rpc.CheckForArduinoCLIUpdatesResponse, error) {
	return updatecheck.CheckForArduinoCLIUpdates(ctx, req)
}

// CleanDownloadCacheDirectory FIXMEDOC
func (s *arduinoCoreServerImpl) CleanDownloadCacheDirectory(ctx context.Context, req *rpc.CleanDownloadCacheDirectoryRequest) (*rpc.CleanDownloadCacheDirectoryResponse, error) {
	return cache.CleanDownloadCacheDirectory(ctx, req)
}
