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
