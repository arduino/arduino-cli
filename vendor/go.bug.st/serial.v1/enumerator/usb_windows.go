//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package enumerator // import "go.bug.st/serial.v1/enumerator"

import (
	"fmt"
	"regexp"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func parseDeviceID(deviceID string, details *PortDetails) {
	// Windows stock USB-CDC driver
	if len(deviceID) >= 3 && deviceID[:3] == "USB" {
		re := regexp.MustCompile("VID_(....)&PID_(....)(\\\\(\\w+)$)?").FindAllStringSubmatch(deviceID, -1)
		if re == nil || len(re[0]) < 2 {
			// Silently ignore unparsable strings
			return
		}
		details.IsUSB = true
		details.VID = re[0][1]
		details.PID = re[0][2]
		if len(re[0]) >= 4 {
			details.SerialNumber = re[0][4]
		}
		return
	}

	// FTDI driver
	if len(deviceID) >= 7 && deviceID[:7] == "FTDIBUS" {
		re := regexp.MustCompile("VID_(....)\\+PID_(....)(\\+(\\w+))?").FindAllStringSubmatch(deviceID, -1)
		if re == nil || len(re[0]) < 2 {
			// Silently ignore unparsable strings
			return
		}
		details.IsUSB = true
		details.VID = re[0][1]
		details.PID = re[0][2]
		if len(re[0]) >= 4 {
			details.SerialNumber = re[0][4]
		}
		return
	}

	// Other unidentified device type
}

// setupapi based
// --------------

//sys setupDiClassGuidsFromNameInternal(class string, guid *guid, guidSize uint32, requiredSize *uint32) (err error) = setupapi.SetupDiClassGuidsFromNameW
//sys setupDiGetClassDevs(guid *guid, enumerator *string, hwndParent uintptr, flags uint32) (set devicesSet, err error) = setupapi.SetupDiGetClassDevsW
//sys setupDiDestroyDeviceInfoList(set devicesSet) (err error) = setupapi.SetupDiDestroyDeviceInfoList
//sys setupDiEnumDeviceInfo(set devicesSet, index uint32, info *devInfoData) (err error) = setupapi.SetupDiEnumDeviceInfo
//sys setupDiGetDeviceInstanceId(set devicesSet, devInfo *devInfoData, devInstanceId unsafe.Pointer, devInstanceIdSize uint32, requiredSize *uint32) (err error) = setupapi.SetupDiGetDeviceInstanceIdW
//sys setupDiOpenDevRegKey(set devicesSet, devInfo *devInfoData, scope dicsScope, hwProfile uint32, keyType uint32, samDesired regsam) (hkey syscall.Handle, err error) = setupapi.SetupDiOpenDevRegKey
//sys setupDiGetDeviceRegistryProperty(set devicesSet, devInfo *devInfoData, property deviceProperty, propertyType *uint32, outValue *byte, outSize *uint32, reqSize *uint32) (res bool) = setupapi.SetupDiGetDeviceRegistryPropertyW

// Device registry property codes
// (Codes marked as read-only (R) may only be used for
// SetupDiGetDeviceRegistryProperty)
//
// These values should cover the same set of registry properties
// as defined by the CM_DRP codes in cfgmgr32.h.
//
// Note that SPDRP codes are zero based while CM_DRP codes are one based!
type deviceProperty uint32

const (
	spdrpDeviceDesc               deviceProperty = 0x00000000 // DeviceDesc = R/W
	spdrpHardwareID                              = 0x00000001 // HardwareID = R/W
	spdrpCompatibleIDS                           = 0x00000002 // CompatibleIDs = R/W
	spdrpUnused0                                 = 0x00000003 // Unused
	spdrpService                                 = 0x00000004 // Service = R/W
	spdrpUnused1                                 = 0x00000005 // Unused
	spdrpUnused2                                 = 0x00000006 // Unused
	spdrpClass                                   = 0x00000007 // Class = R--tied to ClassGUID
	spdrpClassGUID                               = 0x00000008 // ClassGUID = R/W
	spdrpDriver                                  = 0x00000009 // Driver = R/W
	spdrpConfigFlags                             = 0x0000000A // ConfigFlags = R/W
	spdrpMFG                                     = 0x0000000B // Mfg = R/W
	spdrpFriendlyName                            = 0x0000000C // FriendlyName = R/W
	spdrpLocationIinformation                    = 0x0000000D // LocationInformation = R/W
	spdrpPhysicalDeviceObjectName                = 0x0000000E // PhysicalDeviceObjectName = R
	spdrpCapabilities                            = 0x0000000F // Capabilities = R
	spdrpUINumber                                = 0x00000010 // UiNumber = R
	spdrpUpperFilters                            = 0x00000011 // UpperFilters = R/W
	spdrpLowerFilters                            = 0x00000012 // LowerFilters = R/W
	spdrpBusTypeGUID                             = 0x00000013 // BusTypeGUID = R
	spdrpLegactBusType                           = 0x00000014 // LegacyBusType = R
	spdrpBusNumber                               = 0x00000015 // BusNumber = R
	spdrpEnumeratorName                          = 0x00000016 // Enumerator Name = R
	spdrpSecurity                                = 0x00000017 // Security = R/W, binary form
	spdrpSecuritySDS                             = 0x00000018 // Security = W, SDS form
	spdrpDevType                                 = 0x00000019 // Device Type = R/W
	spdrpExclusive                               = 0x0000001A // Device is exclusive-access = R/W
	spdrpCharacteristics                         = 0x0000001B // Device Characteristics = R/W
	spdrpAddress                                 = 0x0000001C // Device Address = R
	spdrpUINumberDescFormat                      = 0X0000001D // UiNumberDescFormat = R/W
	spdrpDevicePowerData                         = 0x0000001E // Device Power Data = R
	spdrpRemovalPolicy                           = 0x0000001F // Removal Policy = R
	spdrpRemovalPolicyHWDefault                  = 0x00000020 // Hardware Removal Policy = R
	spdrpRemovalPolicyOverride                   = 0x00000021 // Removal Policy Override = RW
	spdrpInstallState                            = 0x00000022 // Device Install State = R
	spdrpLocationPaths                           = 0x00000023 // Device Location Paths = R
	spdrpBaseContainerID                         = 0x00000024 // Base ContainerID = R

	spdrpMaximumProperty = 0x00000025 // Upper bound on ordinals
)

// Values specifying the scope of a device property change
type dicsScope uint32

const (
	dicsFlagGlobal          dicsScope = 0x00000001 // make change in all hardware profiles
	dicsFlagConfigSspecific           = 0x00000002 // make change in specified profile only
	dicsFlagConfigGeneral             = 0x00000004 // 1 or more hardware profile-specific
)

// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724878(v=vs.85).aspx
type regsam uint32

const (
	keyAllAccess        regsam = 0xF003F
	keyCreateLink              = 0x00020
	keyCreateSubKey            = 0x00004
	keyEnumerateSubKeys        = 0x00008
	keyExecute                 = 0x20019
	keyNotify                  = 0x00010
	keyQueryValue              = 0x00001
	keyRead                    = 0x20019
	keySetValue                = 0x00002
	keyWOW64_32key             = 0x00200
	keyWOW64_64key             = 0x00100
	keyWrite                   = 0x20006
)

// KeyType values for SetupDiCreateDevRegKey, SetupDiOpenDevRegKey, and
// SetupDiDeleteDevRegKey.
const (
	diregDev  = 0x00000001 // Open/Create/Delete device key
	diregDrv  = 0x00000002 // Open/Create/Delete driver key
	diregBoth = 0x00000004 // Delete both driver and Device key
)

// https://msdn.microsoft.com/it-it/library/windows/desktop/aa373931(v=vs.85).aspx
type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

func (g guid) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		g.data1, g.data2, g.data3,
		g.data4[0], g.data4[1], g.data4[2], g.data4[3],
		g.data4[4], g.data4[5], g.data4[6], g.data4[7])
}

