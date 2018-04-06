//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package serial // import "go.bug.st/serial.v1"

/*

// MSDN article on Serial Communications:
// http://msdn.microsoft.com/en-us/library/ff802693.aspx
// (alternative link) https://msdn.microsoft.com/en-us/library/ms810467.aspx

// Arduino Playground article on serial communication with Windows API:
// http://playground.arduino.cc/Interfacing/CPPWindows

*/

import "syscall"

type windowsPort struct {
	handle syscall.Handle
}

func nativeGetPortsList() ([]string, error) {
	subKey, err := syscall.UTF16PtrFromString("HARDWARE\\DEVICEMAP\\SERIALCOMM\\")
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts}
	}

	var h syscall.Handle
	if syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, subKey, 0, syscall.KEY_READ, &h) != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts}
	}
	defer syscall.RegCloseKey(h)

	var valuesCount uint32
	if syscall.RegQueryInfoKey(h, nil, nil, nil, nil, nil, nil, &valuesCount, nil, nil, nil, nil) != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts}
	}

	list := make([]string, valuesCount)
	for i := range list {
		var data [1024]uint16
		dataSize := uint32(len(data))
		var name [1024]uint16
		nameSize := uint32(len(name))
		if regEnumValue(h, uint32(i), &name[0], &nameSize, nil, nil, &data[0], &dataSize) != nil {
			return nil, &PortError{code: ErrorEnumeratingPorts}
		}
		list[i] = syscall.UTF16ToString(data[:])
	}
	return list, nil
}

func (port *windowsPort) Close() error {
	return syscall.CloseHandle(port.handle)
}

func (port *windowsPort) Read(p []byte) (int, error) {
	var readed uint32
	params := &dcb{}
	ev, err := createOverlappedEvent()
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(ev.HEvent)
	for {
		err := syscall.ReadFile(port.handle, p, &readed, ev)
		switch err {
		case nil:
			// operation completed successfully
		case syscall.ERROR_IO_PENDING:
			// wait for overlapped I/O to complete
			if err := getOverlappedResult(port.handle, ev, &readed, true); err != nil {
				return int(readed), err
			}
		default:
			// error happened
			return int(readed), err
		}

		if readed > 0 {
			return int(readed), nil
		}
		if err := resetEvent(ev.HEvent); err != nil {
			return 0, err
		}

		// At the moment it seems that the only reliable way to check if
		// a serial port is alive in Windows is to check if the SetCommState
		// function fails.

		getCommState(port.handle, params)
		if err := setCommState(port.handle, params); err != nil {
			port.Close()
			return 0, err
		}
	}
}

func (port *windowsPort) Write(p []byte) (int, error) {
	var writed uint32
	ev, err := createOverlappedEvent()
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(ev.HEvent)
	err = syscall.WriteFile(port.handle, p, &writed, ev)
	if err == syscall.ERROR_IO_PENDING {
		// wait for write to complete
		err = getOverlappedResult(port.handle, ev, &writed, true)
	}
	return int(writed), err
}

const (
	purgeRxAbort uint32 = 0x0002
	purgeRxClear        = 0x0008
	purgeTxAbort        = 0x0001
	purgeTxClear        = 0x0004
)

func (port *windowsPort) ResetInputBuffer() error {
	return purgeComm(port.handle, purgeRxClear|purgeRxAbort)
}

func (port *windowsPort) ResetOutputBuffer() error {
	return purgeComm(port.handle, purgeTxClear|purgeTxAbort)
}

const (
	dcbBinary                uint32 = 0x00000001
	dcbParity                       = 0x00000002
	dcbOutXCTSFlow                  = 0x00000004
	dcbOutXDSRFlow                  = 0x00000008
	dcbDTRControlDisableMask        = ^uint32(0x00000030)
	dcbDTRControlEnable             = 0x00000010
	dcbDTRControlHandshake          = 0x00000020
	dcbDSRSensitivity               = 0x00000040
	dcbTXContinueOnXOFF             = 0x00000080
	dcbOutX                         = 0x00000100
	dcbInX                          = 0x00000200
	dcbErrorChar                    = 0x00000400
	dcbNull                         = 0x00000800
	dcbRTSControlDisbaleMask        = ^uint32(0x00003000)
	dcbRTSControlEnable             = 0x00001000
	dcbRTSControlHandshake          = 0x00002000
	dcbRTSControlToggle             = 0x00003000
	dcbAbortOnError                 = 0x00004000
)

