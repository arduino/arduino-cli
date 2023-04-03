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

package test

import (
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/stretchr/testify/require"
)

func TestPrintableCommand(t *testing.T) {
	parts := []string{
		"/path/to/dir with spaces/cmd",
		"arg1",
		"arg-\"with\"-quotes",
		"specialchar-`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?-argument",
		"arg   with spaces",
		"arg\twith\t\ttabs",
		"lastarg",
	}
	correct := "\"/path/to/dir with spaces/cmd\"" +
		" arg1 \"arg-\\\"with\\\"-quotes\"" +
		" \"specialchar-`~!@#$%^&*()-_=+[{]}\\\\|;:'\\\",<.>/?-argument\"" +
		" \"arg   with spaces\" \"arg\twith\t\ttabs\"" +
		" lastarg"
	result := utils.PrintableCommand(parts)
	require.Equal(t, correct, result)
}

func TestMapTrimSpace(t *testing.T) {
	value := "hello, world , how are,you? "
	parts := utils.Map(strings.Split(value, ","), utils.TrimSpace)

	require.Equal(t, 4, len(parts))
	require.Equal(t, "hello", parts[0])
	require.Equal(t, "world", parts[1])
	require.Equal(t, "how are", parts[2])
	require.Equal(t, "you?", parts[3])
}

func TestQuoteCppString(t *testing.T) {
	cases := map[string]string{
		`foo`:                                  `"foo"`,
		`foo\bar`:                              `"foo\\bar"`,
		`foo "is" quoted and \\bar"" escaped\`: `"foo \"is\" quoted and \\\\bar\"\" escaped\\"`,
		// ASCII 0x20 - 0x7e, excluding `
		` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`: `" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~"`,
	}
	for input, expected := range cases {
		require.Equal(t, expected, utils.QuoteCppString(input))
	}
}
