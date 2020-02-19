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
	"fmt"
	cmd "github.com/arduino/arduino-cli/commands/debug"
	"io"

	dbg "github.com/arduino/arduino-cli/rpc/debug"
)

// DebugService implements the `Debug` service
type DebugService struct{}

// Debug returns a stream response that can be used to fetch data from the
// target. The first message passed through the `Debug` request must
// contain DebugConfigReq configuration params, not data.
func (s *DebugService) Debug(stream dbg.Debug_DebugServer) error {

	// grab the first message
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// ensure it's a config message and not data
	req := msg.GetDebugReq()
	if req == nil {
		return fmt.Errorf("first message must contain debug request, not data")
	}

	// launch debug recipe attaching stdin and out to grpc streaming
	resp, err := cmd.Debug(stream.Context(), req,
		copyStream(func() ([]byte, error) {
			command, err := stream.Recv()
			return command.GetData(), err
		}),
		feedStream(func(data []byte) {
			stream.Send(&dbg.DebugResp{Data: data})
		}))
	if err != nil {
		return (err)
	}
	return stream.Send(resp)
}

func copyStream(streamIn func() ([]byte, error)) io.Reader {

	r, w := io.Pipe()
	go func() {
		for {
			if data, err := streamIn(); err != nil {
				return
			} else if _, err := w.Write(data); err != nil {
				return
			}
		}
	}()
	return r
}
