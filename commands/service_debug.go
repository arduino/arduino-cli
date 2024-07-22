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
	"errors"
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"

	"fmt"
	"io"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// Debug returns a stream response that can be used to fetch data from the
// target. The first message passed through the `Debug` request must
// contain DebugRequest configuration params, not data.
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
