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
	"io"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

// BurnBootloaderToServerStreams return a server stream that forwards the output and error streams to the provided io.Writers
func BurnBootloaderToServerStreams(ctx context.Context, outStrem, errStream io.Writer) rpc.ArduinoCoreService_BurnBootloaderServer {
	stream := streamResponseToCallback(ctx, func(resp *rpc.BurnBootloaderResponse) error {
		if outData := resp.GetOutStream(); len(outData) > 0 {
			_, err := outStrem.Write(outData)
			return err
		}
		if errData := resp.GetErrStream(); len(errData) > 0 {
			_, err := errStream.Write(errData)
			return err
		}
		return nil
	})
	return stream
}

// BurnBootloader performs the burn bootloader action
func (s *arduinoCoreServerImpl) BurnBootloader(req *rpc.BurnBootloaderRequest, stream rpc.ArduinoCoreService_BurnBootloaderServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.BurnBootloaderResponse{
			Message: &rpc.BurnBootloaderResponse_OutStream{
				OutStream: data,
			},
		})
	})
	defer outStream.Close()
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.BurnBootloaderResponse{
			Message: &rpc.BurnBootloaderResponse_ErrStream{
				ErrStream: data,
			},
		})
	})
	defer errStream.Close()

	logrus.
		WithField("fqbn", req.GetFqbn()).
		WithField("port", req.GetPort()).
		WithField("programmer", req.GetProgrammer()).
		Trace("BurnBootloader started", req.GetFqbn())

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer release()

	if _, err := s.runProgramAction(
		stream.Context(),
		pme,
		nil, // sketch
		"",  // importFile
		"",  // importDir
		req.GetFqbn(),
		req.GetPort(),
		req.GetProgrammer(),
		req.GetVerbose(),
		req.GetVerify(),
		true, // burnBootloader
		outStream,
		errStream,
		req.GetDryRun(),
		map[string]string{}, // User fields
		req.GetUploadProperties(),
	); err != nil {
		return err
	}
	return syncSend.Send(&rpc.BurnBootloaderResponse{})
}
