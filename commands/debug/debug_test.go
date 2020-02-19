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

package debug

import (
	"fmt"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/common"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var customHardware = paths.New("testdata", "custom_hardware")
var dataDir = paths.New("testdata", "data_dir", "packages")
var sketch = "hello"
var sketchPath = paths.New("testdata", sketch).String()

func TestGetCommandLine(t *testing.T) {
	pm := packagemanager.NewPackageManager(nil, nil, nil, nil)
	pm.LoadHardwareFromDirectory(customHardware)
	pm.LoadHardwareFromDirectory(dataDir)

	req := &dbg.DebugConfigReq{
		Instance:   &rpc.Instance{Id: 1},
		Fqbn:       "arduino-test:samd:arduino_zero_edbg",
		SketchPath: sketchPath,
		Port:       "none",
	}
	packageName := strings.Split(req.Fqbn, ":")[0]
	processor := strings.Split(req.Fqbn, ":")[1]
	boardFamily := "arduino_zero"

	goldCommand := []string{
		fmt.Sprintf("%s/%s/tools/arm-none-eabi-gcc/7-2017q4/bin//arm-none-eabi-gdb", dataDir, packageName),
		fmt.Sprintf("-ex"),
		fmt.Sprintf("target extended-remote | %s/%s/tools/openocd/0.10.0-arduino7/bin/openocd", dataDir, packageName) + " " +
			fmt.Sprintf("-s \"%s/%s/tools/openocd/0.10.0-arduino7/share/openocd/scripts/\"", dataDir, packageName) + " " +
			fmt.Sprintf("--file \"%s/%s/%s/variants/%s/openocd_scripts/arduino_zero.cfg\" -c \"gdb_port pipe\" -c \"telnet_port 0\" -c init -c halt", customHardware, packageName, processor, boardFamily),
		fmt.Sprintf("%s/%s.%s.elf", sketchPath, sketch, strings.ReplaceAll(req.Fqbn, ":", ".")),
	}

	command, err := getCommandLine(req, pm)
	assert.Nil(t, err)
	assert.Equal(t, goldCommand, command)

	// for other samd boards
	req2 := &dbg.DebugConfigReq{
		Instance:   &rpc.Instance{Id: 1},
		Fqbn:       "arduino-test:samd:mkr1000",
		SketchPath: sketchPath,
		Port:       "none",
	}
	packageName2 := strings.Split(req2.Fqbn, ":")[0]
	processor2 := strings.Split(req2.Fqbn, ":")[1]
	name2 := strings.Split(req2.Fqbn, ":")[2]

	goldCommand2 := []string{
		fmt.Sprintf("%s/%s/tools/arm-none-eabi-gcc/7-2017q4/bin//arm-none-eabi-gdb", dataDir, packageName2),
		fmt.Sprintf("-ex"),
		fmt.Sprintf("target extended-remote | %s/%s/tools/openocd/0.10.0-arduino7/bin/openocd", dataDir, packageName2) + " " +
			fmt.Sprintf("-s \"%s/%s/tools/openocd/0.10.0-arduino7/share/openocd/scripts/\"", dataDir, packageName2) + " " +
			fmt.Sprintf("--file \"%s/%s/%s/variants/%s/openocd_scripts/arduino_zero.cfg\" -c \"gdb_port pipe\" -c \"telnet_port 0\" -c init -c halt", customHardware, packageName2, processor2, name2),
		fmt.Sprintf("%s/%s.%s.elf", sketchPath, sketch, strings.ReplaceAll(req2.Fqbn, ":", ".")),
	}

	command2, err := getCommandLine(req2, pm)
	assert.Nil(t, err)
	assert.Equal(t, goldCommand2, command2)

}
