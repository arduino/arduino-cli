//
// Copyright 2014-2017 Lars Knudsen, Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build ignore

package enumerator

import (
	"log"
	"regexp"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func init() {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		log.Fatal("Init error: ", err)
	}
}

func nativeGetDetailedPortsList() ([]*PortDetails, error) {
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
	}
	defer wmi.Release()

	serviceRaw, err := wmi.CallMethod("ConnectServer")
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := "SELECT * FROM Win32_PnPEntity WHERE ConfigManagerErrorCode = 0 and Name like '%(COM%'"
	queryResult, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
	}
	result := queryResult.ToIDispatch()
	defer result.Release()

	countVar, err := result.GetProperty("Count")
	if err != nil {
		return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
	}
	count := int(countVar.Val)

	res := []*PortDetails{}

	// Retrieve all items
	for i := 0; i < count; i++ {
		itemRaw, err := result.CallMethod("ItemIndex", i)
		if err != nil {
			return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
		}
		item := itemRaw.ToIDispatch()
		defer item.Release()

		detail := &PortDetails{}
		if err := getPortDetails(item, detail); err != nil {
			return nil, &PortError{code: ErrorEnumeratingPorts, causedBy: err}
		}
		//  SerialPort{Path: path, VendorId: VID, ProductId: PID, DisplayName: displayName.ToString()}
		res = append(res, detail)
	}

	return res, nil
}

func getPortDetails(item *ole.IDispatch, res *PortDetails) error {
	// Find port name
	itemName, err := item.GetProperty("Name")
	if err != nil {
		return err
	}
	re := regexp.MustCompile("\\((COM[0-9]+)\\)").FindAllStringSubmatch(itemName.ToString(), 1)
	if re == nil || len(re[0]) < 2 {
		// Discard items that are not serial ports
		return nil
	}
	res.Name = re[0][1]

	//itemPnPDeviceID, err := item.GetProperty("PnPDeviceID")
	//if err != nil {
	//	return err
	//}
	//PnPDeviceID := itemPnPDeviceID.ToString()

	itemDeviceID, err := item.GetProperty("DeviceID")
	if err != nil {
		return err
	}
	parseDeviceID(itemDeviceID.ToString(), res)

	itemManufacturer, err := item.GetProperty("Product")
	if err != nil {
		return err
	}
	res.Manufacturer = itemManufacturer.ToString()
	return nil
}
