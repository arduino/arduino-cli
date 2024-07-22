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
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc/metadata"

	"fmt"
	"io"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

type debugServer struct {
	ctx      context.Context
	req      atomic.Pointer[rpc.GetDebugConfigRequest]
	in       io.Reader
	out      io.Writer
	resultCB func(*rpc.DebugResponse_Result)
}

func (s *debugServer) Send(resp *rpc.DebugResponse) error {
	if len(resp.GetData()) > 0 {
		if _, err := s.out.Write(resp.GetData()); err != nil {
			return err
		}
	}
	if res := resp.GetResult(); res != nil {
		s.resultCB(res)
	}
	return nil
}

func (s *debugServer) Recv() (r *rpc.DebugRequest, e error) {
	if conf := s.req.Swap(nil); conf != nil {
		return &rpc.DebugRequest{DebugRequest: conf}, nil
	}
	buff := make([]byte, 4096)
	n, err := s.in.Read(buff)
	if err != nil {
		return nil, err
	}
	return &rpc.DebugRequest{Data: buff[:n]}, nil
}

func (s *debugServer) Context() context.Context     { return s.ctx }
func (s *debugServer) RecvMsg(m any) error          { return nil }
func (s *debugServer) SendHeader(metadata.MD) error { return nil }
func (s *debugServer) SendMsg(m any) error          { return nil }
func (s *debugServer) SetHeader(metadata.MD) error  { return nil }
func (s *debugServer) SetTrailer(metadata.MD)       {}

// DebugServerToStreams creates a debug server that proxies the data to the given io streams.
// The GetDebugConfigRequest is used to configure the debbuger. sig is a channel that can be
// used to send os.Interrupt to the debug process. resultCB is a callback function that will
// receive the Debug result.
func DebugServerToStreams(
	ctx context.Context,
	req *rpc.GetDebugConfigRequest,
	in io.Reader, out io.Writer,
	sig chan os.Signal,
	resultCB func(*rpc.DebugResponse_Result),
) rpc.ArduinoCoreService_DebugServer {
	server := &debugServer{
		ctx:      ctx,
		in:       in,
		out:      out,
		resultCB: resultCB,
	}
	server.req.Store(req)
	return server
}

// Debug starts a debugging session. The first message passed through the `Debug` request must
// contain DebugRequest configuration params and no data.
func (s *arduinoCoreServerImpl) Debug(stream rpc.ArduinoCoreService_DebugServer) error {
	// Utility functions
	syncSend := NewSynchronizedSend(stream.Send)
	sendResult := func(res *rpc.DebugResponse_Result) error {
		return syncSend.Send(&rpc.DebugResponse{Message: &rpc.DebugResponse_Result_{Result: res}})
	}
	sendData := func(data []byte) {
		_ = syncSend.Send(&rpc.DebugResponse{Message: &rpc.DebugResponse_Data{Data: data}})
	}

	// Grab the first message
	debugConfReqMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	// Ensure it's a config message and not data
	debugConfReq := debugConfReqMsg.GetDebugRequest()
	if debugConfReq == nil {
		return errors.New(i18n.Tr("First message must contain debug request, not data"))
	}

	// Launch debug recipe attaching stdin and out to grpc streaming
	signalChan := make(chan os.Signal)
	defer close(signalChan)
	outStream := feedStreamTo(sendData)
	defer outStream.Close()
	inStream := consumeStreamFrom(func() ([]byte, error) {
		command, err := stream.Recv()
		if command.GetSendInterrupt() {
			signalChan <- os.Interrupt
		}
		return command.GetData(), err
	})

	pme, release, err := instances.GetPackageManagerExplorer(debugConfReq.GetInstance())
	if err != nil {
		return err
	}
	defer release()

	// Exec debugger
	commandLine, err := getCommandLine(debugConfReq, pme)
	if err != nil {
		return err
	}
	entry := logrus.NewEntry(logrus.StandardLogger())
	for i, param := range commandLine {
		entry = entry.WithField(fmt.Sprintf("param%d", i), param)
	}
	entry.Debug("Executing debugger")
	cmd, err := paths.NewProcess(pme.GetEnvVarsForSpawnedProcess(), commandLine...)
	if err != nil {
		return &cmderrors.FailedDebugError{Message: i18n.Tr("Cannot execute debug tool"), Cause: err}
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return sendResult(&rpc.DebugResponse_Result{Error: err.Error()})
	}
	defer in.Close()
	cmd.RedirectStdoutTo(io.Writer(outStream))
	cmd.RedirectStderrTo(io.Writer(outStream))
	if err := cmd.Start(); err != nil {
		return sendResult(&rpc.DebugResponse_Result{Error: err.Error()})
	}

	go func() {
		for sig := range signalChan {
			cmd.Signal(sig)
		}
	}()
	go func() {
		io.Copy(in, inStream)
		time.Sleep(time.Second)
		cmd.Kill()
	}()
	if err := cmd.Wait(); err != nil {
		return sendResult(&rpc.DebugResponse_Result{Error: err.Error()})
	}
	return sendResult(&rpc.DebugResponse_Result{})
}