func classGuidsFromName(className string) ([]guid, error) {
	// Determine the number of GUIDs for className
	n := uint32(0)
	if err := setupDiClassGuidsFromNameInternal(className, nil, 0, &n); err != nil {
		// ignore error: UIDs array size too small
	}

	res := make([]guid, n)
	err := setupDiClassGuidsFromNameInternal(className, &res[0], n, &n)
	return res, err
}

const (
	digcfDefault         = 0x00000001 // only valid with digcfDeviceInterface
	digcfPresent         = 0x00000002
	digcfAllClasses      = 0x00000004
	digcfProfile         = 0x00000008
	digcfDeviceInterface = 0x00000010
)

type devicesSet syscall.Handle

func (g *guid) getDevicesSet() (devicesSet, error) {
	return setupDiGetClassDevs(g, nil, 0, digcfPresent)
}

func (set devicesSet) destroy() {
	setupDiDestroyDeviceInfoList(set)
}

// https://msdn.microsoft.com/en-us/library/windows/hardware/ff552344(v=vs.85).aspx
type devInfoData struct {
	size     uint32
	guid     guid
	devInst  uint32
	reserved uintptr
}

type deviceInfo struct {
	set  devicesSet
	data devInfoData
}

func (set devicesSet) getDeviceInfo(index int) (*deviceInfo, error) {
	result := &deviceInfo{set: set}

	result.data.size = uint32(unsafe.Sizeof(result.data))
	err := setupDiEnumDeviceInfo(set, uint32(index), &result.data)
	return result, err
}

