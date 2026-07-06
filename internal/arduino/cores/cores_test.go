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
	semver "go.bug.st/relaxed-semver"
)

func TestRequiresToolRelease(t *testing.T) {
	toolDependencyName := "avr-gcc"
	toolDependencyVersion := "7.3.0-atmel3.6.1-arduino7"
	toolDependencyPackager := "arduino"

	release := PlatformRelease{
		ToolDependencies: ToolDependencies{
			{
				ToolName:     toolDependencyName,
				ToolVersion:  semver.ParseRelaxed(toolDependencyVersion),
				ToolPackager: toolDependencyPackager,
			},
		},
	}

	toolRelease := &ToolRelease{
		Version: semver.ParseRelaxed(toolDependencyVersion + "not"),
		Tool: &Tool{
			Name: toolDependencyName + "not",
			Package: &Package{
				Name: toolDependencyPackager + "not",
			},
		},
	}

	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Tool.Name = toolDependencyName
	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Tool.Package.Name = toolDependencyPackager
	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Version = semver.ParseRelaxed(toolDependencyVersion)
	require.True(t, release.RequiresToolRelease(toolRelease))
}

func TestRequiresToolReleaseDiscovery(t *testing.T) {
	toolDependencyName := "ble-discovery"
	toolDependencyPackager := "arduino"

	release := PlatformRelease{
		DiscoveryDependencies: DiscoveryDependencies{
			{
				Name:     toolDependencyName,
				Packager: toolDependencyPackager,
			},
		},
	}

	toolRelease := &ToolRelease{
		Version: semver.ParseRelaxed("0.1.0"),
		Tool: &Tool{
			Name: toolDependencyName + "not",
			Releases: map[semver.NormalizedString]*ToolRelease{
				"1.0.0": {Version: semver.ParseRelaxed("1.0.0")},
				"0.1.0": {Version: semver.ParseRelaxed("0.1.0")},
				"0.0.1": {Version: semver.ParseRelaxed("0.0.1")},
			},
			Package: &Package{
				Name: toolDependencyPackager + "not",
			},
		},
	}

	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Tool.Name = toolDependencyName
	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Tool.Package.Name = toolDependencyPackager
	require.False(t, release.RequiresToolRelease(toolRelease))
	toolRelease.Version = semver.ParseRelaxed("1.0.0")
	require.True(t, release.RequiresToolRelease(toolRelease))
}

func TestGetLatestCompatibleIndexedRelease(t *testing.T) {
	platform := &Platform{
		Architecture: "avr",
		Releases:     map[semver.NormalizedString]*PlatformRelease{},
	}
	addRelease := func(version string, compatible, indexed bool) {
		v := semver.MustParse(version)
		platform.Releases[v.NormalizedString()] = &PlatformRelease{
			Version:    v,
			Platform:   platform,
			Compatible: compatible,
			Indexed:    indexed,
		}
	}

	// Public indexed releases up to 1.8.4, plus a newer local-only version 2.0.0
	// (e.g. installed from a removed additional-url).
	addRelease("1.8.3", true, true)
	addRelease("1.8.4", true, true)
	addRelease("2.0.0", true, false)

	// The overall latest (including local) is 2.0.0.
	require.Equal(t, "2.0.0", platform.GetLatestCompatibleRelease().Version.String())

	// Excluding local-only versions, the latest is 1.8.4.
	require.Equal(t, "1.8.4", platform.GetLatestCompatibleIndexedRelease().Version.String())

	// An incompatible indexed release must be ignored.
	addRelease("1.9.0", false, true)
	require.Equal(t, "1.8.4", platform.GetLatestCompatibleIndexedRelease().Version.String())

	// If no indexed release is available, the result is nil.
	onlyLocal := &Platform{
		Architecture: "avr",
		Releases:     map[semver.NormalizedString]*PlatformRelease{},
	}
	lv := semver.MustParse("3.0.0")
	onlyLocal.Releases[lv.NormalizedString()] = &PlatformRelease{Version: lv, Platform: onlyLocal, Compatible: true, Indexed: false}
	require.Nil(t, onlyLocal.GetLatestCompatibleIndexedRelease())
}
