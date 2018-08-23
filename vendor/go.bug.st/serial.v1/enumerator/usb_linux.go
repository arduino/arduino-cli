//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package enumerator // import "go.bug.st/serial.v1/enumerator"

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"go.bug.st/serial.v1"
)

func nativeGetDetailedPortsList() ([]*PortDetails, error) {
	// Retrieve the port list
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, &PortEnumerationError{causedBy: err}
	}

	var res []*PortDetails
	for _, port := range ports {
		details, err := nativeGetPortDetails(port)
		if err != nil {
			return nil, &PortEnumerationError{causedBy: err}
		}
		res = append(res, details)
	}
	return res, nil
}

func nativeGetPortDetails(portPath string) (*PortDetails, error) {
	portName := filepath.Base(portPath)
	devicePath := fmt.Sprintf("/sys/class/tty/%s/device", portName)
	if _, err := os.Stat(devicePath); err != nil {
		return &PortDetails{}, nil
	}
	realDevicePath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return nil, fmt.Errorf("Can't determine real path of %s: %s", devicePath, err.Error())
	}
	subSystemPath, err := filepath.EvalSymlinks(filepath.Join(realDevicePath, "subsystem"))
	if err != nil {
		return nil, fmt.Errorf("Can't determine real path of %s: %s", filepath.Join(realDevicePath, "subsystem"), err.Error())
	}
	subSystem := filepath.Base(subSystemPath)

	result := &PortDetails{Name: portPath}
	switch subSystem {
	case "usb-serial":
		err := parseUSBSysFS(filepath.Dir(filepath.Dir(realDevicePath)), result)
		return result, err
	case "usb":
		err := parseUSBSysFS(filepath.Dir(realDevicePath), result)
		return result, err
	// TODO: other cases?
	default:
		return result, nil
	}
}

func parseUSBSysFS(usbDevicePath string, details *PortDetails) error {
	vid, err := readLine(filepath.Join(usbDevicePath, "idVendor"))
	if err != nil {
		return err
	}
	pid, err := readLine(filepath.Join(usbDevicePath, "idProduct"))
	if err != nil {
		return err
	}
	serial, err := readLine(filepath.Join(usbDevicePath, "serial"))
	if err != nil {
		return err
	}
	//manufacturer, err := readLine(filepath.Join(usbDevicePath, "manufacturer"))
	//if err != nil {
	//	return err
	//}
	//product, err := readLine(filepath.Join(usbDevicePath, "product"))
	//if err != nil {
	//	return err
	//}

	details.IsUSB = true
	details.VID = vid
	details.PID = pid
	details.SerialNumber = serial
	//details.Manufacturer = manufacturer
	//details.Product = product
	return nil
}

func readLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line, _, err := reader.ReadLine()
	return string(line), err
}
