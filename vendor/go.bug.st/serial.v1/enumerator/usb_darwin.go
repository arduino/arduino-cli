//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package enumerator // import "go.bug.st/serial.v1/enumerator"

// #cgo LDFLAGS: -framework CoreFoundation -framework IOKit -fconstant-cfstrings
// #include <IOKit/IOKitLib.h>
// #include <CoreFoundation/CoreFoundation.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func nativeGetDetailedPortsList() ([]*PortDetails, error) {
	var ports []*PortDetails

	services, err := getAllServices("IOSerialBSDClient")
	if err != nil {
		return nil, &PortEnumerationError{causedBy: err}
	}
	for _, service := range services {
		defer service.Release()

		port, err := extractPortInfo(C.io_registry_entry_t(service))
		if err != nil {
			return nil, &PortEnumerationError{causedBy: err}
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func extractPortInfo(service C.io_registry_entry_t) (*PortDetails, error) {
	name, err := service.GetStringProperty("IOCalloutDevice")
	if err != nil {
		return nil, fmt.Errorf("Error extracting port info from device: %s", err.Error())
	}
	port := &PortDetails{}
	port.Name = name
	port.IsUSB = false

	usbDevice := service
	for usbDevice.GetClass() != "IOUSBDevice" {
		if usbDevice, err = usbDevice.GetParent("IOService"); err != nil {
			break
		}
	}
	if err == nil {
		// It's an IOUSBDevice
		vid, _ := usbDevice.GetIntProperty("idVendor", C.kCFNumberSInt16Type)
		pid, _ := usbDevice.GetIntProperty("idProduct", C.kCFNumberSInt16Type)
		serialNumber, _ := usbDevice.GetStringProperty("USB Serial Number")
		//product, _ := usbDevice.GetStringProperty("USB Product Name")
		//manufacturer, _ := usbDevice.GetStringProperty("USB Vendor Name")
		//fmt.Println(product + " - " + manufacturer)

		port.IsUSB = true
		port.VID = fmt.Sprintf("%04X", vid)
		port.PID = fmt.Sprintf("%04X", pid)
		port.SerialNumber = serialNumber
	}
	return port, nil
}

func getAllServices(serviceType string) ([]C.io_object_t, error) {
	i, err := getMatchingServices(serviceMatching(serviceType))
	if err != nil {
		return nil, err
	}
	defer i.Release()

	var services []C.io_object_t
	tries := 0
	for tries < 5 {
		// Extract all elements from iterator
		if service, ok := i.Next(); ok {
			services = append(services, service)
			continue
		}
		// If iterator is still valid return the result
		if i.IsValid() {
			return services, nil
		}
		// Otherwise empty the result and retry
		for _, s := range services {
			s.Release()
		}
		services = []C.io_object_t{}
		i.Reset()
		tries++
	}
	// Give up if the iteration continues to fail...
	return nil, fmt.Errorf("IOServiceGetMatchingServices failed, data changed while iterating")
}

// serviceMatching create a matching dictionary that specifies an IOService class match.
func serviceMatching(serviceType string) C.CFMutableDictionaryRef {
	t := C.CString(serviceType)
	defer C.free(unsafe.Pointer(t))
	return C.IOServiceMatching(t)
}

// getMatchingServices look up registered IOService objects that match a matching dictionary.
func getMatchingServices(matcher C.CFMutableDictionaryRef) (C.io_iterator_t, error) {
	var i C.io_iterator_t
	err := C.IOServiceGetMatchingServices(C.kIOMasterPortDefault, matcher, &i)
	if err != C.KERN_SUCCESS {
		return 0, fmt.Errorf("IOServiceGetMatchingServices failed (code %d)", err)
	}
	return i, nil
}

// CFStringRef

func cfStringCreateWithString(s string) C.CFStringRef {
	c := C.CString(s)
	defer C.free(unsafe.Pointer(c))
	return C.CFStringCreateWithCString(
		C.kCFAllocatorDefault, c, C.kCFStringEncodingMacRoman)
}

// io_registry_entry_t

func (me *C.io_registry_entry_t) GetParent(plane string) (C.io_registry_entry_t, error) {
	cPlane := C.CString(plane)
	defer C.free(unsafe.Pointer(cPlane))
	var parent C.io_registry_entry_t
	err := C.IORegistryEntryGetParentEntry(*me, cPlane, &parent)
	if err != 0 {
		return 0, errors.New("No parent device available")
	}
	return parent, nil
}

func (me *C.io_registry_entry_t) GetClass() string {
	obj := (*C.io_object_t)(me)
	return obj.GetClass()
}

func (me *C.io_registry_entry_t) GetStringProperty(key string) (string, error) {
	k := cfStringCreateWithString(key)
	defer C.CFRelease(C.CFTypeRef(k))
	property := C.IORegistryEntryCreateCFProperty(*me, k, C.kCFAllocatorDefault, 0)
	if property == nil {
		return "", errors.New("Property not found: " + key)
	}
	defer C.CFRelease(property)

	if ptr := C.CFStringGetCStringPtr((C.CFStringRef)(unsafe.Pointer(property)), 0); ptr != nil {
		return C.GoString(ptr), nil
	}
	// in certain circumstances CFStringGetCStringPtr may return NULL
	// and we must retrieve the string by copy
	buff := make([]C.char, 1024)
	if C.CFStringGetCString((C.CFStringRef)(property), &buff[0], 1024, 0) != C.true {
		return "", fmt.Errorf("Property '%s' can't be converted", key)
	}
	return C.GoString(&buff[0]), nil
}

func (me *C.io_registry_entry_t) GetIntProperty(key string, intType C.CFNumberType) (int, error) {
	k := cfStringCreateWithString(key)
	defer C.CFRelease(C.CFTypeRef(k))
	property := C.IORegistryEntryCreateCFProperty(*me, k, C.kCFAllocatorDefault, 0)
	if property == nil {
		return 0, errors.New("Property not found: " + key)
	}
	defer C.CFRelease(property)
	var res int
	if C.CFNumberGetValue((C.CFNumberRef)(property), intType, unsafe.Pointer(&res)) != C.true {
		return res, fmt.Errorf("Property '%s' can't be converted or has been truncated", key)
	}
	return res, nil
}

// io_iterator_t

// IsValid checks if an iterator is still valid.
// Some iterators will be made invalid if changes are made to the
// structure they are iterating over. This function checks the iterator
// is still valid and should be called when Next returns zero.
// An invalid iterator can be Reset and the iteration restarted.
func (me *C.io_iterator_t) IsValid() bool {
	return C.IOIteratorIsValid(*me) == C.true
}

func (me *C.io_iterator_t) Reset() {
	C.IOIteratorReset(*me)
}

func (me *C.io_iterator_t) Next() (C.io_object_t, bool) {
	res := C.IOIteratorNext(*me)
	return res, res != 0
}

func (me *C.io_iterator_t) Release() {
	C.IOObjectRelease(C.io_object_t(*me))
}

// io_object_t

func (me *C.io_object_t) Release() {
	C.IOObjectRelease(*me)
}

func (me *C.io_object_t) GetClass() string {
	class := make([]C.char, 1024)
	C.IOObjectGetClass(*me, &class[0])
	return C.GoString(&class[0])
}
