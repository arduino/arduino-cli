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

package arguments_test

import (
	"testing"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goodCores = []struct {
	in       string
	expected *arguments.Reference
}{
	{"arduino:avr", &arguments.Reference{"arduino", "avr", ""}},
	{"arduino:avr@1.6.20", &arguments.Reference{"arduino", "avr", "1.6.20"}},
}

var badCores = []struct {
	in       string
	expected *arguments.Reference
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

func TestArgsStringify(t *testing.T) {
	for _, core := range goodCores {
		require.Equal(t, core.in, core.expected.String())
	}
}

func TestParseReferenceCores(t *testing.T) {
	for _, tt := range goodCores {
		actual, err := arguments.ParseReference(tt.in, true)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual)
	}

	for _, tt := range badCores {
		actual, err := arguments.ParseReference(tt.in, true)
		require.NotNil(t, err, "Testing bad core '%s'", tt.in)
		require.Equal(t, tt.expected, actual, "Testing bad core '%s'", tt.in)
	}
}

func TestParseArgs(t *testing.T) {
	input := []string{}
	for _, tt := range goodCores {
		input = append(input, tt.in)
	}

	refs, err := arguments.ParseReferences(input, true)
	assert.Nil(t, err)
	assert.Equal(t, len(goodCores), len(refs))

	for i, tt := range goodCores {
		assert.Equal(t, tt.expected, refs[i])
	}
}
