/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFQBN(t *testing.T) {
	a, err := ParseFQBN("arduino:avr:uno")
	require.Equal(t, "arduino:avr:uno", a.String())
	require.NoError(t, err)
	require.Equal(t, a.Package, "arduino")
	require.Equal(t, a.PlatformArch, "avr")
	require.Equal(t, a.BoardID, "uno")
	require.Nil(t, a.Configs)

	// Allow empty plaforms or packages
	b1, err := ParseFQBN("arduino::uno")
	require.Equal(t, "arduino::uno", b1.String())
	require.NoError(t, err)
	require.Equal(t, b1.Package, "arduino")
	require.Equal(t, b1.PlatformArch, "")
	require.Equal(t, b1.BoardID, "uno")
	require.Nil(t, b1.Configs)

	b2, err := ParseFQBN(":avr:uno")
	require.Equal(t, ":avr:uno", b2.String())
	require.NoError(t, err)
	require.Equal(t, b2.Package, "")
	require.Equal(t, b2.PlatformArch, "avr")
	require.Equal(t, b2.BoardID, "uno")
	require.Nil(t, b2.Configs)

	b3, err := ParseFQBN("::uno")
	require.Equal(t, "::uno", b3.String())
	require.NoError(t, err)
	require.Equal(t, b3.Package, "")
	require.Equal(t, b3.PlatformArch, "")
	require.Equal(t, b3.BoardID, "uno")
	require.Nil(t, b3.Configs)

	// Do not allow missing board identifier
	_, err = ParseFQBN("arduino:avr:")
	require.Error(t, err)

	// Do not allow partial fqbn
	_, err = ParseFQBN("arduino")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr")
	require.Error(t, err)

	// Sort keys in fbqn config
	s, err := ParseFQBN("arduino:avr:uno:d=x,b=x,a=x,e=x,c=x")
	require.NoError(t, err)
	require.Equal(t, "arduino:avr:uno:a=x,b=x,c=x,d=x,e=x", s.String())

	// Test configs
	c, err := ParseFQBN("arduino:avr:uno:cpu=atmega")
	require.Equal(t, "arduino:avr:uno:cpu=atmega", c.String())
	require.NoError(t, err)
	require.Equal(t, c.Package, "arduino")
	require.Equal(t, c.PlatformArch, "avr")
	require.Equal(t, c.BoardID, "uno")
	require.Equal(t, "properties.Map{\n  \"cpu\": \"atmega\",\n}", c.Configs.Dump())

	d, err := ParseFQBN("arduino:avr:uno:cpu=atmega,speed=1000")
	require.Equal(t, "arduino:avr:uno:cpu=atmega,speed=1000", d.String())
	require.NoError(t, err)
	require.Equal(t, d.Package, "arduino")
	require.Equal(t, d.PlatformArch, "avr")
	require.Equal(t, d.BoardID, "uno")
	require.Equal(t, "properties.Map{\n  \"cpu\": \"atmega\",\n  \"speed\": \"1000\",\n}", d.Configs.Dump())

	// Do not allow empty keys or missing values in config
	_, err = ParseFQBN("arduino:avr:uno:")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr:uno:cpu")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr:uno:=atmega")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr:uno:cpu=atmega,")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr:uno:cpu=atmega,speed")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr:uno:cpu=atmega,=1000")
	require.Error(t, err)

	// Allow keys with empty values
	e, err := ParseFQBN("arduino:avr:uno:cpu=")
	require.Equal(t, "arduino:avr:uno:cpu=", e.String())
	require.NoError(t, err)
	require.Equal(t, e.Package, "arduino")
	require.Equal(t, e.PlatformArch, "avr")
	require.Equal(t, e.BoardID, "uno")
	require.Equal(t, "properties.Map{\n  \"cpu\": \"\",\n}", e.Configs.Dump())

	// Allow "=" in config values
	f, err := ParseFQBN("arduino:avr:uno:cpu=atmega,speed=1000,extra=core=arduino")
	require.Equal(t, "arduino:avr:uno:cpu=atmega,extra=core=arduino,speed=1000", f.String())
	require.NoError(t, err)
	require.Equal(t, f.Package, "arduino")
	require.Equal(t, f.PlatformArch, "avr")
	require.Equal(t, f.BoardID, "uno")
	require.Equal(t, "properties.Map{\n  \"cpu\": \"atmega\",\n  \"extra\": \"core=arduino\",\n  \"speed\": \"1000\",\n}", f.Configs.Dump())
}
