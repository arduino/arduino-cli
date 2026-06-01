// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package diagnostics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	runProcess = mockedRunProcessToGetCompilerVersion
}

func mockedRunProcessToGetCompilerVersion(args ...string) []string {
	if strings.HasSuffix(args[0], "7.3.0-atmel3.6.1-arduino7/bin/avr-g++") && args[1] == "--version" {
		return []string{
			"avr-g++ (GCC) 7.3.0",
			"Copyright (C) 2017 Free Software Foundation, Inc.",
			"This is free software; see the source for copying conditions.  There is NO",
			"warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.",
			"",
		}
	}
	if strings.HasSuffix(args[0], "7.3.0-atmel3.6.1-arduino7/bin/avr-gcc") && args[1] == "--version" {
		return []string{
			"avr-gcc (GCC) 7.3.0",
			"Copyright (C) 2017 Free Software Foundation, Inc.",
			"This is free software; see the source for copying conditions.  There is NO",
			"warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.",
			"",
		}
	}
	if strings.HasSuffix(args[0], "xtensa-esp32-elf-gcc/gcc8_4_0-esp-2021r2-patch3/bin/xtensa-esp32-elf-g++") && args[1] == "--version" {
		return []string{
			"xtensa-esp32-elf-g++ (crosstool-NG esp-2021r2-patch3) 8.4.0",
			"Copyright (C) 2018 Free Software Foundation, Inc.",
			"This is free software; see the source for copying conditions.  There is NO",
			"warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.",
			"",
		}
	}

	panic("missing mock for command line: " + strings.Join(args, " "))
}

func TestCompilerDetection(t *testing.T) {
	comp := DetectCompilerFromCommandLine([]string{"~/.arduino15/packages/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7/bin/avr-g++"}, true)
	require.NotNil(t, comp)
	require.Equal(t, "gcc", comp.Family)
	require.Equal(t, "avr-g++", comp.Name)
	require.Equal(t, "7.3.0", comp.Version.String())

	comp = DetectCompilerFromCommandLine([]string{"~/.arduino15/packages/arduino/tools/avr-gcc/7.3.0-atmel3.6.1-arduino7/bin/avr-gcc"}, true)
	require.NotNil(t, comp)
	require.Equal(t, "gcc", comp.Family)
	require.Equal(t, "avr-gcc", comp.Name)
	require.Equal(t, "7.3.0", comp.Version.String())

	comp = DetectCompilerFromCommandLine([]string{"/home/megabug/.arduino15/packages/esp32/tools/xtensa-esp32-elf-gcc/gcc8_4_0-esp-2021r2-patch3/bin/xtensa-esp32-elf-g++"}, true)
	require.NotNil(t, comp)
	require.Equal(t, "gcc", comp.Family)
	require.Equal(t, "xtensa-esp32-elf-g++", comp.Name)
	require.Equal(t, "8.4.0", comp.Version.String())
}
