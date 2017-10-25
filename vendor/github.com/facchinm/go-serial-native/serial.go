/*

Package serial provides a binding to libserialport for serial port
functionality. Serial ports are commonly used with embedded systems,
such as the Arduino platform.

Example Usage

	package main

	import (
	  "github.com/mikepb/go-serial"
	  "log"
	)

	func main() {
	  options := serial.RawOptions
	  options.BitRate = 115200
	  p, err := options.Open("/dev/tty")
	  if err != nil {
	    log.Panic(err)
	  }

	  defer p.Close()

	  buf := make([]byte, 1)
	  if c, err := p.Read(buf); err != nil {
	    log.Panic(err)
	  } else {
	    log.Println(buf)
	  }
	}

*/
package serial

/*
#cgo CFLAGS: -g -O2 -Wall -Wextra -Wno-unused-parameter -DSP_PRIV= -DSP_API=
#cgo darwin LDFLAGS: -framework IOKit -framework CoreFoundation
#cgo windows LDFLAGS: -lsetupapi

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include "libserialport.h"

void debug_handler(const char *fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vfprintf(stderr, fmt, args);
    va_end(args);
}

void setdebug(int enable) {
	sp_set_debug_handler(enable ? debug_handler : NULL);
}

*/
import "C"

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

// Debug flag
const Debug = false

// Port access modes
const (
	MODE_READ       = C.SP_MODE_READ       // Open port for read access
	MODE_WRITE      = C.SP_MODE_WRITE      // Open port for write access
	MODE_READ_WRITE = C.SP_MODE_READ_WRITE // Open port for read and write access.
)

// Port events.
const (
	EVENT_RX_READY = C.SP_EVENT_RX_READY // Data received and ready to read.
	EVENT_TX_READY = C.SP_EVENT_TX_READY // Ready to transmit new data.
	EVENT_ERROR    = C.SP_EVENT_ERROR    // Error occured.
)

// Parity settings.
const (
	PARITY_INVALID = iota // Special value to indicate setting should be left alone.
	PARITY_NONE           // No parity.
	PARITY_ODD            // Odd parity.
	PARITY_EVEN           // Even parity.
	PARITY_MARK           // Mark parity.
	PARITY_SPACE          // Space parity.
)

// RTS pin behaviour.
const (
	RTS_INVALID      = iota // Special value to indicate setting should be left alone.
	RTS_OFF                 // RTS off.
	RTS_ON                  // RTS on.
	RTS_FLOW_CONTROL        // RTS used for flow control.
)

// CTS pin behaviour.
const (
	CTS_INVALID      = iota // Special value to indicate setting should be left alone.
	CTS_IGNORE              // CTS ignored.
	CTS_FLOW_CONTROL        // CTS used for flow control.
)

// DTR pin behaviour.
const (
	DTR_INVALID      = iota // Special value to indicate setting should be left alone.
	DTR_OFF                 // DTR off.
	DTR_ON                  // DTR on.
	DTR_FLOW_CONTROL        // DTR used for flow control.
)

// DSR pin behaviour.
const (
	DSR_INVALID      = iota // Special value to indicate setting should be left alone.
	DSR_IGNORE              // DSR ignored.
	DSR_FLOW_CONTROL        // DSR used for flow control.
)

// XON/XOFF flow control behaviour.
const (
	XONXOFF_INVALID  = iota // Special value to indicate setting should be left alone.
	XONXOFF_DISABLED        // XON/XOFF disabled.
	XONXOFF_IN              // XON/XOFF enabled for input only.
	XONXOFF_OUT             // XON/XOFF enabled for output only.
	XONXOFF_INOUT           // XON/XOFF enabled for input and output.
)

// Standard flow control combinations.
const (
	_                   = iota
	FLOWCONTROL_NONE    // No flow control.
	FLOWCONTROL_XONXOFF // Software flow control using XON/XOFF characters.
	FLOWCONTROL_RTSCTS  // Hardware flow control using RTS/CTS signals.
	FLOWCONTROL_DTRDSR  // Hardware flow control using DTR/DSR signals.
)

