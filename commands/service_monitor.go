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
	"strconv"
	"sync/atomic"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	pluggableMonitor "github.com/arduino/arduino-cli/internal/arduino/monitor"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

type monitorPipeServer struct {
	ctx context.Context
	req atomic.Pointer[rpc.MonitorPortOpenRequest]
	in  *nio.PipeReader
	out *nio.PipeWriter
}

func (s *monitorPipeServer) Send(resp *rpc.MonitorResponse) error {
	if len(resp.GetRxData()) > 0 {
		if _, err := s.out.Write(resp.GetRxData()); err != nil {
			return err
		}
	}
	return nil
}

func (s *monitorPipeServer) Recv() (r *rpc.MonitorRequest, e error) {
	if conf := s.req.Swap(nil); conf != nil {
		return &rpc.MonitorRequest{Message: &rpc.MonitorRequest_OpenRequest{OpenRequest: conf}}, nil
	}
	buff := make([]byte, 4096)
	n, err := s.in.Read(buff)
	if err != nil {
		return nil, err
	}
	return &rpc.MonitorRequest{Message: &rpc.MonitorRequest_TxData{TxData: buff[:n]}}, nil
}

func (s *monitorPipeServer) Context() context.Context {
	return s.ctx
}

func (s *monitorPipeServer) RecvMsg(m any) error          { return nil }
func (s *monitorPipeServer) SendHeader(metadata.MD) error { return nil }
func (s *monitorPipeServer) SendMsg(m any) error          { return nil }
func (s *monitorPipeServer) SetHeader(metadata.MD) error  { return nil }
func (s *monitorPipeServer) SetTrailer(metadata.MD)       {}

type monitorPipeClient struct {
	in    *nio.PipeReader
	out   *nio.PipeWriter
	close func()
}

func (s *monitorPipeClient) Read(buff []byte) (n int, err error) {
	return s.in.Read(buff)
}

func (s *monitorPipeClient) Write(buff []byte) (n int, err error) {
	return s.out.Write(buff)
}

func (s *monitorPipeClient) Close() error {
	s.in.Close()
	s.out.Close()
	s.close()
	return nil
}

// configurer is the minimal subset of the monitor used to apply settings.
type configurer interface {
	Configure(parameterName, value string) error
}

// applyBufferConfig translates the gRPC buffer config into pluggable-monitor CONFIGURE calls.
func applyBufferConfig(c configurer, cfg *rpc.MonitorBufferConfig) {
	if cfg == nil {
		return
	}
	if v := cfg.GetHighWaterMarkBytes(); v > 0 {
		_ = c.Configure("_buffer.hwm", strconv.Itoa(int(v)))
	}
	// Interval (0 disables)
	_ = c.Configure("_buffer.interval_ms", strconv.Itoa(int(cfg.GetFlushIntervalMs())))
	// Line buffering
	_ = c.Configure("_buffer.line", strconv.FormatBool(cfg.GetLineBuffering()))
	// Queue capacity
	if v := cfg.GetFlushQueueCapacity(); v > 0 {
		_ = c.Configure("_buffer.queue", strconv.Itoa(int(v)))
	}
	// Overflow strategy (default to drop if unspecified)
	switch cfg.GetOverflowStrategy() {
	case rpc.BufferOverflowStrategy_BUFFER_OVERFLOW_STRATEGY_WAIT:
		_ = c.Configure("_buffer.overflow", "wait")
	default: // unspecified or drop
		_ = c.Configure("_buffer.overflow", "drop")
	}
	// Bounded wait for overflow (ms)
	_ = c.Configure("_buffer.overflow_wait_ms", strconv.Itoa(int(cfg.GetOverflowWaitMs())))
}

// MonitorServerToReadWriteCloser creates a monitor server that proxies the data to a ReadWriteCloser.
// The server is returned along with the ReadWriteCloser that can be used to send and receive data
// to the server. The MonitorPortOpenRequest is used to configure the monitor.
func MonitorServerToReadWriteCloser(ctx context.Context, req *rpc.MonitorPortOpenRequest) (rpc.ArduinoCoreService_MonitorServer, io.ReadWriteCloser) {
	server := &monitorPipeServer{}
	client := &monitorPipeClient{}
	server.req.Store(req)
	server.ctx, client.close = context.WithCancel(ctx)
	client.in, server.out = nio.Pipe(buffer.New(32 * 1024))
	server.in, client.out = nio.Pipe(buffer.New(32 * 1024))
	return server, client
}

