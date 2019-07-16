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
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package monitors

import (
	"github.com/pkg/errors"
	serial "go.bug.st/serial.v1"
)

// SerialMonitor is a monitor for serial ports
type SerialMonitor struct {
	port serial.Port
}

// OpenSerialMonitor creates a monitor instance for a serial port
func OpenSerialMonitor(portName string, baudRate int) (*SerialMonitor, error) {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, errors.Wrap(err, "error opening serial monitor")
	}

	return &SerialMonitor{
		port: port,
	}, nil
}

// Close the connection
func (mon *SerialMonitor) Close() error {
	return mon.port.Close()
}

// Read bytes from the port
func (mon *SerialMonitor) Read(bytes []byte) (int, error) {
	return mon.port.Read(bytes)
}

// Write bytes to the port
func (mon *SerialMonitor) Write(bytes []byte) (int, error) {
	return mon.port.Write(bytes)
}
