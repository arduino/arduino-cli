//
// This file is part of arduino-cli.
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
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/arduino/go-paths-helper"
	monitor "github.com/arduino/pluggable-monitor-protocol-handler"
)

func main() {
	monitorServer := monitor.NewServer(NewSerialMonitor())
	if err := monitorServer.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

// SerialMonitor is the implementation of the serial ports pluggable-monitor
type SerialMonitor struct {
	mockedSerialPort io.ReadWriteCloser
	serialSettings   *monitor.PortDescriptor
	openedPort       bool
	muxFile          *paths.Path
}

// NewSerialMonitor will initialize and return a SerialMonitor
func NewSerialMonitor() *SerialMonitor {
	return &SerialMonitor{
		serialSettings: &monitor.PortDescriptor{
			Protocol: "serial",
			ConfigurationParameter: map[string]*monitor.PortParameterDescriptor{
				"baudrate": {
					Label: "Baudrate",
					Type:  "enum",
					Values: []string{
						"300", "600", "750",
						"1200", "2400", "4800", "9600",
						"19200", "31250", "38400", "57600", "74880",
						"115200", "230400", "250000", "460800", "500000", "921600",
						"1000000", "2000000"},
					Selected: "9600",
				},
				"parity": {
					Label:    "Parity",
					Type:     "enum",
					Values:   []string{"none", "even", "odd", "mark", "space"},
					Selected: "none",
				},
				"bits": {
					Label:    "Data bits",
					Type:     "enum",
					Values:   []string{"5", "6", "7", "8", "9"},
					Selected: "8",
				},
				"stop_bits": {
					Label:    "Stop bits",
					Type:     "enum",
					Values:   []string{"1", "1.5", "2"},
					Selected: "1",
				},
				"rts": {
					Label:    "RTS",
					Type:     "enum",
					Values:   []string{"on", "off"},
					Selected: "on",
				},
				"dtr": {
					Label:    "DTR",
					Type:     "enum",
					Values:   []string{"on", "off"},
					Selected: "on",
				},
			},
		},
		openedPort: false,
	}
}

// Hello is the handler for the pluggable-monitor HELLO command
func (d *SerialMonitor) Hello(userAgent string, protocol int) error {
	return nil
}

// Describe is the handler for the pluggable-monitor DESCRIBE command
func (d *SerialMonitor) Describe() (*monitor.PortDescriptor, error) {
	return d.serialSettings, nil
}

// Configure is the handler for the pluggable-monitor CONFIGURE command
func (d *SerialMonitor) Configure(parameterName string, value string) error {
	parameter, ok := d.serialSettings.ConfigurationParameter[parameterName]
	if !ok {
		return fmt.Errorf("could not find parameter named %s", parameterName)
	}
	if !slices.Contains(parameter.Values, value) {
		return fmt.Errorf("invalid value for parameter %s: %s", parameterName, value)
	}
	// Set configuration
	parameter.Selected = value

	return nil
}

// Open is the handler for the pluggable-monitor OPEN command
func (d *SerialMonitor) Open(boardPort string) (io.ReadWriter, error) {
	if d.openedPort {
		return nil, fmt.Errorf("port already opened: %s", boardPort)
	}
	d.openedPort = true
	sideA, sideB := newBidirectionalPipe()
	d.mockedSerialPort = sideA
	if muxFile, err := paths.MkTempFile(nil, ""); err == nil {
		d.muxFile = paths.NewFromFile(muxFile)
		muxFile.Close()
	}
	go func() {
		buff := make([]byte, 1024)
		d.mockedSerialPort.Write([]byte("Opened port: " + boardPort + "\n"))
		if d.muxFile != nil {
			d.mockedSerialPort.Write([]byte("Tmpfile: " + d.muxFile.String() + "\n"))
		}
		for parameter, descriptor := range d.serialSettings.ConfigurationParameter {
			d.mockedSerialPort.Write([]byte(
				fmt.Sprintf("Configuration %s = %s\n", parameter, descriptor.Selected)))
		}
		for {
			n, err := d.mockedSerialPort.Read(buff)
			if err != nil {
				d.mockedSerialPort.Close()
				return
			}
			if strings.Contains(string(buff[:n]), "QUIT") {
				d.mockedSerialPort.Close()
				return
			}
			d.mockedSerialPort.Write([]byte("Received: >"))
			d.mockedSerialPort.Write(buff[:n])
			d.mockedSerialPort.Write([]byte("<\n"))
		}
	}()
	return sideB, nil
}

func newBidirectionalPipe() (io.ReadWriteCloser, io.ReadWriteCloser) {
	in1, out1 := io.Pipe()
	in2, out2 := io.Pipe()
	a := &bidirectionalPipe{in: in1, out: out2}
	b := &bidirectionalPipe{in: in2, out: out1}
	return a, b
}

type bidirectionalPipe struct {
	in  io.Reader
	out io.WriteCloser
}

func (p *bidirectionalPipe) Read(buff []byte) (int, error) {
	return p.in.Read(buff)
}

func (p *bidirectionalPipe) Write(buff []byte) (int, error) {
	return p.out.Write(buff)
}

func (p *bidirectionalPipe) Close() error {
	return p.out.Close()
}

// Close is the handler for the pluggable-monitor CLOSE command
func (d *SerialMonitor) Close() error {
	if !d.openedPort {
		return errors.New("port already closed")
	}
	d.mockedSerialPort.Close()
	d.openedPort = false
	if d.muxFile != nil {
		time.Sleep(500 * time.Millisecond) // Emulate a small delay closing the port to check gRPC synchronization
		d.muxFile.Remove()
		d.muxFile = nil
	}
	return nil
}

// Quit is the handler for the pluggable-monitor QUIT command
func (d *SerialMonitor) Quit() {}