// Input signals
const (
	SIG_CTS = C.SP_SIG_CTS // Clear to send
	SIG_DSR = C.SP_SIG_DSR // Data set ready
	SIG_DCD = C.SP_SIG_DCD // Data carrier detect
	SIG_RI  = C.SP_SIG_RI  // Ring indicator
)

// Transport types.
const (
	TRANSPORT_NATIVE    = C.SP_TRANSPORT_NATIVE    // Native platform serial port.
	TRANSPORT_USB       = C.SP_TRANSPORT_USB       // USB serial port adapter.
	TRANSPORT_BLUETOOTH = C.SP_TRANSPORT_BLUETOOTH // Bluetooh serial port adapter.
)

// Serial port info.
type Info struct {
	p      *C.struct_sp_port
	opened bool
}

// Serial port options.
type Options struct {
	Mode        int // read, write; default is read
	BitRate     int // number of bits per second (baudrate)
	DataBits    int // number of data bits (5, 6, 7, 8)
	StopBits    int // number of stop bits (1, 2)
	Parity      int // none, odd, even, mark, space
	FlowControl int // none, xonxoff, rtscts, dtrdsr

	RTS int
	CTS int
	DTR int
	DSR int
}

// Serial port.
type Port struct {
	Info
	c             *C.struct_sp_port_config
	readDeadline  time.Time
	writeDeadline time.Time
}

// Implementation of net.Addr
type Addr struct {
	name string
}

// Implementation of net.Error
type Error struct {
	msg       string
	timeout   bool
	temporary bool
}

var RawOptions = Options{
	DataBits:    8,
	Parity:      PARITY_NONE,
	StopBits:    1,
	FlowControl: FLOWCONTROL_NONE,
}

var ErrInvalidArguments = &Error{msg: "Invalid arguments were passed to the function"}
var ErrSystem = &Error{msg: "A system error occured while executing the operation"}
var ErrMemoryAllocation = &Error{msg: "A memory allocation failed while executing the operation"}
var ErrUnsupportedOperation = &Error{msg: "The requested operation is not supported by this system or device"}
var ErrTimeout = &Error{msg: "Operation timed out", timeout: true}

// Map error codes to errors.
func errmsg(err C.enum_sp_return) error {
	switch err {
	case C.SP_ERR_ARG:
		return ErrInvalidArguments
	case C.SP_ERR_FAIL:
		return ErrSystem
	case C.SP_ERR_MEM:
		return ErrMemoryAllocation
	case C.SP_ERR_SUPP:
		return ErrUnsupportedOperation
	}
	return nil
}

// Wrap a sp_port struct in a go Port struct and set finalizer for
// garbage collection.
func newInfo(p *C.struct_sp_port) (*Info, error) {
	info := &Info{p: p}
	runtime.SetFinalizer(info, (*Info).free)
	return info, nil
}

// Finalizer callback for garbage collection.
func (i *Info) free() {
	if i.p != nil {
		if i.opened {
			C.sp_close(i.p)
		}
		C.sp_free_port(i.p)
	}
	i.opened = false
	i.p = nil
}

// Wrap a sp_port struct in a go Port struct and set finalizer for
// garbage collection.
func newPort(info *Info) (*Port, error) {
	port := &Port{}

	// copy info
	if info != nil {
		if err := errmsg(C.sp_copy_port(info.p, &port.p)); err != nil {
			return nil, err
		}
	}

	// set finalizers
	runtime.SetFinalizer(port, (*Port).free)

	return port, nil
}

// Finalizer callback for garbage collection.
func (p *Port) free() {
	p.Info.free()
	if p.c != nil {
		C.sp_free_config(p.c)
	}
	p.c = nil
}

// calculate milliseconds until deadline (rounded up)
func deadline2millis(deadline time.Time) int64 {
	delta := deadline.Sub(time.Now())

	duration := time.Duration(delta.Nanoseconds())
	duration += duration + time.Millisecond - time.Nanosecond
	duration /= time.Millisecond

	millis := int64(duration)

	if Debug {
		log.Printf("timeout: %d ns %d ms", delta, millis)
	}

	return millis
}

