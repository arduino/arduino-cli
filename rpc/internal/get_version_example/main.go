// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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
	"log"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Establish a connection with the gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	cancel()
	if err != nil {
		log.Println(err)
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// Create an instance of the gRPC clients.
	cli := rpc.NewArduinoCoreServiceClient(conn)

	// Now we can call various methods of the API...
	versionResp, err := cli.Version(context.Background(), &rpc.VersionRequest{})
	if err != nil {
		log.Fatalf("Error getting version: %s", err)
	}
	log.Printf("arduino-cli version: %v", versionResp.GetVersion())
}
