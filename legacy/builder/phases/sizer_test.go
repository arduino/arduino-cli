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
 * Copyright 2016 Arduino LLC (http://www.arduino.cc/)
 */

package phases

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSizerWithAVRData(t *testing.T) {
	output := []byte(`/tmp/test597119152/sketch.ino.elf  :
section           size      addr
.data               36   8388864
.text             3966         0
.bss               112   8388900
.comment            17         0
.debug_aranges     704         0
.debug_info      21084         0
.debug_abbrev     4704         0
.debug_line       5456         0
.debug_frame      1964         0
.debug_str        8251         0
.debug_loc        7747         0
.debug_ranges      856         0
Total            54897
`)

	size, err := computeSize(`^(?:\.text|\.data|\.bootloader)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 4002, size)

	size, err = computeSize(`^(?:\.data|\.bss|\.noinit)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 148, size)

	size, err = computeSize(`^(?:\.eeprom)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 0, size)
}

func TestSizerWithSAMDData(t *testing.T) {
	output := []byte(`/tmp/test737785204/sketch_usbhost.ino.elf  :
section             size        addr
.text               8076        8192
.data                152   536870912
.bss                1984   536871064
.ARM.attributes       40           0
.comment             128           0
.debug_info       178841           0
.debug_abbrev      14968           0
.debug_aranges      2080           0
.debug_ranges       3840           0
.debug_line        26068           0
.debug_str         55489           0
.debug_frame        5036           0
.debug_loc         20978           0
Total             317680
`)

	size, err := computeSize(`\.text\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 8076, size)
}

func TestSizerEmptyRegexpReturnsMinusOne(t *testing.T) {
	size, err := computeSize(``, []byte(`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`))
	require.NoError(t, err)
	require.Equal(t, -1, size)
}

func TestSizerWithInvalidRegexp(t *testing.T) {
	_, err := computeSize(`[xx`, []byte(`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`))
	require.Error(t, err)
}