// Print libserialport debug messages to stderr.
func SetDebug(enable bool) {
	if enable {
		C.setdebug(1)
	} else {
		C.setdebug(0)
	}
}

// Get a port by name.
func PortByName(name string) (*Info, error) {
	if p, err := portByName(name); err != nil {
		return nil, err
	} else {
		return newInfo(p)
	}
}
func portByName(name string) (*C.struct_sp_port, error) {
	var p *C.struct_sp_port

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if err := errmsg(C.sp_get_port_by_name(cname, &p)); err != nil {
		return nil, err
	}

	return p, nil
}

// List the serial ports available on the system.
func ListPorts() ([]*Info, error) {
	var p **C.struct_sp_port

	if err := C.sp_list_ports(&p); err != C.SP_OK {
		return nil, errmsg(err)
	}
	defer C.sp_free_port_list(p)

	// Convert the C array into a Go slice
	// See: https://code.google.com/p/go-wiki/wiki/cgo
	pp := (*[1 << 15]*C.struct_sp_port)(unsafe.Pointer(p))

	// count number of ports
	c := 0
	for ; uintptr(unsafe.Pointer(pp[c])) != 0; c++ {
	}

	// populate
	ports := make([]*Info, c)
	for j := 0; j < c; j++ {
		var pc *C.struct_sp_port
		if err := errmsg(C.sp_copy_port(pp[j], &pc)); err != nil {
			return nil, err
		}
		if sp, err := newInfo(pc); err != nil {
			return nil, err
		} else {
			ports[j] = sp
		}
	}

	return ports, nil
}

// Get the name of a port.
func (i *Info) Name() string {
	return C.GoString(C.sp_get_port_name(i.p))
}

// Get a description for a port, to present to end user.
func (i *Info) Description() string {
	return C.GoString(C.sp_get_port_description(i.p))
}

// Get the transport type used by a port.
func (i *Info) Transport() int {
	t := C.sp_get_port_transport(i.p)
	return int(t)
}

// Get the USB bus number and address on bus of a USB serial adapter port.
func (i *Info) USBBusAddress() (int, int, error) {
	var bus, address C.int
	if err := errmsg(C.sp_get_port_usb_bus_address(i.p, &bus, &address)); err != nil {
		return 0, 0, err
	}
	return int(bus), int(address), nil
}

// Get the USB Vendor ID and Product ID of a USB serial adapter port.
func (i *Info) USBVIDPID() (int, int, error) {
	var vid, pid C.int
	if err := errmsg(C.sp_get_port_usb_vid_pid(i.p, &vid, &pid)); err != nil {
		return 0, 0, err
	}
	return int(vid), int(pid), nil
}

// Get the USB manufacturer string of a USB serial adapter port.
func (i *Info) USBManufacturer() string {
	cdesc := C.sp_get_port_usb_manufacturer(i.p)
	return C.GoString(cdesc)
}

// Get the USB product string of a USB serial adapter port.
func (i *Info) USBProduct() string {
	cdesc := C.sp_get_port_usb_product(i.p)
	return C.GoString(cdesc)
}

// Get the USB serial number string of a USB serial adapter port.
func (i *Info) USBSerialNumber() string {
	cdesc := C.sp_get_port_usb_serial(i.p)
	return C.GoString(cdesc)
}

// Get the MAC address of a Bluetooth serial adapter port.
func (i *Info) BluetoothAddress() string {
	cdesc := C.sp_get_port_bluetooth_address(i.p)
	return C.GoString(cdesc)
}

// Open the port for reading and default options.
func (i *Info) Open() (*Port, error) {
	return i.OpenPort(&Options{Mode: MODE_READ})
}

