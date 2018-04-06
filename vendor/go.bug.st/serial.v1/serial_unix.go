//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build linux darwin freebsd openbsd

package serial // import "go.bug.st/serial.v1"

import (
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"

	"go.bug.st/serial.v1/unixutils"
)

type unixPort struct {
	handle int

	closeLock   sync.RWMutex
	closeSignal *unixutils.Pipe
	opened      bool
}

func (port *unixPort) Close() error {
	// Close port
	port.releaseExclusiveAccess()
	if err := unix.Close(port.handle); err != nil {
		return err
	}
	port.opened = false

	if port.closeSignal != nil {
		// Send close signal to all pending reads (if any)
		port.closeSignal.Write([]byte{0})

		// Wait for all readers to complete
		port.closeLock.Lock()
		defer port.closeLock.Unlock()

		// Close signaling pipe
		if err := port.closeSignal.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (port *unixPort) Read(p []byte) (n int, err error) {
	port.closeLock.RLock()
	defer port.closeLock.RUnlock()
	if !port.opened {
		return 0, &PortError{code: PortClosed}
	}

	fds := unixutils.NewFDSet(port.handle, port.closeSignal.ReadFD())
	res, err := unixutils.Select(fds, nil, fds, -1)
	if err != nil {
		return 0, err
	}
	if res.IsReadable(port.closeSignal.ReadFD()) {
		return 0, &PortError{code: PortClosed}
	}
	return unix.Read(port.handle, p)
}

func (port *unixPort) Write(p []byte) (n int, err error) {
	return unix.Write(port.handle, p)
}

func (port *unixPort) ResetInputBuffer() error {
	return ioctl(port.handle, ioctlTcflsh, unix.TCIFLUSH)
}

func (port *unixPort) ResetOutputBuffer() error {
	return ioctl(port.handle, ioctlTcflsh, unix.TCOFLUSH)
}

func (port *unixPort) SetMode(mode *Mode) error {
	settings, err := port.getTermSettings()
	if err != nil {
		return err
	}
	if err := setTermSettingsBaudrate(mode.BaudRate, settings); err != nil {
		return err
	}
	if err := setTermSettingsParity(mode.Parity, settings); err != nil {
		return err
	}
	if err := setTermSettingsDataBits(mode.DataBits, settings); err != nil {
		return err
	}
	if err := setTermSettingsStopBits(mode.StopBits, settings); err != nil {
		return err
	}
	return port.setTermSettings(settings)
}

func (port *unixPort) SetDTR(dtr bool) error {
	status, err := port.getModemBitsStatus()
	if err != nil {
		return err
	}
	if dtr {
		status |= unix.TIOCM_DTR
	} else {
		status &^= unix.TIOCM_DTR
	}
	return port.setModemBitsStatus(status)
}

func (port *unixPort) SetRTS(rts bool) error {
	status, err := port.getModemBitsStatus()
	if err != nil {
		return err
	}
	if rts {
		status |= unix.TIOCM_RTS
	} else {
		status &^= unix.TIOCM_RTS
	}
	return port.setModemBitsStatus(status)
}

func (port *unixPort) GetModemStatusBits() (*ModemStatusBits, error) {
	status, err := port.getModemBitsStatus()
	if err != nil {
		return nil, err
	}
	return &ModemStatusBits{
		CTS: (status & unix.TIOCM_CTS) != 0,
		DCD: (status & unix.TIOCM_CD) != 0,
		DSR: (status & unix.TIOCM_DSR) != 0,
		RI:  (status & unix.TIOCM_RI) != 0,
	}, nil
}

func nativeOpen(portName string, mode *Mode) (*unixPort, error) {
	h, err := unix.Open(portName, unix.O_RDWR|unix.O_NOCTTY|unix.O_NDELAY, 0)
	if err != nil {
		switch err {
		case unix.EBUSY:
			return nil, &PortError{code: PortBusy}
		case unix.EACCES:
			return nil, &PortError{code: PermissionDenied}
		}
		return nil, err
	}
	port := &unixPort{
		handle: h,
		opened: true,
	}

	// Setup serial port
	if port.SetMode(mode) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	settings, err := port.getTermSettings()
	if err != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	// Set raw mode
	setRawMode(settings)

	// Explicitly disable RTS/CTS flow control
	setTermSettingsCtsRts(false, settings)

	if port.setTermSettings(settings) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	unix.SetNonblock(h, false)

	port.acquireExclusiveAccess()

	// This pipe is used as a signal to cancel blocking Read
	pipe := &unixutils.Pipe{}
	if err := pipe.Open(); err != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort, causedBy: err}
	}
	port.closeSignal = pipe

	return port, nil
}

func nativeGetPortsList() ([]string, error) {
	files, err := ioutil.ReadDir(devFolder)
	if err != nil {
		return nil, err
	}

	ports := make([]string, 0, len(files))
	for _, f := range files {
		// Skip folders
		if f.IsDir() {
			continue
		}

		// Keep only devices with the correct name
		match, err := regexp.MatchString(regexFilter, f.Name())
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}

		portName := devFolder + "/" + f.Name()

		// Check if serial port is real or is a placeholder serial port "ttySxx"
		if strings.HasPrefix(f.Name(), "ttyS") {
			port, err := nativeOpen(portName, &Mode{})
			if err != nil {
				serr, ok := err.(*PortError)
				if ok && serr.Code() == InvalidSerialPort {
					continue
				}
			} else {
				port.Close()
			}
		}

		// Save serial port in the resulting list
		ports = append(ports, portName)
	}

	return ports, nil
}