type dcb struct {
	DCBlength uint32
	BaudRate  uint32

	// Flags field is a bitfield
	//  fBinary            :1
	//  fParity            :1
	//  fOutxCtsFlow       :1
	//  fOutxDsrFlow       :1
	//  fDtrControl        :2
	//  fDsrSensitivity    :1
	//  fTXContinueOnXoff  :1
	//  fOutX              :1
	//  fInX               :1
	//  fErrorChar         :1
	//  fNull              :1
	//  fRtsControl        :2
	//  fAbortOnError      :1
	//  fDummy2            :17
	Flags uint32

	wReserved  uint16
	XonLim     uint16
	XoffLim    uint16
	ByteSize   byte
	Parity     byte
	StopBits   byte
	XonChar    byte
	XoffChar   byte
	ErrorChar  byte
	EOFChar    byte
	EvtChar    byte
	wReserved1 uint16
}

type commTimeouts struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

const (
	noParity    = 0
	oddParity   = 1
	evenParity  = 2
	markParity  = 3
	spaceParity = 4
)

var parityMap = map[Parity]byte{
	NoParity:    noParity,
	OddParity:   oddParity,
	EvenParity:  evenParity,
	MarkParity:  markParity,
	SpaceParity: spaceParity,
}

const (
	oneStopBit   = 0
	one5StopBits = 1
	twoStopBits  = 2
)

var stopBitsMap = map[StopBits]byte{
	OneStopBit:           oneStopBit,
	OnePointFiveStopBits: one5StopBits,
	TwoStopBits:          twoStopBits,
}

const (
	commFunctionSetXOFF  = 1
	commFunctionSetXON   = 2
	commFunctionSetRTS   = 3
	commFunctionClrRTS   = 4
	commFunctionSetDTR   = 5
	commFunctionClrDTR   = 6
	commFunctionSetBreak = 8
	commFunctionClrBreak = 9
)

const (
	msCTSOn  = 0x0010
	msDSROn  = 0x0020
	msRingOn = 0x0040
	msRLSDOn = 0x0080
)

func (port *windowsPort) SetMode(mode *Mode) error {
	params := dcb{}
	if getCommState(port.handle, &params) != nil {
		port.Close()
		return &PortError{code: InvalidSerialPort}
	}
	if mode.BaudRate == 0 {
		params.BaudRate = 9600 // Default to 9600
	} else {
		params.BaudRate = uint32(mode.BaudRate)
	}
	if mode.DataBits == 0 {
		params.ByteSize = 8 // Default to 8 bits
	} else {
		params.ByteSize = byte(mode.DataBits)
	}
	params.StopBits = stopBitsMap[mode.StopBits]
	params.Parity = parityMap[mode.Parity]
	if setCommState(port.handle, &params) != nil {
		port.Close()
		return &PortError{code: InvalidSerialPort}
	}
	return nil
}

func (port *windowsPort) SetDTR(dtr bool) error {
	// Like for RTS there are problems with the escapeCommFunction
	// observed behaviour was that DTR is set from false -> true
	// when setting RTS from true -> false
	// 1) Connect 		-> RTS = true 	(low) 	DTR = true 	(low) 	OKAY
	// 2) SetDTR(false) -> RTS = true 	(low) 	DTR = false (heigh)	OKAY
	// 3) SetRTS(false)	-> RTS = false 	(heigh)	DTR = true 	(low) 	ERROR: DTR toggled
	//
	// In addition this way the CommState Flags are not updated
	/*
		var res bool
		if dtr {
			res = escapeCommFunction(port.handle, commFunctionSetDTR)
		} else {
			res = escapeCommFunction(port.handle, commFunctionClrDTR)
		}
		if !res {
			return &PortError{}
		}
		return nil
	*/

	// The following seems a more reliable way to do it

	params := &dcb{}
	if err := getCommState(port.handle, params); err != nil {
		return &PortError{causedBy: err}
	}
	params.Flags &= dcbDTRControlDisableMask
	if dtr {
		params.Flags |= dcbDTRControlEnable
	}
	if err := setCommState(port.handle, params); err != nil {
		return &PortError{causedBy: err}
	}

	return nil
}

