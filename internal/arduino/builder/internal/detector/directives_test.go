// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package detector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectiveLines(t *testing.T) {
	source := "#include <Arduino.h>\r\n" +
		"#line 1 \"/tmp/sketch/sketch.ino\"\n" +
		"  #include <Wire.h>\n" +
		"/* once */ #define LIMIT 10\n" +
		"#if defined(A) || \\\n" +
		"    defined(B)\n" +
		"void setup() {}\n" +
		"void loop() { Serial.println(\"#1\"); }\n" +
		"#endif"
	require.Equal(t, []string{
		"#include <Arduino.h>",
		"#line 1 \"/tmp/sketch/sketch.ino\"",
		"  #include <Wire.h>",
		"/* once */ #define LIMIT 10",
		"#if defined(A) || \\",
		"    defined(B)",
		"void loop() { Serial.println(\"#1\"); }",
		"#endif",
	}, directiveLines(source))
}

func TestDirectiveLinesOfPlainCode(t *testing.T) {
	require.Equal(t, []string{}, directiveLines("void setup() {}\nvoid loop() {}\n"))
}