// getCommandLine compose a debug command represented by a core recipe
func getCommandLine(req *rpc.GetDebugConfigRequest, pme *packagemanager.Explorer) ([]string, error) {
	debugInfo, err := getDebugProperties(req, pme, false)
	if err != nil {
		return nil, err
	}

	cmdArgs := []string{}
	add := func(s string) { cmdArgs = append(cmdArgs, s) }

	// Add path to GDB Client to command line
	var gdbPath *paths.Path
	switch debugInfo.GetToolchain() {
	case "gcc":
		gdbexecutable := debugInfo.GetToolchainPrefix() + "-gdb"
		if runtime.GOOS == "windows" {
			gdbexecutable += ".exe"
		}
		gdbPath = paths.New(debugInfo.GetToolchainPath()).Join(gdbexecutable)
	default:
		return nil, &cmderrors.FailedDebugError{Message: i18n.Tr("Toolchain '%s' is not supported", debugInfo.GetToolchain())}
	}
	add(gdbPath.String())

	// Set GDB interpreter (default value should be "console")
	gdbInterpreter := req.GetInterpreter()
	if gdbInterpreter == "" {
		gdbInterpreter = "console"
	}
	add("--interpreter=" + gdbInterpreter)
	if gdbInterpreter != "console" {
		add("-ex")
		add("set pagination off")
	}

	// Add extra GDB execution commands
	add("-ex")
	add("set remotetimeout 5")

	// Extract path to GDB Server
	switch debugInfo.GetServer() {
	case "openocd":
		var openocdConf rpc.DebugOpenOCDServerConfiguration
		if err := debugInfo.GetServerConfiguration().UnmarshalTo(&openocdConf); err != nil {
			return nil, err
		}

		serverCmd := fmt.Sprintf(`target extended-remote | "%s"`, debugInfo.GetServerPath())

		if cfg := openocdConf.GetScriptsDir(); cfg != "" {
			serverCmd += fmt.Sprintf(` -s "%s"`, cfg)
		}

		for _, script := range openocdConf.GetScripts() {
			serverCmd += fmt.Sprintf(` --file "%s"`, script)
		}

		serverCmd += ` -c "gdb_port pipe"`
		serverCmd += ` -c "telnet_port 0"`

		add("-ex")
		add(serverCmd)

	default:
		return nil, &cmderrors.FailedDebugError{Message: i18n.Tr("GDB server '%s' is not supported", debugInfo.GetServer())}
	}

	// Add executable
	add(debugInfo.GetExecutable())

	// Transform every path to forward slashes (on Windows some tools further
	// escapes the command line so the backslash "\" gets in the way).
	for i, param := range cmdArgs {
		cmdArgs[i] = filepath.ToSlash(param)
	}

	return cmdArgs, nil
}