// Open the port with the specified options.
func (i *Info) OpenPort(options *Options) (*Port, error) {
	// create port
	port, err := newPort(i)
	if err != nil {
		return nil, err
	}

	// open port
	mode := MODE_READ
	if options.Mode != 0 {
		mode = options.Mode
	}
	if err = port.open(mode); err != nil {
		return nil, err
	}

	// apply options
	if err = port.Apply(options); err != nil {
		port.Close()
		return nil, err
	}

	return port, nil
}

func (i *Info) createPortAndInvalidateInfo() (*Port, error) {
	port, err := newPort(nil)
	if err != nil {
		return nil, err
	}

	port.p = i.p
	port.opened = i.opened

	i.p = nil
	i.opened = false

	return port, nil
}

// Open a port at the given name using the options object.
func (o *Options) Open(name string) (port *Port, err error) {
	if info, err := PortByName(name); err != nil {
		return nil, err
	} else {
		return info.OpenPort(o)
	}
}

// Open a port at the given info using the options object.
func (o *Options) OpenAt(info *Info) (port *Port, err error) {
	return info.OpenPort(o)
}

// Open a port for reading.
func Open(name string) (port *Port, err error) {
	// get the port by name
	if info, err := PortByName(name); err != nil {
		return nil, err
	} else if p, err := info.createPortAndInvalidateInfo(); p != nil {
		return nil, err
	} else {
		port = p
	}

	// open port with read mode
	if err = port.open(MODE_READ); err != nil {
		return nil, err
	}

	return
}

func (p *Port) open(mode int) error {
	if p.opened {
		panic("already opened")
	}
	err := errmsg(C.sp_open(p.p, C.enum_sp_mode(mode)))
	p.opened = err == nil
	return p.getConf()
}

// Close the serial port.
func (p *Port) Close() error {
	if !p.opened {
		panic("already closed")
	}
	err := errmsg(C.sp_close(p.p))
	p.opened = false
	return err
}

func (p *Port) getConf() error {
	if p.c == nil {
		if err := errmsg(C.sp_new_config(&p.c)); err != nil {
			return err
		}
	}
	return errmsg(C.sp_get_config(p.p, p.c))
}

// Apply port options.
func (p *Port) Apply(o *Options) (err error) {
	// get port config
	var conf *C.struct_sp_port_config
	if err = errmsg(C.sp_new_config(&conf)); err != nil {
		return
	}
	defer C.sp_free_config(conf)

	// set bit rate
	if o.BitRate != 0 {
		err = errmsg(C.sp_set_config_baudrate(conf, C.int(o.BitRate)))
		if err != nil {
			return
		}
	}

	// set data bits
	if o.DataBits != 0 {
		err = errmsg(C.sp_set_config_bits(conf, C.int(o.DataBits)))
		if err != nil {
			return
		}
	}

	// set stop bits
	if o.StopBits != 0 {
		err = errmsg(C.sp_set_config_stopbits(conf, C.int(o.StopBits)))
		if err != nil {
			return
		}
	}

	// set parity
	if o.Parity != 0 {
		cparity := parity2c(o.Parity)
		if err = errmsg(C.sp_set_config_parity(conf, cparity)); err != nil {
			return
		}
	}

	// set flow control
	if o.FlowControl != 0 {
		cfc, err := flow2c(o.FlowControl)
		if err != nil {
			return err
		}
		if err = errmsg(C.sp_set_config_flowcontrol(conf, cfc)); err != nil {
			return err
		}
	}

	// set RTS
	if o.RTS != 0 {
		crts := rts2c(o.RTS)
		if err = errmsg(C.sp_set_config_rts(conf, crts)); err != nil {
			return
		}
	}

	// set CTS
	if o.CTS != 0 {
		ccts := cts2c(o.CTS)
		if err = errmsg(C.sp_set_config_cts(conf, ccts)); err != nil {
			return
		}
	}

	// set DTR
	if o.DTR != 0 {
		cdtr := dtr2c(o.DTR)
		if err = errmsg(C.sp_set_config_dtr(conf, cdtr)); err != nil {
			return
		}
	}

	// set DSR
	if o.DSR != 0 {
		cdsr := dsr2c(o.DSR)
		if err = errmsg(C.sp_set_config_dsr(conf, cdsr)); err != nil {
			return
		}
	}

	// apply config
	if err = errmsg(C.sp_set_config(p.p, conf)); err != nil {
		return
	}

	// update local config
	return p.getConf()
}

