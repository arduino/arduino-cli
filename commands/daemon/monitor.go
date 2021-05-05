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
	"errors"
	"io"
	"sync/atomic"

	"github.com/arduino/arduino-cli/arduino/monitors"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/monitor/v1"
)

// MonitorService implements the `Monitor` service
type MonitorService struct {
	rpc.UnimplementedMonitorServiceServer
}

// StreamingOpen returns a stream response that can be used to fetch data from the
// monitor target. The first message passed through the `StreamingOpenReq` must
// contain monitor configuration params, not data.
func (s *MonitorService) StreamingOpen(stream rpc.MonitorService_StreamingOpenServer) error {
	// grab the first message
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// ensure it's a config message and not data
	config := msg.GetConfig()
	if config == nil {
		return errors.New("first message must contain monitor configuration, not data")
	}

	// select which type of monitor we need
	var mon monitors.Monitor
	switch config.GetType() {
	case rpc.MonitorConfig_TARGET_TYPE_SERIAL:
		// grab port speed from additional config data
		var baudRate float64
		addCfg := config.GetAdditionalConfig()
		for k, v := range addCfg.GetFields() {
			if k == "BaudRate" {
				baudRate = v.GetNumberValue()
				break
			}
		}

		// get the Monitor instance
		var err error
		if mon, err = monitors.OpenSerialMonitor(config.GetTarget(), int(baudRate)); err != nil {
			return err
		}

	case rpc.MonitorConfig_TARGET_TYPE_NULL:
		if addCfg, ok := config.GetAdditionalConfig().AsMap()["OutputRate"]; !ok {
			mon = monitors.OpenNullMonitor(100.0) // 100 bytes per second as default
		} else if outputRate, ok := addCfg.(float64); !ok {
			return errors.New("OutputRate in Null monitor must be a float64")
		} else {
			// get the Monitor instance
			mon = monitors.OpenNullMonitor(outputRate)
		}
	}

	// we'll use these channels to communicate with the goroutines
	// handling the stream and the target respectively
	streamClosed := make(chan error)
	targetClosed := make(chan error)

	// set rate limiting window
	bufferSize := int(config.GetRecvRateLimitBuffer())
	rateLimitEnabled := (bufferSize > 0)
	if !rateLimitEnabled {
		bufferSize = 1024
	}
	buffer := make([]byte, bufferSize)
	bufferUsed := 0

	var writeSlots int32

	// now we can read the other messages and re-route to the monitor...
	go func() {
		for {
			msg, err := stream.Recv()
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

			if rateLimitEnabled {
				// Increase rate limiter write slots
				atomic.AddInt32(&writeSlots, msg.GetRecvAcknowledge())
			}

			if _, err := mon.Write(msg.GetData()); err != nil {
				// error writing to target
				targetClosed <- err
				break
			}
		}
	}()

	// ...and read from the monitor and forward to the output stream
	go func() {
		dropBuffer := make([]byte, 10240)
		dropped := 0
		for {
			if bufferUsed < bufferSize {
				if n, err := mon.Read(buffer[bufferUsed:]); err != nil {
					// error reading from target
					targetClosed <- err
					break
				} else if n == 0 {
					// target was closed
					targetClosed <- nil
					break
				} else {
					bufferUsed += n
				}
			} else {
				// FIXME: a very rare condition but still...
				// we may be waiting here while, in the meantime, a transmit slot is
				// freed: in this case the (filled) buffer will stay in the server
				// until the following Read exits (-> the next char arrives from the
				// monitor).

				if n, err := mon.Read(dropBuffer); err != nil {
					// error reading from target
					targetClosed <- err
					break
				} else if n == 0 {
					// target was closed
					targetClosed <- nil
					break
				} else {
					dropped += n
				}
			}

			slots := atomic.LoadInt32(&writeSlots)
			if !rateLimitEnabled || slots > 0 {
				if err = stream.Send(&rpc.StreamingOpenResponse{
					Data:    buffer[:bufferUsed],
					Dropped: int32(dropped),
				}); err != nil {
					// error sending to stream
					streamClosed <- err
					break
				}
				bufferUsed = 0
				dropped = 0

				// Rate limit, filling all the available window
				if rateLimitEnabled {
					slots = atomic.AddInt32(&writeSlots, -1)
				}
			}
		}
	}()

	// let goroutines route messages from/to the monitor
	// until either the client closes the stream or the
	// monitor target is closed
	for {
		select {
		case err := <-streamClosed:
			mon.Close()
			return err
		case err := <-targetClosed:
			return err
		}
	}
}
