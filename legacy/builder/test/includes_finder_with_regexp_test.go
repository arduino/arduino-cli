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
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIncludesFinderWithRegExp(t *testing.T) {
	ctx := &types.Context{}

	output := "/some/path/sketch.ino:1:17: fatal error: SPI.h: No such file or directory\n" +
		"#include <SPI.h>\n" +
		"^\n" +
		"compilation terminated."
	include := builder.IncludesFinderWithRegExp(ctx, output)

	require.Equal(t, "SPI.h", include)
}

func TestIncludesFinderWithRegExpEmptyOutput(t *testing.T) {
	ctx := &types.Context{}

	include := builder.IncludesFinderWithRegExp(ctx, "")

	require.Equal(t, "", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes(t *testing.T) {
	ctx := &types.Context{}

	output := "/some/path/sketch.ino:1:33: fatal error: Wire.h: No such file or directory\n" +
		" #               include <Wire.h>\n" +
		"                                 ^\n" +
		"compilation terminated.\n"
	include := builder.IncludesFinderWithRegExp(ctx, output)

	require.Equal(t, "Wire.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes2(t *testing.T) {
	ctx := &types.Context{}

	output := "/some/path/sketch.ino:1:33: fatal error: Wire.h: No such file or directory\n" +
		" #\t\t\tinclude <Wire.h>\n" +
		"                                 ^\n" +
		"compilation terminated.\n"
	include := builder.IncludesFinderWithRegExp(ctx, output)

	require.Equal(t, "Wire.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes3(t *testing.T) {
	ctx := &types.Context{}

	output := "/some/path/sketch.ino:1:33: fatal error: SPI.h: No such file or directory\n" +
		"compilation terminated.\n"

	include := builder.IncludesFinderWithRegExp(ctx, output)

	require.Equal(t, "SPI.h", include)
}

func TestIncludesFinderWithRegExpPaddedIncludes4(t *testing.T) {
	ctx := &types.Context{}

	output := "In file included from /tmp/arduino_modified_sketch_815412/binouts.ino:52:0:\n" +
		"/tmp/arduino_build_static/sketch/regtable.h:31:22: fatal error: register.h: No such file or directory\n"

	include := builder.IncludesFinderWithRegExp(ctx, output)

	require.Equal(t, "register.h", include)
}
