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

// discovery_client is a command line UI client to test pluggable discoveries.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/internal/arduino/discovery"
	"github.com/arduino/arduino-cli/internal/arduino/discovery/discoverymanager"
	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one discovery.")
		os.Exit(1)
	}
	logrus.SetLevel(logrus.ErrorLevel)
	dm := discoverymanager.New()
	for _, discCmd := range os.Args[1:] {
		disc := discovery.New(discCmd, discCmd)
		dm.Add(disc)
	}
	dm.Start()

	watcher, err := dm.Watch()
	if err != nil {
		log.Fatalf("failed to start discoveries: %v", err)
	}

	for ev := range watcher.Feed() {
		port := ev.Port
		fmt.Printf("> Port %s\n", ev.Type)
		fmt.Printf("   Address: %s\n", port.Address)
		fmt.Printf("  Protocol: %s\n", port.Protocol)
		if ev.Type == "add" {
			if port.Properties != nil {
				keys := port.Properties.Keys()
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("            %s=%s\n", k, port.Properties.Get(k))
				}
			}
		}
		fmt.Println()
	}
}
