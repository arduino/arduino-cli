// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
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
	"io"
	"log"

	"github.com/arduino/arduino-cli/arduino/monitors"
	rpc "github.com/arduino/arduino-cli/rpc/monitor"
)

// MonitorService implements the `Monitor` service
type MonitorService struct {
}

// StreamingOpen returns a stream response that can be used to fetch data from the
// monitor target. The first message passed through the `StreamingOpenReq` must
// contain monitor configuration params, not data.
func (s *MonitorService) StreamingOpen(stream rpc.Monitor_StreamingOpenServer) error {
	// grab the first message
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// ensure it's a config message and not data
	config := msg.GetMonitorConfig()
	if config == nil {
		return fmt.Errorf("first message must contain monitor configuration, not data")
	}

	// select which type of monitor we need
	var mon monitors.Monitor
	switch config.GetType() {
	case rpc.MonitorConfig_SERIAL:
		// grab port speed from additional config data
		var baudRate float64
		addCfg := config.GetAdditionalConfig()
		for k, v := range addCfg.GetFields() {
			if k == "BaudRate" {
				baudRate = v.GetNumberValue()
				break
			}
		}

		// default baud rate if not provided
		if baudRate == 0 {
			baudRate = 9600
		}

		var err error
		if mon, err = monitors.OpenSerialMonitor(config.GetTarget(), int(baudRate)); err != nil {
			return err
		}
	}

	// now we can stream the other messages and re-route to the monitor
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				// connection closed, exit
				break
			}

			if err != nil {
				// error, exit
				log.Fatal(err)
				break
			}

			mon.Write(msg.GetData())
		}
	}()

	// read from the monitor and forward to the output stream

	buf := make([]byte, 8)
	for {
		n, err := mon.Read(buf)
		if err != nil {
			return err
		}

		if n == 0 {
			// port was closed
			return nil
		}

		if err = stream.Send(&rpc.StreamingOpenResp{
			Data: buf[:n],
		}); err != nil {
			return err
		}
	}
}
