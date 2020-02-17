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

package debug

import (
	"context"
	"fmt"
	"io"

	"github.com/arduino/arduino-cli/executils"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
)

// Debug FIXMEDOC
func Debug(ctx context.Context, req *dbg.DebugReq, inStream dbg.Debug_StreamingOpenServer, out io.Writer) (*dbg.StreamingOpenResp, error) {
	cmdArgs := []string{"gdb"}
	// Run Tool
	cmd, err := executils.Command(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("cannot execute upload tool: %s", err)
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("%v\n", err)
		return &dbg.StreamingOpenResp{}, nil // TODO: send error in response
	}
	defer in.Close()

	cmd.Stdout = out

	err = cmd.Start()
	if err != nil {
		fmt.Println("%v\n", err)
		return &dbg.StreamingOpenResp{}, nil // TODO: send error in response
	}

	// we'll use these channels to communicate with the goroutines
	// handling the stream and the target respectively
	streamClosed := make(chan error)
	targetClosed := make(chan error)
	defer close(streamClosed)
	defer close(targetClosed)

	// now we can read the other commands and re-route to the Debug Client...
	go func() {
		for {
			command, err := inStream.Recv()
			if err == io.EOF {
				// stream was closed
				streamClosed <- nil
				break
			}

			if err != nil {
				// error reading from stream
				streamClosed <- err
				break
			}

			if _, err := in.Write(command.GetData()); err != nil {
				// error writing to target
				targetClosed <- err
				break
			}
		}
	}()

	// let goroutines route messages from/to the Debug
	// until either the client closes the stream or the
	// Debug target is closed
	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			cmd.Wait()
		case err := <-streamClosed:
			fmt.Println("streamClosed")
			cmd.Process.Kill()
			cmd.Wait()
			return &dbg.StreamingOpenResp{}, err // TODO: send error in response
		case err := <-targetClosed:
			fmt.Println("targetClosed")
			cmd.Process.Kill()
			cmd.Wait()
			return &dbg.StreamingOpenResp{}, err // TODO: send error in response
		}
	}

	return &dbg.StreamingOpenResp{}, nil
}
