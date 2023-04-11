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

package cpp_test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/stretchr/testify/require"
)

func TestParseString(t *testing.T) {
	_, _, ok := cpp.ParseString(`foo`)
	require.Equal(t, false, ok)

	_, _, ok = cpp.ParseString(`"foo`)
	require.Equal(t, false, ok)

	str, rest, ok := cpp.ParseString(`"foo"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = cpp.ParseString(`"foo\\bar"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo\bar`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = cpp.ParseString(`"foo \"is\" quoted and \\\\bar\"\" escaped\\" and "then" some`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo "is" quoted and \\bar"" escaped\`, str)
	require.Equal(t, ` and "then" some`, rest)

	str, rest, ok = cpp.ParseString(`" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~"`)
	require.Equal(t, true, ok)
	require.Equal(t, ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = cpp.ParseString(`"/home/ççç/"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = cpp.ParseString(`"/home/ççç/ /$sdsdd\\"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/ /$sdsdd\`, str)
	require.Equal(t, ``, rest)
}

func TestQuoteString(t *testing.T) {
	cases := map[string]string{
		`foo`:                                  `"foo"`,
		`foo\bar`:                              `"foo\\bar"`,
		`foo "is" quoted and \\bar"" escaped\`: `"foo \"is\" quoted and \\\\bar\"\" escaped\\"`,
		// ASCII 0x20 - 0x7e, excluding `
		` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`: `" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~"`,
	}
	for input, expected := range cases {
		require.Equal(t, expected, cpp.QuoteString(input))
	}
}