// Get the baud rate from a port configuration. The port must be
// opened for this operation.
func (p *Port) BitRate() (int, error) {
	var bitrate C.int
	if err := errmsg(C.sp_get_config_baudrate(p.c, &bitrate)); err != nil {
		return 0, err
	}
	return int(bitrate), nil
}

// Set the baud rate for the serial port. The port must be opened for
// this operation. Call p.ApplyConfig() to apply the change.
func (p *Port) SetBitRate(bitrate int) error {
	if err := errmsg(C.sp_set_baudrate(p.p, C.int(bitrate))); err != nil {
		return err
	}
	return p.getConf()
}

// Get the data bits from a port configuration. The port must be
// opened for this operation.
func (p *Port) DataBits() (int, error) {
	var bits C.int
	if err := errmsg(C.sp_get_config_bits(p.c, &bits)); err != nil {
		return 0, err
	}
	return int(bits), nil
}

// Set the number of data bits for the serial port. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetDataBits(bits int) error {
	if err := errmsg(C.sp_set_config_bits(p.c, C.int(bits))); err != nil {
		return err
	}
	return p.getConf()
}

// Get the stop bits from a port configuration. The port must be
// opened for this operation.
func (p *Port) StopBits() (int, error) {
	var stopbits C.int
	if err := errmsg(C.sp_get_config_stopbits(p.c, &stopbits)); err != nil {
		return 0, err
	}
	return int(stopbits), nil
}

// Set the stop bits for the serial port. The port must be opened for
// this operation. Call p.ApplyConfig() to apply the change.
func (p *Port) SetStopBits(stopbits int) error {
	if err := errmsg(C.sp_set_config_stopbits(p.c, C.int(stopbits))); err != nil {
		return err
	}
	return p.getConf()
}

// Get the parity setting from a port configuration. The port must be
// opened for this operation.
func (p *Port) Parity() (int, error) {
	cparity := C.enum_sp_parity(C.SP_PARITY_INVALID)
	if err := errmsg(C.sp_get_config_parity(p.c, &cparity)); err != nil {
		return 0, err
	}
	return c2parity(cparity), nil
}

// Set the parity setting for the serial port. The port must be opened
// for this operation. Call p.ApplyConfig() to apply the change.
func (p *Port) SetParity(parity int) error {
	cparity := parity2c(parity)
	if err := errmsg(C.sp_set_config_parity(p.c, cparity)); err != nil {
		return err
	}
	return p.getConf()
}

func c2parity(cparity C.enum_sp_parity) int {
	switch cparity {
	case C.SP_PARITY_NONE:
		return PARITY_NONE
	case C.SP_PARITY_ODD:
		return PARITY_ODD
	case C.SP_PARITY_EVEN:
		return PARITY_EVEN
	case C.SP_PARITY_MARK:
		return PARITY_MARK
	case C.SP_PARITY_SPACE:
		return PARITY_SPACE
	default:
		return PARITY_INVALID
	}
}

func parity2c(parity int) C.enum_sp_parity {
	switch parity {
	case PARITY_NONE:
		return C.SP_PARITY_NONE
	case PARITY_ODD:
		return C.SP_PARITY_ODD
	case PARITY_EVEN:
		return C.SP_PARITY_EVEN
	case PARITY_MARK:
		return C.SP_PARITY_MARK
	case PARITY_SPACE:
		return C.SP_PARITY_SPACE
	default:
		return C.SP_PARITY_INVALID
	}
}

// Get the RTS pin behaviour from a port configuration. The port must
// be opened for this operation.
func (p *Port) RTS() (int, error) {
	rts := C.enum_sp_rts(C.SP_RTS_INVALID)
	if err := errmsg(C.sp_get_config_rts(p.c, &rts)); err != nil {
		return 0, err
	}
	return c2rts(rts), nil
}

