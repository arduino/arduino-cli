// This file is part of arduino-cli.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arduino/arduino-cli/rpc/monitor"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	dataDir string
)

// This program exercise monitor rate limiting functionality.

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// Open a monitor instance
	mon := monitor.NewMonitorClient(conn)
	stream, err := mon.StreamingOpen(context.Background())
	if err != nil {
		log.Fatal("Error opening stream:", err)
	}

	additionalConf, err := structpb.NewStruct(
		map[string]interface{}{"OutputRate": float64(1000000)},
	)
	if err != nil {
		log.Fatal("Error creating config:", err)
	}

	if err := stream.Send(&monitor.StreamingOpenReq{
		Content: &monitor.StreamingOpenReq_MonitorConfig{
			MonitorConfig: &monitor.MonitorConfig{
				Type:                monitor.MonitorConfig_NULL,
				AdditionalConfig:    additionalConf,
				RecvRateLimitBuffer: 1024,
			},
		},
	}); err != nil {
		log.Fatal("Error opening stream:", err)
	}

	if err := stream.Send(&monitor.StreamingOpenReq{
		Content: &monitor.StreamingOpenReq_RecvAcknowledge{
			RecvAcknowledge: 5,
		},
	}); err != nil {
		log.Fatal("Error replenishing recv window:", err)
	}

	for {
		r, err := stream.Recv()
		if err != nil {
			log.Fatal("Error receiving from server:", err)
		}
		if l := len(r.Data); l > 0 {
			fmt.Printf("RECV %d bytes\n", l)
		}
		if r.Dropped > 0 {
			fmt.Printf("DROPPED %d bytes!!!\n", r.Dropped)
		}
		if err := stream.Send(&monitor.StreamingOpenReq{
			Content: &monitor.StreamingOpenReq_RecvAcknowledge{
				RecvAcknowledge: 1,
			},
		}); err != nil {
			log.Fatal("Error replenishing recv window:", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
