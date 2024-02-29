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
	"fmt"
	"io"
	"log"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a new ArduinoCoreServer
	srv := commands.NewArduinoCoreServer()

	// Disable logging
	logrus.SetOutput(io.Discard)

	// Create a new instance in the server
	ctx := context.Background()
	resp, err := srv.Create(ctx, &rpc.CreateRequest{})
	if err != nil {
		log.Fatal("Error creating instance:", err)
	}
	instance := resp.GetInstance()

	// Defer the destruction of the instance
	defer func() {
		if _, err := srv.Destroy(ctx, &rpc.DestroyRequest{Instance: instance}); err != nil {
			log.Fatal("Error destroying instance:", err)
		}
		fmt.Println("Instance successfully destroyed")
	}()

	// Initialize the instance
	initStream := commands.InitStreamResponseToCallbackFunction(ctx, func(r *rpc.InitResponse) error {
		fmt.Println("INIT> ", r)
		return nil
	})
	if err := srv.Init(&rpc.InitRequest{Instance: instance}, initStream); err != nil {
		log.Fatal("Error during initialization:", err)
	}

	// Search for platforms and output the result
	searchResp, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{Instance: instance})
	if err != nil {
		log.Fatal("Error searching for platforms:", err)
	}
	for _, platformSummary := range searchResp.GetSearchOutput() {
		installed := platformSummary.GetInstalledRelease()
		meta := platformSummary.GetMetadata()
		fmt.Printf("%30s %8s %s\n", meta.GetId(), installed.GetVersion(), installed.GetName())
	}
}
