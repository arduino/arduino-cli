// This file is part of arduino-cli.
//
// Copyright (C) Arduino s.r.l. and/or its affiliated companies
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html

package builder

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	semver "go.bug.st/relaxed-semver"

	"github.com/stretchr/testify/require"
)

func TestLibrarySlug(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"PlainName", "Bridge", "BRIDGE"},
		{"AlreadyUpper", "WIFI101", "WIFI101"},
		{"SpacesAndDash", "Bridge Client-2", "BRIDGE_CLIENT_2"},
		{"Dots", "Adafruit.SSD1306", "ADAFRUIT_SSD1306"},
		{"ConsecutiveSeparators", "Foo   Bar--Baz", "FOO_BAR_BAZ"},
		{"LeadingDigit", "101Lib", "_101LIB"},
		{"Empty", "", ""},
		{"OnlySeparators", "---", "_"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, librarySlug(c.in))
		})
	}
}

func TestEncodeLibraryVersion(t *testing.T) {
	cases := []struct {
		name string
		in   *semver.Version
		want uint32
	}{
		{"Full", parseVersion(t, "1.2.3"), 0x010203},
		{"MajorMinorOnly", parseVersion(t, "1.2"), 0x010200},
		{"MajorOnly", parseVersion(t, "1"), 0x010000},
		{"Nil", nil, 1},
		{"EmptyRawVersion", parseVersion(t, ""), 1},
		{"OverflowComponentClamped", parseVersion(t, "300.0.0"), 0xFF0000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, encodeLibraryVersion(c.in))
		})
	}
}

func TestLibrarySourceSuffix(t *testing.T) {
	cases := []struct {
		name string
		in   libraries.LibraryLocation
		want string
	}{
		{"User", libraries.User, "IN_SKETCHBOOK"},
		{"PlatformBuiltIn", libraries.PlatformBuiltIn, "IN_PLATFORM"},
		{"ReferencedPlatformBuiltIn", libraries.ReferencedPlatformBuiltIn, "IN_PLATFORM"},
		{"Profile", libraries.Profile, "IN_PROFILE"},
		{"IDEBuiltIn", libraries.IDEBuiltIn, "IN_IDE"},
		{"Unmanaged", libraries.Unmanaged, "IN_SPECIFIED_PATH"},
		{"OutOfRange", libraries.LibraryLocation(99), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, librarySourceSuffix(c.in))
		})
	}
}

func TestBuildLibraryDiscoveryFlags(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		require.Equal(t, "", buildLibraryDiscoveryFlags(libraries.List{}))
	})

	t.Run("SingleLibrary", func(t *testing.T) {
		libs := libraries.List{
			{Name: "Bridge", Version: parseVersion(t, "1.2.3"), Location: libraries.User},
		}
		require.Equal(t,
			"-DFOUND_BRIDGE_LIB=0x010203 -DFOUND_BRIDGE_LIB_IN_SKETCHBOOK=1",
			buildLibraryDiscoveryFlags(libs))
	})

	t.Run("MultipleLibrariesJoinedWithSpace", func(t *testing.T) {
		libs := libraries.List{
			{Name: "Bridge", Version: parseVersion(t, "1.2.3"), Location: libraries.User},
			{Name: "Bridge Client-2", Version: nil, Location: libraries.PlatformBuiltIn},
		}
		require.Equal(t,
			"-DFOUND_BRIDGE_LIB=0x010203 -DFOUND_BRIDGE_LIB_IN_SKETCHBOOK=1 -DFOUND_BRIDGE_CLIENT_2_LIB=0x000001 -DFOUND_BRIDGE_CLIENT_2_LIB_IN_PLATFORM=1",
			buildLibraryDiscoveryFlags(libs))
	})

	t.Run("UnrecognizedLocationGetsNoSourceMacro", func(t *testing.T) {
		libs := libraries.List{
			{Name: "Bridge", Version: parseVersion(t, "1.2.3"), Location: libraries.LibraryLocation(99)},
		}
		require.Equal(t, "-DFOUND_BRIDGE_LIB=0x010203", buildLibraryDiscoveryFlags(libs))
	})
}

func parseVersion(t *testing.T, s string) *semver.Version {
	t.Helper()
	v, err := semver.Parse(s)
	require.NoError(t, err)
	return v
}
