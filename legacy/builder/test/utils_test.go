/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package test

import (
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestCommandLineParser(t *testing.T) {
	command := "\"/home/federico/materiale/works_Arduino/Arduino/build/hardware/tools/coan\" source -m -E -P -kb -c -g -Os -w -ffunction-sections -fdata-sections -MMD -mmcu=atmega32u4 -DF_CPU=16000000L -DARDUINO=010600 -DARDUINO_AVR_LEONARDO -DARDUINO_ARCH_AVR  -DUSB_VID=0x2341 -DUSB_PID=0x8036 '-DUSB_MANUFACTURER=' '-DUSB_PRODUCT=\"Arduino Leonardo\"' \"/tmp/sketch321469072.cpp\""

	parts, err := utils.ParseCommandLine(command, i18n.HumanLogger{})
	NoError(t, err)

	require.Equal(t, 23, len(parts))

	require.Equal(t, "/home/federico/materiale/works_Arduino/Arduino/build/hardware/tools/coan", parts[0])
	require.Equal(t, "source", parts[1])
	require.Equal(t, "-m", parts[2])
	require.Equal(t, "-E", parts[3])
	require.Equal(t, "-P", parts[4])
	require.Equal(t, "-kb", parts[5])
	require.Equal(t, "-c", parts[6])
	require.Equal(t, "-g", parts[7])
	require.Equal(t, "-Os", parts[8])
	require.Equal(t, "-w", parts[9])
	require.Equal(t, "-ffunction-sections", parts[10])
	require.Equal(t, "-fdata-sections", parts[11])
	require.Equal(t, "-MMD", parts[12])
	require.Equal(t, "-mmcu=atmega32u4", parts[13])
	require.Equal(t, "-DF_CPU=16000000L", parts[14])
	require.Equal(t, "-DARDUINO=010600", parts[15])
	require.Equal(t, "-DARDUINO_AVR_LEONARDO", parts[16])
	require.Equal(t, "-DARDUINO_ARCH_AVR", parts[17])
	require.Equal(t, "-DUSB_VID=0x2341", parts[18])
	require.Equal(t, "-DUSB_PID=0x8036", parts[19])
	require.Equal(t, "-DUSB_MANUFACTURER=", parts[20])
	require.Equal(t, "-DUSB_PRODUCT=\"Arduino Leonardo\"", parts[21])
	require.Equal(t, "/tmp/sketch321469072.cpp", parts[22])
}

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

func TestCommandLineParserError(t *testing.T) {
	command := "\"command missing quote"

	_, err := utils.ParseCommandLine(command, i18n.HumanLogger{})
	require.Error(t, err)
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

func TestParseCppString(t *testing.T) {
	str, rest, ok := utils.ParseCppString(`foo`)
	require.Equal(t, false, ok)

	str, rest, ok = utils.ParseCppString(`"foo`)
	require.Equal(t, false, ok)

	str, rest, ok = utils.ParseCppString(`"foo"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = utils.ParseCppString(`"foo\\bar"`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo\bar`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = utils.ParseCppString(`"foo \"is\" quoted and \\\\bar\"\" escaped\\" and "then" some`)
	require.Equal(t, true, ok)
	require.Equal(t, `foo "is" quoted and \\bar"" escaped\`, str)
	require.Equal(t, ` and "then" some`, rest)

	str, rest, ok = utils.ParseCppString(`" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~"`)
	require.Equal(t, true, ok)
	require.Equal(t, ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = utils.ParseCppString(`"/home/ççç/"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/`, str)
	require.Equal(t, ``, rest)

	str, rest, ok = utils.ParseCppString(`"/home/ççç/ /$sdsdd\\"`)
	require.Equal(t, true, ok)
	require.Equal(t, `/home/ççç/ /$sdsdd\`, str)
	require.Equal(t, ``, rest)
}
