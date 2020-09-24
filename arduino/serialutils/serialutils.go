// This file is part of arduino-cli
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

package serialutils

import (
	"time"

	"github.com/pkg/errors"
	"go.bug.st/serial"
)

// Reset a board using the 1200 bps port-touch. If wait is true, it will wait
// for a new port to appear (which could change sometimes) and returns that.
// The error is set if the port listing fails.
func Reset(port string, wait bool) (string, error) {
	// Touch port at 1200bps
	if err := TouchSerialPortAt1200bps(port); err != nil {
		return "", errors.New("1200bps Touch")
	}

	if wait {
		// Wait for port to disappear and reappear
		if p, err := WaitForNewSerialPortOrDefaultTo(port); err == nil {
			port = p
		} else {
			return "", errors.WithMessage(err, "detecting upload port")
		}
	}

	return port, nil
}

// TouchSerialPortAt1200bps open and close the serial port at 1200 bps. This
// is used on many Arduino boards as a signal to put the board in "bootloader"
// mode.
func TouchSerialPortAt1200bps(port string) error {
	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return errors.WithMessage(err, "opening port at 1200bps")
	}

	// Set DTR to false
	if err = p.SetDTR(false); err != nil {
		p.Close()
		return errors.WithMessage(err, "setting DTR to OFF")
	}

	// Close serial port
	p.Close()

	// Scanning for available ports seems to open the port or
	// otherwise assert DTR, which would cancel the WDT reset if
	// it happens within 250 ms. So we wait until the reset should
	// have already occurred before going on.
	time.Sleep(500 * time.Millisecond)

	return nil
}

// WaitForNewSerialPortOrDefaultTo is meant to be called just after a reset. It watches the ports connected
// to the machine until a port appears. The new appeared port is returned or, if the operation
// timeouts, the default port provided as parameter is returned.
func WaitForNewSerialPortOrDefaultTo(defaultPort string) (string, error) {
	if p, err := WaitForNewSerialPort(); err != nil {
		return "", errors.WithMessage(err, "detecting upload port")
	} else if p != "" {
		// on OS X, if the port is opened too quickly after it is detected,
		// a "Resource busy" error occurs, add a delay to workaround.
		// This apply to other platforms as well.
		time.Sleep(500 * time.Millisecond)

		return p, nil
	}
	return defaultPort, nil
}

// WaitForNewSerialPort is meant to be called just after a reset. It watches the ports connected
// to the machine until a port appears. The new appeared port is returned.
func WaitForNewSerialPort() (string, error) {
	getPortMap := func() (map[string]bool, error) {
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, errors.WithMessage(err, "listing serial ports")
		}
		res := map[string]bool{}
		for _, port := range ports {
			res[port] = true
		}
		return res, nil
	}

	last, err := getPortMap()
	if err != nil {
		return "", err
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		now, err := getPortMap()
		if err != nil {
			return "", err
		}

		for p := range now {
			if !last[p] {
				return p, nil // Found it!
			}
		}

		last = now
		time.Sleep(250 * time.Millisecond)
	}

	return "", nil
}
