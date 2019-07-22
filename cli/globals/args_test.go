/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package globals_test

import (
	"testing"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/stretchr/testify/assert"
)

var goodCores = []struct {
	in       string
	expected *globals.ReferenceArg
}{
	{"arduino:avr", &globals.ReferenceArg{"arduino", "avr", ""}},
	{"arduino:avr@1.6.20", &globals.ReferenceArg{"arduino", "avr", "1.6.20"}},
	{"arduino:avr@", &globals.ReferenceArg{"arduino", "avr", ""}},
}

var goodLibs = []struct {
	in       string
	expected *globals.ReferenceArg
}{
	{"mylib", &globals.ReferenceArg{"mylib", "", ""}},
	{"mylib@1.0", &globals.ReferenceArg{"mylib", "", "1.0"}},
	{"mylib@", &globals.ReferenceArg{"mylib", "", ""}},
}

var badCores = []struct {
	in       string
	expected *globals.ReferenceArg
}{
	{"arduino:avr:avr", nil},
	{"arduino@1.6.20:avr", nil},
	{"arduino:avr:avr@1.6.20", nil},
}

func TestParseReferenceArgCores(t *testing.T) {
	for _, tt := range goodCores {
		actual, err := globals.ParseReferenceArg(tt.in, true)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual)
	}

	for _, tt := range badCores {
		actual, err := globals.ParseReferenceArg(tt.in, true)
		assert.NotNil(t, err)
		assert.Equal(t, tt.expected, actual)
	}

	// library refs are not good as core's
	for _, tt := range goodLibs {
		_, err := globals.ParseReferenceArg(tt.in, true)
		assert.NotNil(t, err)
	}
}

func TestParseReferenceArgLibs(t *testing.T) {
	for _, tt := range goodLibs {
		actual, err := globals.ParseReferenceArg(tt.in, false)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual)
	}

	// good libs are bad when requiring Arch to be present
	for _, tt := range goodLibs {
		_, err := globals.ParseReferenceArg(tt.in, true)
		assert.NotNil(t, err)
	}
}

func TestParseArgs(t *testing.T) {
	input := []string{}
	for _, tt := range goodCores {
		input = append(input, tt.in)
	}

	refs, err := globals.ParseReferenceArgs(input, true)
	assert.Nil(t, err)
	assert.Equal(t, len(goodCores), len(refs))

	for i, tt := range goodCores {
		assert.Equal(t, tt.expected, refs[i])
	}
}
