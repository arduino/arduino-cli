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
	"time"

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
		fmt.Printf("%v\n", err)
		return &dbg.StreamingOpenResp{}, nil // TODO: send error in response
	}
	defer in.Close()

	cmd.Stdout = out

	err = cmd.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
		return &dbg.StreamingOpenResp{}, nil // TODO: send error in response
	}

	// now we can read the other commands and re-route to the Debug Client...
	go func() {
		for {
			if command, err := inStream.Recv(); err != nil {
				break
			} else if _, err := in.Write(command.GetData()); err != nil {
				break
			}
		}
		time.Sleep(time.Second)
		cmd.Process.Kill()
	}()

	err = cmd.Wait() // TODO: handle err
	return &dbg.StreamingOpenResp{}, nil
}
