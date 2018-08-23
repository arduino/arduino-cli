//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package enumerator // import "go.bug.st/serial.v1/enumerator"

//go:generate go run $GOROOT/src/syscall/mksyscall_windows.go -output syscall_windows.go usb_windows.go

// PortDetails contains detailed information about USB serial port.
// Use GetDetailedPortsList function to retrieve it.
type PortDetails struct {
	Name         string
	IsUSB        bool
	VID          string
	PID          string
	SerialNumber string

	// Manufacturer string
	// Product      string
}

// GetDetailedPortsList retrieve ports details like USB VID/PID.
// Please note that this function may not be available on all OS:
// in that case a FunctionNotImplemented error is returned.
func GetDetailedPortsList() ([]*PortDetails, error) {
	return nativeGetDetailedPortsList()
}

// PortEnumerationError is the error type for serial ports enumeration
type PortEnumerationError struct {
	causedBy error
}

// Error returns the complete error code with details on the cause of the error
func (e PortEnumerationError) Error() string {
	reason := "Error while enumerating serial ports"
	if e.causedBy != nil {
		reason += ": " + e.causedBy.Error()
	}
	return reason
}