// Set the RTS pin behaviour in a port configuration. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetRTS(rts int) error {
	crts := rts2c(rts)
	if err := errmsg(C.sp_set_config_rts(p.c, crts)); err != nil {
		return err
	}
	return p.getConf()
}

func c2rts(rts C.enum_sp_rts) int {
	switch rts {
	case C.SP_RTS_OFF:
		return RTS_OFF
	case C.SP_RTS_ON:
		return RTS_ON
	case C.SP_RTS_FLOW_CONTROL:
		return RTS_FLOW_CONTROL
	default:
		return RTS_INVALID
	}
}

func rts2c(rts int) C.enum_sp_rts {
	switch rts {
	case RTS_OFF:
		return C.SP_RTS_OFF
	case RTS_ON:
		return C.SP_RTS_ON
	case RTS_FLOW_CONTROL:
		return C.SP_RTS_FLOW_CONTROL
	default:
		return C.SP_RTS_INVALID
	}
}

// Get the CTS pin behaviour from a port configuration. The port must
// be opened for this operation.
func (p *Port) CTS() (int, error) {
	cts := C.enum_sp_cts(C.SP_CTS_INVALID)
	if err := errmsg(C.sp_get_config_cts(p.c, &cts)); err != nil {
		return 0, err
	}
	return c2cts(cts), nil
}

// Set the CTS pin behaviour in a port configuration. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetCTS(cts int) error {
	ccts := cts2c(cts)
	if err := errmsg(C.sp_set_config_cts(p.c, ccts)); err != nil {
		return err
	}
	return p.getConf()
}

func c2cts(cts C.enum_sp_cts) int {
	switch cts {
	case C.SP_CTS_IGNORE:
		return CTS_IGNORE
	case C.SP_CTS_FLOW_CONTROL:
		return CTS_FLOW_CONTROL
	default:
		return CTS_INVALID
	}
}

func cts2c(cts int) C.enum_sp_cts {
	switch cts {
	case CTS_IGNORE:
		return C.SP_CTS_IGNORE
	case CTS_FLOW_CONTROL:
		return C.SP_CTS_FLOW_CONTROL
	default:
		return C.SP_CTS_INVALID
	}
}

// Get the DTR pin behaviour from a port configuration. The port must
// be opened for this operation.
func (p *Port) DTR() (int, error) {
	dtr := C.enum_sp_dtr(C.SP_DTR_INVALID)
	if err := errmsg(C.sp_get_config_dtr(p.c, &dtr)); err != nil {
		return 0, err
	}
	return c2dtr(dtr), nil
}

// Set the DTR pin behaviour in a port configuration. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetDTR(dtr int) error {
	cdtr := dtr2c(dtr)
	if err := errmsg(C.sp_set_config_dtr(p.c, cdtr)); err != nil {
		return err
	}
	return p.getConf()
}

func c2dtr(dtr C.enum_sp_dtr) int {
	switch dtr {
	case C.SP_DTR_OFF:
		return DTR_OFF
	case C.SP_DTR_ON:
		return DTR_ON
	case C.SP_DTR_FLOW_CONTROL:
		return DTR_FLOW_CONTROL
	default:
		return DTR_INVALID
	}
}

func dtr2c(dtr int) C.enum_sp_dtr {
	switch dtr {
	case DTR_OFF:
		return C.SP_DTR_OFF
	case DTR_ON:
		return C.SP_DTR_ON
	case DTR_FLOW_CONTROL:
		return C.SP_DTR_FLOW_CONTROL
	default:
		return C.SP_DTR_INVALID
	}
}

// Get the DSR pin behaviour from a port configuration. The port must
// be opened for this operation.
func (p *Port) DSR() (int, error) {
	dsr := C.enum_sp_dsr(C.SP_DSR_INVALID)
	if err := errmsg(C.sp_get_config_dsr(p.c, &dsr)); err != nil {
		return 0, err
	}
	return c2dsr(dsr), nil
}

