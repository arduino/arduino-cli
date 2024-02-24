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
	require.Zero(t, a.Configs.Size())

	// Allow empty platforms or packages (aka. vendors + architectures)
	b1, err := ParseFQBN("arduino::uno")
	require.Equal(t, "arduino::uno", b1.String())
	require.NoError(t, err)
	require.Equal(t, b1.Package, "arduino")
	require.Equal(t, b1.PlatformArch, "")
	require.Equal(t, b1.BoardID, "uno")
	require.Zero(t, b1.Configs.Size())

	b2, err := ParseFQBN(":avr:uno")
	require.Equal(t, ":avr:uno", b2.String())
	require.NoError(t, err)
	require.Equal(t, b2.Package, "")
	require.Equal(t, b2.PlatformArch, "avr")
	require.Equal(t, b2.BoardID, "uno")
	require.Zero(t, b2.Configs.Size())

	b3, err := ParseFQBN("::uno")
	require.Equal(t, "::uno", b3.String())
	require.NoError(t, err)
	require.Equal(t, b3.Package, "")
	require.Equal(t, b3.PlatformArch, "")
	require.Equal(t, b3.BoardID, "uno")
	require.Zero(t, b3.Configs.Size())

	// Do not allow missing board identifier
	_, err = ParseFQBN("arduino:avr:")
	require.Error(t, err)

	// Do not allow partial fqbn
	_, err = ParseFQBN("arduino")
	require.Error(t, err)
	_, err = ParseFQBN("arduino:avr")
	require.Error(t, err)

	// Keeps the config keys order
	s1, err := ParseFQBN("arduino:avr:uno:d=x,b=x,a=x,e=x,c=x")
	require.NoError(t, err)
	require.Equal(t, "arduino:avr:uno:d=x,b=x,a=x,e=x,c=x", s1.String())
	require.Equal(t,
		"properties.Map{\n  \"d\": \"x\",\n  \"b\": \"x\",\n  \"a\": \"x\",\n  \"e\": \"x\",\n  \"c\": \"x\",\n}",
		s1.Configs.Dump())

	s2, err := ParseFQBN("arduino:avr:uno:a=x,b=x,c=x,d=x,e=x")
	require.NoError(t, err)
	require.Equal(t, "arduino:avr:uno:a=x,b=x,c=x,d=x,e=x", s2.String())
	require.Equal(t,
		"properties.Map{\n  \"a\": \"x\",\n  \"b\": \"x\",\n  \"c\": \"x\",\n  \"d\": \"x\",\n  \"e\": \"x\",\n}",
		s2.Configs.Dump())

	// The config keys order is insignificant when comparing two FQBNs
	require.True(t, s1.Match(s2))
	require.NotEqual(t, s1.String(), s2.String())

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
	_, err = ParseFQBN("arduino:avr:uno,")
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
	require.Equal(t, "arduino:avr:uno:cpu=atmega,speed=1000,extra=core=arduino", f.String())
	require.NoError(t, err)
	require.Equal(t, f.Package, "arduino")
	require.Equal(t, f.PlatformArch, "avr")
	require.Equal(t, f.BoardID, "uno")
	require.Equal(t,
		"properties.Map{\n  \"cpu\": \"atmega\",\n  \"speed\": \"1000\",\n  \"extra\": \"core=arduino\",\n}",
		f.Configs.Dump())
}

func TestMatch(t *testing.T) {
	expectedMatches := [][]string{
		{"arduino:avr:uno", "arduino:avr:uno"},
		{"arduino:avr:uno", "arduino:avr:uno:opt1=1,opt2=2"},
		{"arduino:avr:uno:opt1=1", "arduino:avr:uno:opt1=1,opt2=2"},
		{"arduino:avr:uno:opt1=1,opt2=2", "arduino:avr:uno:opt1=1,opt2=2"},
		{"arduino:avr:uno:opt3=3,opt1=1,opt2=2", "arduino:avr:uno:opt2=2,opt3=3,opt1=1,opt4=4"},
	}

	for _, pair := range expectedMatches {
		a, err := ParseFQBN(pair[0])
		require.NoError(t, err)
		b, err := ParseFQBN(pair[1])
		require.NoError(t, err)
		require.True(t, b.Match(a))
	}

	expectedMismatches := [][]string{
		{"arduino:avr:uno", "arduino:avr:due"},
		{"arduino:avr:uno", "arduino:avr:due:opt1=1,opt2=2"},
		{"arduino:avr:uno:opt1=1", "arduino:avr:uno"},
		{"arduino:avr:uno:opt1=1,opt2=", "arduino:avr:uno:opt1=1,opt2=3"},
		{"arduino:avr:uno:opt1=1,opt2=2", "arduino:avr:uno:opt2=2"},
	}

	for _, pair := range expectedMismatches {
		a, err := ParseFQBN(pair[0])
		require.NoError(t, err)
		b, err := ParseFQBN(pair[1])
		require.NoError(t, err)
		require.False(t, b.Match(a))
	}
}

func TestValidCharacters(t *testing.T) {
	// These FQBNs contain valid characters
	validFqbns := []string{"ardui_no:av_r:un_o", "arduin.o:av.r:un.o", "arduin-o:av-r:un-o", "arduin-o:av-r:un-o:a=b=c=d"}
	for _, fqbn := range validFqbns {
		_, err := ParseFQBN(fqbn)
		require.NoError(t, err)
	}
	// These FQBNs contain invalid characters
	invalidFqbns := []string{"arduin-o:av-r:un=o", "arduin?o:av-r:uno", "arduino:av*r:uno"}
	for _, fqbn := range invalidFqbns {
		_, err := ParseFQBN(fqbn)
		require.Error(t, err)
	}
}
