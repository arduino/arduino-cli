// This file is part of arduino-cli.
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

package packagemanager

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestVidPidConvertionToPluggableDiscovery(t *testing.T) {
	m, err := properties.LoadFromBytes([]byte(`
arduino_zero_edbg.name=Arduino Zero (Programming Port)
arduino_zero_edbg.vid.0=0x03eb
arduino_zero_edbg.pid.0=0x2157
arduino_zero_edbg_2.name=Arduino Zero (Programming Port)
arduino_zero_edbg_2.vid=0x03eb
arduino_zero_edbg_2.pid=0x2157
arduino_zero_edbg_3.name=Arduino Zero (Programming Port)
arduino_zero_edbg_3.vid=0x03eb
arduino_zero_edbg_3.pid=0x2157
arduino_zero_edbg_3.vid.0=0x03ea
arduino_zero_edbg_3.pid.0=0x2157
arduino_zero_native.name=Arduino Zero (Native USB Port)
arduino_zero_native.vid.0=0x2341
arduino_zero_native.pid.0=0x804d
arduino_zero_native.vid.1=0x2341
arduino_zero_native.pid.1=0x004d
arduino_zero_native.vid.2=0x2341
arduino_zero_native.pid.2=0x824d
arduino_zero_native.vid.3=0x2341
arduino_zero_native.pid.3=0x024d
`))
	require.NoError(t, err)

	zero := m.SubTree("arduino_zero_edbg")
	convertVidPidIdentificationPropertiesToPluggableDiscovery(zero)
	require.Equal(t, `properties.Map{
  "name": "Arduino Zero (Programming Port)",
  "vid.0": "0x03eb",
  "pid.0": "0x2157",
  "upload_port.0.vid": "0x03eb",
  "upload_port.0.pid": "0x2157",
}`, zero.Dump())

	zero2 := m.SubTree("arduino_zero_edbg_2")
	convertVidPidIdentificationPropertiesToPluggableDiscovery(zero2)
	require.Equal(t, `properties.Map{
  "name": "Arduino Zero (Programming Port)",
  "vid": "0x03eb",
  "pid": "0x2157",
  "upload_port.0.vid": "0x03eb",
  "upload_port.0.pid": "0x2157",
}`, zero2.Dump())

	zero3 := m.SubTree("arduino_zero_edbg_3")
	convertVidPidIdentificationPropertiesToPluggableDiscovery(zero3)
	require.Equal(t, `properties.Map{
  "name": "Arduino Zero (Programming Port)",
  "vid": "0x03eb",
  "pid": "0x2157",
  "vid.0": "0x03ea",
  "pid.0": "0x2157",
  "upload_port.0.vid": "0x03eb",
  "upload_port.0.pid": "0x2157",
  "upload_port.1.vid": "0x03ea",
  "upload_port.1.pid": "0x2157",
}`, zero3.Dump())

	zero4 := m.SubTree("arduino_zero_native")
	convertVidPidIdentificationPropertiesToPluggableDiscovery(zero4)
	require.Equal(t, `properties.Map{
  "name": "Arduino Zero (Native USB Port)",
  "vid.0": "0x2341",
  "pid.0": "0x804d",
  "vid.1": "0x2341",
  "pid.1": "0x004d",
  "vid.2": "0x2341",
  "pid.2": "0x824d",
  "vid.3": "0x2341",
  "pid.3": "0x024d",
  "upload_port.0.vid": "0x2341",
  "upload_port.0.pid": "0x804d",
  "upload_port.1.vid": "0x2341",
  "upload_port.1.pid": "0x004d",
  "upload_port.2.vid": "0x2341",
  "upload_port.2.pid": "0x824d",
  "upload_port.3.vid": "0x2341",
  "upload_port.3.pid": "0x024d",
}`, zero4.Dump())
}