// Set the DSR pin behaviour in a port configuration. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetDSR(dsr int) error {
	cdsr := dsr2c(dsr)
	if err := errmsg(C.sp_set_config_dsr(p.c, cdsr)); err != nil {
		return err
	}
	return p.getConf()
}

func c2dsr(dsr C.enum_sp_dsr) int {
	switch dsr {
	case C.SP_DSR_IGNORE:
		return DSR_IGNORE
	case C.SP_DSR_FLOW_CONTROL:
		return DSR_FLOW_CONTROL
	default:
		return DSR_INVALID
	}
}

func dsr2c(dsr int) C.enum_sp_dsr {
	switch dsr {
	case DSR_IGNORE:
		return C.SP_DSR_IGNORE
	case DSR_FLOW_CONTROL:
		return C.SP_DSR_FLOW_CONTROL
	default:
		return C.SP_DSR_INVALID
	}
}

// Get the XON/XOFF configuration from a port configuration. The port
// must be opened for this operation.
func (p *Port) XonXoff() (int, error) {
	xon := C.enum_sp_xonxoff(C.SP_XONXOFF_INVALID)
	if err := errmsg(C.sp_get_config_xon_xoff(p.c, &xon)); err != nil {
		return 0, err
	}
	return c2xon(xon), nil
}

// Set the XON/XOFF configuration in a port configuration. The port
// must be opened for this operation. Call p.ApplyConfig() to apply
// the change.
func (p *Port) SetXonXoff(xon int) error {
	cxon := xon2c(xon)
	if err := errmsg(C.sp_set_config_xon_xoff(p.c, cxon)); err != nil {
		return err
	}
	return p.getConf()
}

func c2xon(xon C.enum_sp_xonxoff) int {
	switch xon {
	case C.SP_XONXOFF_DISABLED:
		return XONXOFF_DISABLED
	case C.SP_XONXOFF_IN:
		return XONXOFF_IN
	case C.SP_XONXOFF_OUT:
		return XONXOFF_OUT
	default:
		return XONXOFF_INVALID
	}
}

func xon2c(xon int) C.enum_sp_xonxoff {
	switch xon {
	case XONXOFF_DISABLED:
		return C.SP_XONXOFF_DISABLED
	case XONXOFF_IN:
		return C.SP_XONXOFF_IN
	case XONXOFF_OUT:
		return C.SP_XONXOFF_OUT
	default:
		return C.SP_XONXOFF_INVALID
	}
}

// Set the flow control type in a port configuration. The port must be
// opened for this operation. Call p.ApplyConfig() to apply the
// change.
func (p *Port) SetFlowControl(fc int) error {
	cfc, err := flow2c(fc)
	if err != nil {
		return err
	}

	if err := errmsg(C.sp_set_config_flowcontrol(p.c, cfc)); err != nil {
		return err
	}

	return p.getConf()
}

func flow2c(fc int) (cfc C.enum_sp_flowcontrol, err error) {
	switch fc {
	case FLOWCONTROL_NONE:
		cfc = C.SP_FLOWCONTROL_NONE
	case FLOWCONTROL_XONXOFF:
		cfc = C.SP_FLOWCONTROL_XONXOFF
	case FLOWCONTROL_RTSCTS:
		cfc = C.SP_FLOWCONTROL_RTSCTS
	case FLOWCONTROL_DTRDSR:
		cfc = C.SP_FLOWCONTROL_DTRDSR
	default:
		err = ErrInvalidArguments
	}
	return
}

// Implementation of io.Reader interface.
func (p *Port) Read(b []byte) (int, error) {
	var c int32
	var start time.Time

	if Debug {
		start = time.Now()
	}

	buf, size := unsafe.Pointer(&b[0]), C.size_t(len(b))

	if p.readDeadline.IsZero() {

		// no deadline
		c = C.sp_blocking_read(p.p, buf, size, 0)

	} else if millis := deadline2millis(p.readDeadline); millis <= 0 {

		// call nonblocking read
		c = C.sp_nonblocking_read(p.p, buf, size)

	} else {

		// call blocking read
		c = C.sp_blocking_read(p.p, buf, size, C.uint(millis))

	}

	if Debug {
		log.Printf("read time: %d ns", time.Since(start).Nanoseconds())
	}

	n := int(c)

	// check for error
	if n < 0 {
		return 0, errmsg(c)
	} else if n != len(b) {
		return n, ErrTimeout
	}

	// update slice length
	reflect.ValueOf(&b).Elem().SetLen(int(c))

	return n, nil
}