func (dev *deviceInfo) getInstanceID() (string, error) {
	n := uint32(0)
	setupDiGetDeviceInstanceId(dev.set, &dev.data, nil, 0, &n)
	buff := make([]uint16, n)
	if err := setupDiGetDeviceInstanceId(dev.set, &dev.data, unsafe.Pointer(&buff[0]), uint32(len(buff)), &n); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buff[:]), nil
}

func (dev *deviceInfo) openDevRegKey(scope dicsScope, hwProfile uint32, keyType uint32, samDesired regsam) (syscall.Handle, error) {
	return setupDiOpenDevRegKey(dev.set, &dev.data, scope, hwProfile, keyType, samDesired)
}

func nativeGetDetailedPortsList() ([]*PortDetails, error) {
	guids, err := classGuidsFromName("Ports")
	if err != nil {
		return nil, &PortEnumerationError{causedBy: err}
	}

	var res []*PortDetails
	for _, g := range guids {
		devsSet, err := g.getDevicesSet()
		if err != nil {
			return nil, &PortEnumerationError{causedBy: err}
		}
		defer devsSet.destroy()

		for i := 0; ; i++ {
			device, err := devsSet.getDeviceInfo(i)
			if err != nil {
				break
			}
			details := &PortDetails{}
			portName, err := retrievePortNameFromDevInfo(device)
			if err != nil {
				continue
			}
			if len(portName) < 3 || portName[0:3] != "COM" {
				// Accept only COM ports
				continue
			}
			details.Name = portName

			if err := retrievePortDetailsFromDevInfo(device, details); err != nil {
				return nil, &PortEnumerationError{causedBy: err}
			}
			res = append(res, details)
		}
	}
	return res, nil
}

func retrievePortNameFromDevInfo(device *deviceInfo) (string, error) {
	h, err := device.openDevRegKey(dicsFlagGlobal, 0, diregDev, keyRead)
	if err != nil {
		return "", err
	}
	defer syscall.RegCloseKey(h)

	var name [1024]uint16
	nameP := (*byte)(unsafe.Pointer(&name[0]))
	nameSize := uint32(len(name) * 2)
	if err := syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr("PortName"), nil, nil, nameP, &nameSize); err != nil {
		return "", err
	}
	return syscall.UTF16ToString(name[:]), nil
}

func retrievePortDetailsFromDevInfo(device *deviceInfo, details *PortDetails) error {
	deviceID, err := device.getInstanceID()
	if err != nil {
		return err
	}
	parseDeviceID(deviceID, details)

	var friendlyName [1024]uint16
	friendlyNameP := (*byte)(unsafe.Pointer(&friendlyName[0]))
	friendlyNameSize := uint32(len(friendlyName) * 2)
	if setupDiGetDeviceRegistryProperty(device.set, &device.data, spdrpDeviceDesc /* spdrpFriendlyName */, nil, friendlyNameP, &friendlyNameSize, nil) {
		//details.Product = syscall.UTF16ToString(friendlyName[:])
	}

	return nil
}