func TestLoadDiscoveries(t *testing.T) {
	// Create all the necessary data to load discoveries
	fakePath := paths.New("fake-path")

	createTestPackageManager := func() *PackageManager {
		packageManager := NewPackageManager(fakePath, fakePath, fakePath, fakePath)
		pack := packageManager.Packages.GetOrCreatePackage("arduino")
		// ble-discovery tool
		tool := pack.GetOrCreateTool("ble-discovery")
		toolRelease := tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath
		tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))

		// serial-discovery tool
		tool = pack.GetOrCreateTool("serial-discovery")
		tool.GetOrCreateRelease(semver.ParseRelaxed("1.0.0"))
		toolRelease = tool.GetOrCreateRelease(semver.ParseRelaxed("0.1.0"))
		// We set this to fake the tool is installed
		toolRelease.InstallDir = fakePath

		platform := pack.GetOrCreatePlatform("avr")
		release := platform.GetOrCreateRelease(semver.MustParse("1.0.0"))
		release.InstallDir = fakePath

		return packageManager
	}

	packageManager := createTestPackageManager()
	release := packageManager.Packages["arduino"].Platforms["avr"].Releases["1.0.0"]
	release.Properties = properties.NewFromHashmap(map[string]string{
		"pluggable_discovery.required": "arduino:ble-discovery",
	})

	errs := packageManager.LoadDiscoveries()
	require.Len(t, errs, 2)
	require.Equal(t, errs[0].Message(), "discovery not found: builtin:serial-discovery")
	require.Equal(t, errs[1].Message(), "discovery not found: builtin:mdns-discovery")
	discoveries := packageManager.DiscoveryManager().IDs()
	require.Len(t, discoveries, 1)
	require.Contains(t, discoveries, "arduino:ble-discovery")

	packageManager = createTestPackageManager()
	release = packageManager.Packages["arduino"].Platforms["avr"].Releases["1.0.0"]
	release.Properties = properties.NewFromHashmap(map[string]string{
		"pluggable_discovery.required.0": "arduino:ble-discovery",
		"pluggable_discovery.required.1": "arduino:serial-discovery",
	})

	errs = packageManager.LoadDiscoveries()
	require.Len(t, errs, 2)
	require.Equal(t, errs[0].Message(), "discovery not found: builtin:serial-discovery")
	require.Equal(t, errs[1].Message(), "discovery not found: builtin:mdns-discovery")
	discoveries = packageManager.DiscoveryManager().IDs()
	require.Len(t, discoveries, 2)
	require.Contains(t, discoveries, "arduino:ble-discovery")
	require.Contains(t, discoveries, "arduino:serial-discovery")

	packageManager = createTestPackageManager()
	release = packageManager.Packages["arduino"].Platforms["avr"].Releases["1.0.0"]
	release.Properties = properties.NewFromHashmap(map[string]string{
		"pluggable_discovery.required.0":     "arduino:ble-discovery",
		"pluggable_discovery.required.1":     "arduino:serial-discovery",
		"pluggable_discovery.teensy.pattern": "\"{runtime.tools.teensy_ports.path}/hardware/tools/teensy_ports\" -J2",
	})

	errs = packageManager.LoadDiscoveries()
	require.Len(t, errs, 2)
	require.Equal(t, errs[0].Message(), "discovery not found: builtin:serial-discovery")
	require.Equal(t, errs[1].Message(), "discovery not found: builtin:mdns-discovery")
	discoveries = packageManager.DiscoveryManager().IDs()
	require.Len(t, discoveries, 3)
	require.Contains(t, discoveries, "arduino:ble-discovery")
	require.Contains(t, discoveries, "arduino:serial-discovery")
	require.Contains(t, discoveries, "teensy")

	packageManager = createTestPackageManager()
	release = packageManager.Packages["arduino"].Platforms["avr"].Releases["1.0.0"]
	release.Properties = properties.NewFromHashmap(map[string]string{
		"pluggable_discovery.required":       "arduino:some-discovery",
		"pluggable_discovery.required.0":     "arduino:ble-discovery",
		"pluggable_discovery.required.1":     "arduino:serial-discovery",
		"pluggable_discovery.teensy.pattern": "\"{runtime.tools.teensy_ports.path}/hardware/tools/teensy_ports\" -J2",
	})

	errs = packageManager.LoadDiscoveries()
	require.Len(t, errs, 2)
	require.Equal(t, errs[0].Message(), "discovery not found: builtin:serial-discovery")
	require.Equal(t, errs[1].Message(), "discovery not found: builtin:mdns-discovery")
	discoveries = packageManager.DiscoveryManager().IDs()
	require.Len(t, discoveries, 3)
	require.Contains(t, discoveries, "arduino:ble-discovery")
	require.Contains(t, discoveries, "arduino:serial-discovery")
	require.Contains(t, discoveries, "teensy")
}

func TestConvertUploadToolsToPluggableDiscovery(t *testing.T) {
	props, err := properties.LoadFromBytes([]byte(`
upload.tool=avrdude
upload.protocol=arduino
upload.maximum_size=32256
upload.maximum_data_size=2048
upload.speed=115200
bootloader.tool=avrdude
bootloader.low_fuses=0xFF
bootloader.high_fuses=0xDE
bootloader.extended_fuses=0xFD
bootloader.unlock_bits=0x3F
bootloader.lock_bits=0x0F
bootloader.file=optiboot/optiboot_atmega328.hex
name=AVR ISP
communication=serial
protocol=stk500v1
program.protocol=stk500v1
program.tool=avrdude
program.extra_params=-P{serial.port}
`))
	require.NoError(t, err)

	convertUploadToolsToPluggableDiscovery(props)

	expectedProps, err := properties.LoadFromBytes([]byte(`
upload.tool=avrdude
upload.tool.default=avrdude
upload.protocol=arduino
upload.maximum_size=32256
upload.maximum_data_size=2048
upload.speed=115200
bootloader.tool=avrdude
bootloader.tool.default=avrdude
bootloader.low_fuses=0xFF
bootloader.high_fuses=0xDE
bootloader.extended_fuses=0xFD
bootloader.unlock_bits=0x3F
bootloader.lock_bits=0x0F
bootloader.file=optiboot/optiboot_atmega328.hex
name=AVR ISP
communication=serial
protocol=stk500v1
program.protocol=stk500v1
program.tool=avrdude
program.tool.default=avrdude
program.extra_params=-P{serial.port}
`))
	require.NoError(t, err)

	require.Equal(t, expectedProps.AsMap(), props.AsMap())
}