// Implementation of io.Writer interface.
func (p *Port) Write(b []byte) (int, error) {
	var c int32
	var start time.Time

	if Debug {
		start = time.Now()
	}

	buf, size := unsafe.Pointer(&b[0]), C.size_t(len(b))

	if p.writeDeadline.IsZero() {

		// no deadline
		c = C.sp_blocking_write(p.p, buf, size, 0)

	} else if millis := deadline2millis(p.writeDeadline); millis <= 0 {

		// call nonblocking write
		c = C.sp_nonblocking_write(p.p, buf, size)

	} else {

		// call blocking write
		c = C.sp_blocking_write(p.p, buf, size, C.uint(millis))

	}

	if Debug {
		log.Printf("write time: %d ns", time.Since(start).Nanoseconds())
	}

	n := int(c)

	// check for error
	if n < 0 {
		return 0, errmsg(c)
	} else if n != len(b) {
		return n, ErrTimeout
	}

	return n, nil
}

// WriteString is like Write, but writes the contents of string s
// rather than a slice of bytes.
func (p *Port) WriteString(s string) (int, error) {
	return p.Write(bytes.NewBufferString(s).Bytes())
}

// Implementation of net.Conn.LocalAddr
func (p *Port) LocalAddr() net.Addr {
	return &Addr{name: p.Name()}
}

// Implementation of net.Conn.RemoteAddr
func (p *Port) RemoteAddr() net.Addr {
	return &Addr{name: p.Name()}
}

// Implementation of net.Conn.SetDeadline
func (p *Port) SetDeadline(t time.Time) error {
	p.readDeadline = t
	p.writeDeadline = t
	return nil
}

// Implementation of net.Conn.SetReadDeadline
func (p *Port) SetReadDeadline(t time.Time) error {
	p.readDeadline = t
	return nil
}

// Implementation of net.Conn.SetWriteDeadline
func (p *Port) SetWriteDeadline(t time.Time) error {
	p.writeDeadline = t
	return nil
}

// Gets the number of bytes waiting in the input buffer.
func (p *Port) InputWaiting() (int, error) {
	c := C.sp_input_waiting(p.p)
	if c < 0 {
		return 0, errmsg(c)
	}
	return int(c), nil
}

// Gets the number of bytes waiting in the output buffer.
func (p *Port) OutputWaiting() (int, error) {
	c := C.sp_output_waiting(p.p)
	if c < 0 {
		return 0, errmsg(c)
	}
	return int(c), nil
}

// Wait for buffered data to be transmitted.
func (p *Port) Sync() error {
	return errmsg(C.sp_drain(p.p))
}

// Discard buffered data.
func (p *Port) Reset() error {
	return errmsg(C.sp_flush(p.p, C.SP_BUF_BOTH))
}

// Discard buffered input data.
func (p *Port) ResetInput() error {
	return errmsg(C.sp_flush(p.p, C.SP_BUF_INPUT))
}

// Discard buffered output data.
func (p *Port) ResetOutput() error {
	return errmsg(C.sp_flush(p.p, C.SP_BUF_OUTPUT))
}

// Implementation of net.Addr.Network()
func (a *Addr) Network() string {
	return "serial"
}

// Implementation of net.Addr.String()
func (a *Addr) String() string {
	return a.name
}

// Implementation of error.Error()
func (e *Error) Error() string {
	return e.msg
}

// Implementation of net.Error.Timeout()
func (e *Error) Timeout() bool {
	return e.timeout
}

// Implementation of net.Error.Temporary()
func (e *Error) Temporary() bool {
	return e.temporary
}
