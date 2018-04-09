//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

/*
Package serial is a cross-platform serial library for the go language.

The canonical import for this library is go.bug.st/serial.v1 so the import line
is the following:

	import "go.bug.st/serial.v1"

It is possible to get the list of available serial ports with the
GetPortsList function:

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

The serial port can be opened with the Open function:

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Fatal(err)
	}

The Open function needs a "mode" parameter that specifies the configuration
options for the serial port. If not specified the default options are 9600_N81,
in the example above only the speed is changed so the port is opened using 115200_N81.
The following snippets shows how to declare a configuration for 57600_E71:

	mode := &serial.Mode{
		BaudRate: 57600,
		Parity: serial.EvenParity,
		DataBits: 7,
		StopBits: serial.OneStopBit,
	}

The configuration can be changed at any time with the SetMode function:

	err := port.SetMode(mode)
	if err != nil {
		log.Fatal(err)
	}

The port object implements the io.ReadWriteCloser interface, so we can use
the usual Read, Write and Close functions to send and receive data from the
serial port:

	n, err := port.Write([]byte("10,20,30\n\r"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	buff := make([]byte, 100)
	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			fmt.Println("\nEOF")
			break
		}
		fmt.Printf("%v", string(buff[:n]))
	}

If a port is a virtual USB-CDC serial port (for example an USB-to-RS232
cable or a microcontroller development board) is possible to retrieve
the USB metadata, like VID/PID or USB Serial Number, with the
GetDetailedPortsList function in the enumerator package:

	import "go.bug.st/serial.v1/enumerator"

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		return
	}
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}

for details on USB port enumeration see the documentation of the specific package.

This library tries to avoid the use of the "C" package (and consequently the need
of cgo) to simplify cross compiling.
Unfortunately the USB enumeration package for darwin (MacOSX) requires cgo
to access the IOKit framework. This means that if you need USB enumeration
on darwin you're forced to use cgo.
*/
package serial // import "go.bug.st/serial.v1"
