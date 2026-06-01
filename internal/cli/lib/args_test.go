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

package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goodLibs = []struct {
	in       string
	expected *LibraryReferenceArg
}{
	{"mylib", &LibraryReferenceArg{"mylib", ""}},
	{"mylib@1.0", &LibraryReferenceArg{"mylib", "1.0"}},
}

var badLibs = []struct {
	in       string
	expected *LibraryReferenceArg
}{
	{"", nil},
	{"mylib@", nil},
}

func TestArgsStringify(t *testing.T) {
	for _, lib := range goodLibs {
		require.Equal(t, lib.in, lib.expected.String())
	}
}

func TestParseReferenceArgLibs(t *testing.T) {
	for _, tt := range goodLibs {
		actual, err := ParseLibraryReferenceArg(tt.in)
		assert.Nil(t, err, "Testing good arg '%s'", tt.in)
		assert.Equal(t, tt.expected, actual, "Testing good arg '%s'", tt.in)
	}
	for _, tt := range badLibs {
		res, err := ParseLibraryReferenceArg(tt.in)
		require.Nil(t, res, "Testing bad arg '%s'", tt.in)
		require.NotNil(t, err, "Testing bad arg '%s'", tt.in)
	}
}

func TestParseLibraryReferenceArgs(t *testing.T) {
	args := []string{}
	for _, tt := range goodLibs {
		args = append(args, tt.in)
	}
	refs, err := ParseLibraryReferenceArgs(args)
	require.Nil(t, err)
	require.Len(t, refs, len(goodLibs))
	for i, tt := range goodLibs {
		assert.Equal(t, tt.expected, refs[i])
	}
}
