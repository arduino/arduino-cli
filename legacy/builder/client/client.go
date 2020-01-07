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

// Package main implements a simple gRPC client that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It interacts with the route guide service whose definition can be found in routeguide/route_guide.proto.
package main

import (
	"io"
	"log"

	pb "github.com/arduino/arduino-cli/legacy/builder/grpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// printFeature gets the feature for the given point.
func autocomplete(client pb.BuilderClient, in *pb.BuildParams) {
	resp, err := client.Autocomplete(context.Background(), in)
	if err != nil {
		log.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
	}
	log.Println(resp)
}

// printFeatures lists all the features within the given bounding Rectangle.
func build(client pb.BuilderClient, in *pb.BuildParams) {
	stream, err := client.Build(context.Background(), in)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println(line)
	}
}

func main() {
	conn, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuilderClient(conn)

	exampleParames := pb.BuildParams{
		BuiltInLibrariesFolders: "/ssd/Arduino-master/build/linux/work/libraries",
		CustomBuildProperties:   "build.warn_data_percentage=75",
		FQBN:                    "arduino:avr:mega:cpu=atmega2560",
		HardwareFolders:         "/ssd/Arduino-master/build/linux/work/hardware,/home/martino/.arduino15/packages,/home/martino/eslov-sk/hardware",
		OtherLibrariesFolders:   "/home/martino/eslov-sk/libraries",
		ArduinoAPIVersion:       "10805",
		SketchLocation:          "/home/martino/eslov-sk/libraries/WiFi101/examples/ScanNetworks/ScanNetworks.ino",
		ToolsFolders:            "/ssd/Arduino-master/build/linux/work/tools-builder,/ssd/Arduino-master/build/linux/work/hardware/tools/avr,/home/martino/.arduino15/packages",
		Verbose:                 true,
		WarningsLevel:           "all",
		BuildCachePath:          "/tmp/arduino_cache_761418/",
		CodeCompleteAt:          "/home/martino/eslov-sk/libraries/WiFi101/examples/ScanNetworks/ScanNetworks.ino:56:9",
	}

	//build(client, &exampleParames)
	autocomplete(client, &exampleParames)
}
