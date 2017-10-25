/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2017 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
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
 */

package properties

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropertiesBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "boards.txt"))

	require.NoError(t, err)

	require.Equal(t, "Processor", p["menu.cpu"])
	require.Equal(t, "32256", p["ethernet.upload.maximum_size"])
	require.Equal(t, "{build.usb_flags}", p["robotMotor.build.extra_flags"])

	ethernet := p.SubTree("ethernet")
	require.Equal(t, "Arduino Ethernet", ethernet["name"])
}

func TestPropertiesTestTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "test.txt"))

	require.NoError(t, err)

	require.Equal(t, 4, len(p))
	require.Equal(t, "value = 1", p["key"])

	switch value := runtime.GOOS; value {
	case "linux":
		require.Equal(t, "is linux", p["which.os"])
	case "windows":
		require.Equal(t, "is windows", p["which.os"])
	case "darwin":
		require.Equal(t, "is macosx", p["which.os"])
	default:
		require.FailNow(t, "unsupported OS")
	}
}

func TestExpandPropsInString(t *testing.T) {
	aMap := make(Map)
	aMap["key1"] = "42"
	aMap["key2"] = "{key1}"

	str := "{key1} == {key2} == true"

	str = aMap.ExpandPropsInString(str)
	require.Equal(t, "42 == 42 == true", str)
}

func TestExpandPropsInString2(t *testing.T) {
	p := make(Map)
	p["key2"] = "{key2}"
	p["key1"] = "42"

	str := "{key1} == {key2} == true"

	str = p.ExpandPropsInString(str)
	require.Equal(t, "42 == {key2} == true", str)
}

func TestDeleteUnexpandedPropsFromString(t *testing.T) {
	p := make(Map)
	p["key1"] = "42"
	p["key2"] = "{key1}"

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, "42 == 42 ==  == true", str)
}

func TestDeleteUnexpandedPropsFromString2(t *testing.T) {
	p := make(Map)
	p["key2"] = "42"

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, " == 42 ==  == true", str)
}

func TestPropertiesRedBeearLabBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "redbearlab_boards.txt"))

	require.NoError(t, err)

	require.Equal(t, 83, len(p))
	require.Equal(t, "Blend", p["blend.name"])
	require.Equal(t, "arduino:arduino", p["blend.build.core"])
	require.Equal(t, "0x2404", p["blendmicro16.pid.0"])

	ethernet := p.SubTree("blend")
	require.Equal(t, "arduino:arduino", ethernet["build.core"])
}

func TestPropertiesBroken(t *testing.T) {
	_, err := Load(filepath.Join("testdata", "broken.txt"))

	require.Error(t, err)
}
