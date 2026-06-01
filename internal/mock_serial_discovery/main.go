//
// This file is part arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
)

type mockSerialDiscovery struct {
	startSyncCount int
	closeChan      chan<- bool
	tmpFile        string
}

func main() {
	// Write a file in a $TMP/mock_serial_discovery folder.
	// This file will be used by the integration tests to detect if the discovery is running.
	tmpDir := paths.TempDir().Join("mock_serial_discovery")
	if err := tmpDir.MkdirAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	tmpFile, err := paths.MkTempFile(tmpDir, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Run mock discovery
	dummy := &mockSerialDiscovery{tmpFile: tmpFile.Name()}
	server := discovery.NewServer(dummy)
	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		os.Exit(1)
	}
}

// Hello does nothing here...
func (d *mockSerialDiscovery) Hello(userAgent string, protocol int) error {
	return nil
}

// Quit does nothing here...
func (d *mockSerialDiscovery) Quit() {}

// Stop terminates the discovery loop
func (d *mockSerialDiscovery) Stop() error {
	if d.closeChan != nil {
		d.closeChan <- true
		close(d.closeChan)
		d.closeChan = nil
	}
	return nil
}

// StartSync starts the goroutine that generates fake Ports.
func (d *mockSerialDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	// Every 5 starts produce an error
	d.startSyncCount++
	if d.startSyncCount%5 == 0 {
		return errors.New("could not start_sync every 5 times")
	}

	c := make(chan bool)
	d.closeChan = c

	// Start asynchronous event emitter
	go func() {
		var closeChan <-chan bool = c

		// Output initial port state
		eventCB("add", &discovery.Port{
			Address:       "/dev/ttyCIAO",
			AddressLabel:  "Mocked Serial port",
			Protocol:      "serial",
			ProtocolLabel: "Serial",
			HardwareID:    "123456",
			Properties: properties.NewFromHashmap(map[string]string{
				"vid":           "0x2341",
				"pid":           "0x0041",
				"serial":        "123456",
				"discovery_tmp": d.tmpFile,
			}),
		})

		select {
		case <-closeChan:
			return
		case <-time.After(5 * time.Second):
			errorCB("unrecoverable error, cannot send more events")
		}

		<-closeChan
	}()

	return nil
}