// termios manipulation functions

func setTermSettingsBaudrate(speed int, settings *unix.Termios) error {
	baudrate, ok := baudrateMap[speed]
	if !ok {
		return &PortError{code: InvalidSpeed}
	}
	// revert old baudrate
	for _, rate := range baudrateMap {
		settings.Cflag &^= rate
	}
	// set new baudrate
	settings.Cflag |= baudrate
	settings.Ispeed = toTermiosSpeedType(baudrate)
	settings.Ospeed = toTermiosSpeedType(baudrate)
	return nil
}

func setTermSettingsParity(parity Parity, settings *unix.Termios) error {
	switch parity {
	case NoParity:
		settings.Cflag &^= unix.PARENB
		settings.Cflag &^= unix.PARODD
		settings.Cflag &^= tcCMSPAR
		settings.Iflag &^= unix.INPCK
	case OddParity:
		settings.Cflag |= unix.PARENB
		settings.Cflag |= unix.PARODD
		settings.Cflag &^= tcCMSPAR
		settings.Iflag |= unix.INPCK
	case EvenParity:
		settings.Cflag |= unix.PARENB
		settings.Cflag &^= unix.PARODD
		settings.Cflag &^= tcCMSPAR
		settings.Iflag |= unix.INPCK
	case MarkParity:
		if tcCMSPAR == 0 {
			return &PortError{code: InvalidParity}
		}
		settings.Cflag |= unix.PARENB
		settings.Cflag |= unix.PARODD
		settings.Cflag |= tcCMSPAR
		settings.Iflag |= unix.INPCK
	case SpaceParity:
		if tcCMSPAR == 0 {
			return &PortError{code: InvalidParity}
		}
		settings.Cflag |= unix.PARENB
		settings.Cflag &^= unix.PARODD
		settings.Cflag |= tcCMSPAR
		settings.Iflag |= unix.INPCK
	default:
		return &PortError{code: InvalidParity}
	}
	return nil
}

func setTermSettingsDataBits(bits int, settings *unix.Termios) error {
	databits, ok := databitsMap[bits]
	if !ok {
		return &PortError{code: InvalidDataBits}
	}
	// Remove previous databits setting
	settings.Cflag &^= unix.CSIZE
	// Set requested databits
	settings.Cflag |= databits
	return nil
}

func setTermSettingsStopBits(bits StopBits, settings *unix.Termios) error {
	switch bits {
	case OneStopBit:
		settings.Cflag &^= unix.CSTOPB
	case OnePointFiveStopBits:
		return &PortError{code: InvalidStopBits}
	case TwoStopBits:
		settings.Cflag |= unix.CSTOPB
	default:
		return &PortError{code: InvalidStopBits}
	}
	return nil
}

func setTermSettingsCtsRts(enable bool, settings *unix.Termios) {
	if enable {
		settings.Cflag |= tcCRTSCTS
	} else {
		settings.Cflag &^= tcCRTSCTS
	}
}

func setRawMode(settings *unix.Termios) {
	// Set local mode
	settings.Cflag |= unix.CREAD
	settings.Cflag |= unix.CLOCAL

	// Set raw mode
	settings.Lflag &^= unix.ICANON
	settings.Lflag &^= unix.ECHO
	settings.Lflag &^= unix.ECHOE
	settings.Lflag &^= unix.ECHOK
	settings.Lflag &^= unix.ECHONL
	settings.Lflag &^= unix.ECHOCTL
	settings.Lflag &^= unix.ECHOPRT
	settings.Lflag &^= unix.ECHOKE
	settings.Lflag &^= unix.ISIG
	settings.Lflag &^= unix.IEXTEN

	settings.Iflag &^= unix.IXON
	settings.Iflag &^= unix.IXOFF
	settings.Iflag &^= unix.IXANY
	settings.Iflag &^= unix.INPCK
	settings.Iflag &^= unix.IGNPAR
	settings.Iflag &^= unix.PARMRK
	settings.Iflag &^= unix.ISTRIP
	settings.Iflag &^= unix.IGNBRK
	settings.Iflag &^= unix.BRKINT
	settings.Iflag &^= unix.INLCR
	settings.Iflag &^= unix.IGNCR
	settings.Iflag &^= unix.ICRNL
	settings.Iflag &^= tcIUCLC

	settings.Oflag &^= unix.OPOST

	// Block reads until at least one char is available (no timeout)
	settings.Cc[unix.VMIN] = 1
	settings.Cc[unix.VTIME] = 0
}

// native syscall wrapper functions

func (port *unixPort) getTermSettings() (*unix.Termios, error) {
	settings := &unix.Termios{}
	err := ioctl(port.handle, ioctlTcgetattr, uintptr(unsafe.Pointer(settings)))
	return settings, err
}

func (port *unixPort) setTermSettings(settings *unix.Termios) error {
	return ioctl(port.handle, ioctlTcsetattr, uintptr(unsafe.Pointer(settings)))
}

func (port *unixPort) getModemBitsStatus() (int, error) {
	var status int
	err := ioctl(port.handle, unix.TIOCMGET, uintptr(unsafe.Pointer(&status)))
	return status, err
}

func (port *unixPort) setModemBitsStatus(status int) error {
	return ioctl(port.handle, unix.TIOCMSET, uintptr(unsafe.Pointer(&status)))
}

func (port *unixPort) acquireExclusiveAccess() error {
	return ioctl(port.handle, unix.TIOCEXCL, 0)
}

func (port *unixPort) releaseExclusiveAccess() error {
	return ioctl(port.handle, unix.TIOCNXCL, 0)
}
