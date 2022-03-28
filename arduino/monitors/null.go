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

package monitors

import (
	"log"
	"time"
)

// NullMonitor outputs zeros at a constant rate and discards anything sent
type NullMonitor struct {
	started time.Time
	sent    int
	bps     float64
}

// OpenNullMonitor creates a monitor that outputs the same character at a fixed
// rate.
func OpenNullMonitor(bytesPerSecondRate float64) *NullMonitor {
	log.Printf("Started streaming at %f\n", bytesPerSecondRate)
	return &NullMonitor{
		started: time.Now(),
		bps:     bytesPerSecondRate,
	}
}

// Close the connection
func (mon *NullMonitor) Close() error {
	return nil
}

// Read bytes from the port
func (mon *NullMonitor) Read(bytes []byte) (int, error) {
	for {
		elapsed := time.Since(mon.started).Seconds()
		n := int(elapsed*mon.bps) - mon.sent
		if n == 0 {
			// Delay until the next char...
			time.Sleep(time.Millisecond)
			continue
		}
		if len(bytes) < n {
			n = len(bytes)
		}
		mon.sent += n
		for i := 0; i < n; i++ {
			bytes[i] = 0
		}
		return n, nil
	}
}

// Write bytes to the port
func (mon *NullMonitor) Write(bytes []byte) (int, error) {
	// Discard all chars
	return len(bytes), nil
}
