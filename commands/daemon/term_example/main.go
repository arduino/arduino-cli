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
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This program exercise CLI monitor functionality.

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// Create and initialize a CLI instance
	cli := commands.NewArduinoCoreServiceClient(conn)

	var instance *commands.Instance
	if resp, err := cli.Create(context.Background(), &commands.CreateRequest{}); err != nil {
		log.Fatal("Create:", err)
	} else {
		instance = resp.Instance
	}

	if respStream, err := cli.Init(context.Background(), &commands.InitRequest{Instance: instance}); err != nil {
		log.Fatal("Init:", err)
	} else {
		for {
			resp, err := respStream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Fatal("Init:", err)
			}
			fmt.Println(resp)
		}
	}

	// List boards and take the first available port
	var port *commands.Port
	if resp, err := cli.BoardList(context.Background(), &commands.BoardListRequest{Instance: instance}); err != nil {
		log.Fatal("BoardList:", err)
	} else {
		ports := resp.GetPorts()
		if len(ports) == 0 {
			log.Fatal("No port to connect!")
		}
		port = ports[0].Port
	}
	fmt.Println("Detected port:", port.Label, port.ProtocolLabel)

	connectToPort(cli, instance, port)
	time.Sleep(5 * time.Second)
	connectToPort(cli, instance, port)
	time.Sleep(5 * time.Second)
}

func connectToPort(cli commands.ArduinoCoreServiceClient, instance *commands.Instance, port *commands.Port) {
	// Connect to the port monitor
	fmt.Println("Connecting to port", port)

	ctx, cancel := context.WithCancel(context.Background())
	monitorClient, err := cli.Monitor(ctx)
	if err != nil {
		log.Fatal("Error opening Monitor:", err)
	}
	if err := monitorClient.Send(&commands.MonitorRequest{
		Instance: instance,
		Port:     port,
	}); err != nil {
		log.Fatal("Error sending Monitor config:", err)
	}

	go func() {
		for {
			resp, err := monitorClient.Recv()
			if err != nil {
				fmt.Println("     RECV-ERR:", err)
				break
			}
			fmt.Println("     RECV:", resp)
		}
	}()

	hello := &commands.MonitorRequest{
		TxData: []byte("HELLO!"),
	}
	fmt.Println("Send:", hello)
	if err := monitorClient.Send(hello); err != nil {
		log.Fatal("Monitor send HELLO:", err)
	}

	time.Sleep(15 * time.Second)

	fmt.Println("Closing Monitor")
	if err := monitorClient.CloseSend(); err != nil {
		log.Fatal("Monitor close send:", err)
	}
	<-monitorClient.Context().Done()

	cancel()
}
