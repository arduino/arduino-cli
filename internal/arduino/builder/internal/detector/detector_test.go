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

package detector_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/detector"
	"github.com/stretchr/testify/require"
)

func TestIncludesFinderWithRegExp(t *testing.T) {
	output := "/some/path/sketch.ino:1:17: fatal error: SPI.h: No such file or directory\n" +
		"#include <SPI.h>\n" +
		"^\n" +
		"compilation terminated."
	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "SPI.h", include)
}

func TestIncludesFinderWithRegExpEmptyOutput(t *testing.T) {
	include := detector.IncludesFinderWithRegExp("")

	require.Equal(t, "", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes(t *testing.T) {
	output := "/some/path/sketch.ino:1:33: fatal error: Wire.h: No such file or directory\n" +
		" #               include <Wire.h>\n" +
		"                                 ^\n" +
		"compilation terminated.\n"
	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "Wire.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes2(t *testing.T) {
	output := "/some/path/sketch.ino:1:33: fatal error: Wire.h: No such file or directory\n" +
		" #\t\t\tinclude <Wire.h>\n" +
		"                                 ^\n" +
		"compilation terminated.\n"
	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "Wire.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes3(t *testing.T) {
	output := "/some/path/sketch.ino:1:33: fatal error: SPI.h: No such file or directory\n" +
		"compilation terminated.\n"

	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "SPI.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes4(t *testing.T) {
	output := "In file included from /tmp/arduino_modified_sketch_815412/binouts.ino:52:0:\n" +
		"/tmp/arduino_build_static/sketch/regtable.h:31:22: fatal error: register.h: No such file or directory\n"

	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "register.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes5(t *testing.T) {
	output := "/some/path/sketch.ino:23:42: fatal error: 'Foobar.h' file not found" +
              "   23 | #include <Foobar.h>" +
              "      |          ^~~~~~~~~~"

	include := detector.IncludesFinderWithRegExp(output)

	require.Equal(t, "Foobar.h", include)
}