// Monitor opens a port monitor and streams data back and forth until the request is kept alive.
func (s *arduinoCoreServerImpl) Monitor(stream rpc.ArduinoCoreService_MonitorServer) error {
	// The configuration must be sent on the first message
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	openReq := req.GetOpenRequest()
	if openReq == nil {
		return &cmderrors.InvalidInstanceError{}
	}

	pme, release, err := instances.GetPackageManagerExplorer(openReq.GetInstance())
	if err != nil {
		return err
	}
	monitor, boardSettings, err := findMonitorAndSettingsForProtocolAndBoard(pme, openReq.GetPort().GetProtocol(), openReq.GetFqbn())
	release()
	if err != nil {
		return err
	}
	if err := monitor.Run(); err != nil {
		return &cmderrors.FailedMonitorError{Cause: err}
	}
	if _, err := monitor.Describe(); err != nil {
		monitor.Quit()
		return &cmderrors.FailedMonitorError{Cause: err}
	}
	if portConfig := openReq.GetPortConfiguration(); portConfig != nil {
		for _, setting := range portConfig.GetSettings() {
			boardSettings.Remove(setting.GetSettingId())
			if err := monitor.Configure(setting.GetSettingId(), setting.GetValue()); err != nil {
				logrus.Errorf("Could not set configuration %s=%s: %s", setting.GetSettingId(), setting.GetValue(), err)
			}
		}
	}
	for setting, value := range boardSettings.AsMap() {
		monitor.Configure(setting, value)
	}
	applyBufferConfig(monitor, openReq.GetBufferConfig())
	monitorIO, err := monitor.Open(openReq.GetPort().GetAddress(), openReq.GetPort().GetProtocol())
	if err != nil {
		monitor.Quit()
		return &cmderrors.FailedMonitorError{Cause: err}
	}
	logrus.Infof("Port %s successfully opened", openReq.GetPort().GetAddress())
	monitorClose := func() error {
		monitor.Close()
		return monitor.Quit()
	}

	// Send a message with Success set to true to notify the caller of the port being now active
	syncSend := NewSynchronizedSend(stream.Send)
	_ = syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_Success{Success: true}})

	ctx, cancel := context.WithCancel(stream.Context())
	gracefulCloseInitiated := &atomic.Bool{}
	gracefulCloseCtx, gracefulCloseCancel := context.WithCancel(context.Background())

	// gRPC stream receiver (gRPC data -> monitor, config, close)
	go func() {
		defer cancel()
		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_Error{Error: err.Error()}})
				return
			}
			if conf := msg.GetUpdatedConfiguration(); conf != nil {
				for _, c := range conf.GetSettings() {
					if err := monitor.Configure(c.GetSettingId(), c.GetValue()); err != nil {
						syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_Error{Error: err.Error()}})
					}
				}
			}
			if closeMsg := msg.GetClose(); closeMsg {
				gracefulCloseInitiated.Store(true)
				if err := monitorClose(); err != nil {
					logrus.WithError(err).Debug("Error closing monitor port")
				}
				gracefulCloseCancel()
			}
			tx := msg.GetTxData()
			for len(tx) > 0 {
				n, err := monitorIO.Write(tx)
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_Error{Error: err.Error()}})
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
			n, err := monitorIO.Read(buff)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_Error{Error: err.Error()}})
				break
			}
			if err := syncSend.Send(&rpc.MonitorResponse{Message: &rpc.MonitorResponse_RxData{RxData: buff[:n]}}); err != nil {
				break
			}
		}
	}()

	<-ctx.Done()
	if gracefulCloseInitiated.Load() {
		// Port closing has been initiated in the receiver
		<-gracefulCloseCtx.Done()
	} else {
		monitorClose()
	}
	return nil
}

func findMonitorAndSettingsForProtocolAndBoard(pme *packagemanager.Explorer, protocol, fqbnIn string) (*pluggableMonitor.PluggableMonitor, *properties.Map, error) {
	if protocol == "" {
		return nil, nil, &cmderrors.MissingPortProtocolError{}
	}

	var monitorDepOrRecipe *cores.MonitorDependency
	boardSettings := properties.NewMap()

	// If a board is specified search the monitor in the board package first
	if fqbnIn != "" {
		fqbn, err := fqbn.Parse(fqbnIn)
		if err != nil {
			return nil, nil, &cmderrors.InvalidFQBNError{Cause: err}
		}

		_, boardPlatform, _, boardProperties, _, err := pme.ResolveFQBN(fqbn)
		if err != nil {
			return nil, nil, &cmderrors.UnknownFQBNError{Cause: err}
		}

		boardSettings = cores.GetMonitorSettings(protocol, boardProperties)

		if mon, ok := boardPlatform.Monitors[protocol]; ok {
			monitorDepOrRecipe = mon
		} else if recipe, ok := boardPlatform.MonitorsDevRecipes[protocol]; ok {
			// If we have a recipe we must resolve it
			cmdLine := boardProperties.ExpandPropsInString(recipe)
			cmdArgs, _ := properties.SplitQuotedString(cmdLine, `"'`, false)
			id := fmt.Sprintf("%s-%s", boardPlatform, protocol)
			return pluggableMonitor.New(id, cmdArgs...), boardSettings, nil
		}
	}

	if monitorDepOrRecipe == nil {
		// Otherwise look in all package for a suitable monitor
		for _, platformRel := range pme.InstalledPlatformReleases() {
			if mon, ok := platformRel.Monitors[protocol]; ok {
				monitorDepOrRecipe = mon
				break
			}
		}
	}

	if monitorDepOrRecipe == nil {
		return nil, nil, &cmderrors.NoMonitorAvailableForProtocolError{Protocol: protocol}
	}

	// If it is a monitor dependency, resolve tool and create a monitor client
	tool := pme.FindMonitorDependency(monitorDepOrRecipe)
	if tool == nil {
		return nil, nil, &cmderrors.MonitorNotFoundError{Monitor: monitorDepOrRecipe.String()}
	}

	return pluggableMonitor.New(
		monitorDepOrRecipe.Name,
		tool.InstallDir.Join(monitorDepOrRecipe.Name).String(),
	), boardSettings, nil
}