func (port *windowsPort) SetRTS(rts bool) error {
	// It seems that there is a bug in the Windows VCP driver:
	// it doesn't send USB control message when the RTS bit is
	// changed, so the following code not always works with
	// USB-to-serial adapters.
	//
	// In addition this way the CommState Flags are not updated

	/*
		var res bool
		if rts {
			res = escapeCommFunction(port.handle, commFunctionSetRTS)
		} else {
			res = escapeCommFunction(port.handle, commFunctionClrRTS)
		}
		if !res {
			return &PortError{}
		}
		return nil
	*/

	// The following seems a more reliable way to do it

	params := &dcb{}
	if err := getCommState(port.handle, params); err != nil {
		return &PortError{causedBy: err}
	}
	params.Flags &= dcbRTSControlDisbaleMask
	if rts {
		params.Flags |= dcbRTSControlEnable
	}
	if err := setCommState(port.handle, params); err != nil {
		return &PortError{causedBy: err}
	}
	return nil
}

func (port *windowsPort) GetModemStatusBits() (*ModemStatusBits, error) {
	var bits uint32
	if !getCommModemStatus(port.handle, &bits) {
		return nil, &PortError{}
	}
	return &ModemStatusBits{
		CTS: (bits & msCTSOn) != 0,
		DCD: (bits & msRLSDOn) != 0,
		DSR: (bits & msDSROn) != 0,
		RI:  (bits & msRingOn) != 0,
	}, nil
}

func createOverlappedEvent() (*syscall.Overlapped, error) {
	h, err := createEvent(nil, true, false, nil)
	return &syscall.Overlapped{HEvent: h}, err
}

func nativeOpen(portName string, mode *Mode) (*windowsPort, error) {
	portName = "\\\\.\\" + portName
	path, err := syscall.UTF16PtrFromString(portName)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(
		path,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0, nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_OVERLAPPED,
		0)
	if err != nil {
		switch err {
		case syscall.ERROR_ACCESS_DENIED:
			return nil, &PortError{code: PortBusy}
		case syscall.ERROR_FILE_NOT_FOUND:
			return nil, &PortError{code: PortNotFound}
		}
		return nil, err
	}
	// Create the serial port
	port := &windowsPort{
		handle: handle,
	}

	// Set port parameters
	if port.SetMode(mode) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	params := &dcb{}
	if getCommState(port.handle, params) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}
	params.Flags &= dcbRTSControlDisbaleMask
	params.Flags |= dcbRTSControlEnable
	params.Flags &= dcbDTRControlDisableMask
	params.Flags |= dcbDTRControlEnable
	params.Flags &^= dcbOutXCTSFlow
	params.Flags &^= dcbOutXDSRFlow
	params.Flags &^= dcbDSRSensitivity
	params.Flags |= dcbTXContinueOnXOFF
	params.Flags &^= dcbInX
	params.Flags &^= dcbOutX
	params.Flags &^= dcbErrorChar
	params.Flags &^= dcbNull
	params.Flags &^= dcbAbortOnError
	params.XonLim = 2048
	params.XoffLim = 512
	params.XonChar = 17  // DC1
	params.XoffChar = 19 // C3
	if setCommState(port.handle, params) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	// Set timeouts to 1 second
	timeouts := &commTimeouts{
		ReadIntervalTimeout:         0xFFFFFFFF,
		ReadTotalTimeoutMultiplier:  0xFFFFFFFF,
		ReadTotalTimeoutConstant:    1000, // 1 sec
		WriteTotalTimeoutConstant:   0,
		WriteTotalTimeoutMultiplier: 0,
	}
	if setCommTimeouts(port.handle, timeouts) != nil {
		port.Close()
		return nil, &PortError{code: InvalidSerialPort}
	}

	return port, nil
}
