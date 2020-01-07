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

package globals_test

import (
	"testing"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goodCores = []struct {
	in       string
	expected *globals.ReferenceArg
}{
	{"arduino:avr", &globals.ReferenceArg{"arduino", "avr", ""}},
	{"arduino:avr@1.6.20", &globals.ReferenceArg{"arduino", "avr", "1.6.20"}},
}

var goodLibs = []struct {
	in       string
	expected *globals.LibraryReferenceArg
}{
	{"mylib", &globals.LibraryReferenceArg{"mylib", ""}},
	{"mylib@1.0", &globals.LibraryReferenceArg{"mylib", "1.0"}},
}

var badCores = []struct {
	in       string
	expected *globals.ReferenceArg
}{
	{"arduino:avr:avr", nil},
	{"arduino@1.6.20:avr", nil},
	{"arduino:avr:avr@1.6.20", nil},
	{"arduino:@1.6.20", nil},
	{":avr@1.5.0", nil},
	{"@1.5.0", nil},
	{"arduino:avr@", nil},
	{"", nil},
}

var badLibs = []struct {
	in       string
	expected *globals.LibraryReferenceArg
}{
	{"", nil},
	{"mylib@", nil},
}

func TestArgsStringify(t *testing.T) {
	for _, lib := range goodLibs {
		require.Equal(t, lib.in, lib.expected.String())
	}
	for _, core := range goodCores {
		require.Equal(t, core.in, core.expected.String())
	}
}

func TestParseReferenceArgCores(t *testing.T) {
	for _, tt := range goodCores {
		actual, err := globals.ParseReferenceArg(tt.in, true)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual)
	}

	for _, tt := range badCores {
		actual, err := globals.ParseReferenceArg(tt.in, true)
		require.NotNil(t, err, "Testing bad core '%s'", tt.in)
		require.Equal(t, tt.expected, actual, "Testing bad core '%s'", tt.in)
	}
}

func TestParseReferenceArgLibs(t *testing.T) {
	for _, tt := range goodLibs {
		actual, err := globals.ParseLibraryReferenceArg(tt.in)
		assert.Nil(t, err, "Testing good arg '%s'", tt.in)
		assert.Equal(t, tt.expected, actual, "Testing good arg '%s'", tt.in)
	}
	for _, tt := range badLibs {
		res, err := globals.ParseLibraryReferenceArg(tt.in)
		require.Nil(t, res, "Testing bad arg '%s'", tt.in)
		require.NotNil(t, err, "Testing bad arg '%s'", tt.in)
	}
}

func TestParseLibraryReferenceArgs(t *testing.T) {
	args := []string{}
	for _, tt := range goodLibs {
		args = append(args, tt.in)
	}
	refs, err := globals.ParseLibraryReferenceArgs(args)
	require.Nil(t, err)
	require.Len(t, refs, len(goodLibs))
	for i, tt := range goodLibs {
		assert.Equal(t, tt.expected, refs[i])
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
