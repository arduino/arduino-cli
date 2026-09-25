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

//go:build linux

package libraries

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

// newFullyPopulatedLibrary returns a Library with every field set to a
// non-nil/non-empty value, so that a MarshalBinary/UnmarshalBinary round
// trip can be checked field-by-field (and as a whole) without running into
// the nil-vs-empty ambiguities that some of the collection fields have.
func newFullyPopulatedLibrary(prefix *paths.Path, name string) *Library {
	props := properties.NewMap()
	props.Set("name", name)
	props.Set("version", "1.2.3")

	examples := paths.NewPathList()
	examples.Add(prefix.Join(name, "examples", "Blink"))
	examples.Add(paths.New("/opt/outside/examples/Other"))

	return &Library{
		Name:                   name,
		Author:                 "John Doe",
		Maintainer:             "John Doe <john@example.com>",
		Sentence:               "A test library",
		Paragraph:              "This is a longer description of the library.",
		Website:                "https://example.com/" + name,
		Category:               "Other",
		Architectures:          []string{"avr", "esp32"},
		Types:                  []string{"Contributed"},
		InstallDir:             prefix.Join(name),
		DirName:                name,
		SourceDir:              prefix.Join(name, "src"),
		UtilityDir:             paths.New("/opt/outside/utility"),
		Location:               User,
		Layout:                 RecursiveLayout,
		DotALinkage:            true,
		Precompiled:            true,
		PrecompiledWithSources: true,
		LDflags:                "-lm",
		IsLegacy:               false,
		InDevelopment:          true,
		Version:                semver.MustParse("1.2.3"),
		License:                "MIT",
		Properties:             props,
		Examples:               examples,
		declaredHeaders:        []string{name + ".h"},
		sourceHeaders:          []string{name + ".h", "utility.h"},
		CompatibleWith:         map[string]bool{"avr": true, "esp32": false},
	}
}

func TestLibraryMarshalUnmarshalBinary(t *testing.T) {
	prefix := paths.New("/home/build/libraries")
	lib := newFullyPopulatedLibrary(prefix, "MyLib")

	var buf bytes.Buffer
	require.NoError(t, lib.MarshalBinary(&buf, prefix))

	got := &Library{}
	require.NoError(t, got.UnmarshalBinary(&buf, prefix))

	// ContainerPlatform is intentionally not (un)marshaled.
	require.Nil(t, got.ContainerPlatform)
	require.Equal(t, lib, got)

	// The whole message must have been consumed.
	require.Equal(t, 0, buf.Len())
}

func TestLibraryMarshalUnmarshalBinaryPathOutsidePrefix(t *testing.T) {
	// When a path is not inside the given prefix, it must be stored (and
	// restored) as an absolute path instead of a relative one.
	prefix := paths.New("/home/build/libraries")
	lib := newFullyPopulatedLibrary(prefix, "MyLib")
	lib.SourceDir = paths.New("/completely/different/location/src")

	var buf bytes.Buffer
	require.NoError(t, lib.MarshalBinary(&buf, prefix))

	got := &Library{}
	require.NoError(t, got.UnmarshalBinary(&buf, prefix))
	require.Equal(t, lib.SourceDir.String(), got.SourceDir.String())
}

func TestLibraryMarshalUnmarshalBinaryNilAndEmptyFields(t *testing.T) {
	// Document current round-trip behaviour for zero-valued/nil fields:
	// - nil paths round-trip to nil
	// - nil string slices round-trip to nil
	// - a nil CompatibleWith map and a nil Properties map round-trip to
	//   non-nil empty values, since the reader always allocates a map.
	prefix := paths.New("/home/build/libraries")
	lib := &Library{
		Name:    "MinimalLib",
		DirName: "MinimalLib",
		Version: semver.MustParse("0.0.1"),
	}

	var buf bytes.Buffer
	require.NoError(t, lib.MarshalBinary(&buf, prefix))

	got := &Library{}
	require.NoError(t, got.UnmarshalBinary(&buf, prefix))

	require.Equal(t, "MinimalLib", got.Name)
	require.Equal(t, "MinimalLib", got.DirName)
	require.Equal(t, "0.0.1", got.Version.String())

	require.Nil(t, got.InstallDir)
	require.Nil(t, got.SourceDir)
	require.Nil(t, got.UtilityDir)
	require.Nil(t, got.Architectures)
	require.Nil(t, got.Types)
	require.Nil(t, got.declaredHeaders)
	require.Nil(t, got.sourceHeaders)

	require.NotNil(t, got.CompatibleWith)
	require.Empty(t, got.CompatibleWith)
	require.NotNil(t, got.Properties)
	require.Empty(t, got.Properties.Keys())
	require.NotNil(t, got.Examples)
	require.Empty(t, got.Examples)
}

func TestLibraryMarshalBinaryStringTooLong(t *testing.T) {
	prefix := paths.New("/home/build/libraries")
	lib := newFullyPopulatedLibrary(prefix, "MyLib")
	lib.Name = strings.Repeat("a", 1<<16) // one too many for a uint16 length prefix

	var buf bytes.Buffer
	err := lib.MarshalBinary(&buf, prefix)
	require.ErrorContains(t, err, "out of allowed range")
}

func TestListMarshalUnmarshalBinary(t *testing.T) {
	prefix := paths.New("/home/build/libraries")
	list := List{
		newFullyPopulatedLibrary(prefix, "LibOne"),
		newFullyPopulatedLibrary(prefix, "LibTwo"),
	}

	var buf bytes.Buffer
	require.NoError(t, list.MarshalBinary(&buf, prefix))

	var got List
	require.NoError(t, got.UnmarshalBinary(&buf, prefix))

	require.Equal(t, list, got)
	require.Equal(t, 0, buf.Len())
}

func TestListUnmarshalBinaryInvalidMagicNumber(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, binary.Write(&buf, binary.NativeEndian, uint32(0xDEADBEEF)))
	require.NoError(t, binary.Write(&buf, binary.NativeEndian, int32(0)))

	var list List
	err := list.UnmarshalBinary(&buf, paths.New("/home/build/libraries"))
	require.ErrorContains(t, err, "invalid cache version")
}
