// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package upgrade_test

import (
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Updates index for cores and libraries
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-index")
	require.NoError(t, err)

	// Installs an outdated core and library
	_, _, err = cli.Run("core", "install", "arduino:avr@1.6.3")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "USBHost@1.0.0")
	require.NoError(t, err)

	// Installs an outdated core and library
	_, _, err = cli.Run("core", "install", "arduino:samd")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "ArduinoJson")
	require.NoError(t, err)

	// Verifies outdated core and libraries are shown
	stdout, _, err := cli.Run("outdated")
	require.NoError(t, err)
	lines := strings.Split(string(stdout), "\n")
	require.Contains(t, lines[1], "Arduino AVR Boards")
	require.Contains(t, lines[4], "USBHost")

	_, _, err = cli.Run("upgrade")
	require.NoError(t, err)

	// Verifies cores and libraries have been updated
	stdout, _, err = cli.Run("outdated")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "No libraries update is available.")
}

func TestUpgradeUsingLibraryWithInvalidVersion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install latest version of a library
	_, _, err = cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)

	// Verifies library is not shown
	stdout, _, err := cli.Run("outdated")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "WiFi101")

	// Changes the version of the currently installed library so that it's invalid
	libPropPath := cli.SketchbookDir().Join("libraries", "WiFi101", "library.properties")
	err = libPropPath.WriteFile([]byte("name=WiFi101\nversion=1.0001"))
	require.NoError(t, err)

	// Verifies library gets upgraded
	stdout, _, err = cli.Run("upgrade")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "WiFi101")
}

func TestUpgradeUnusedCoreToolsAreRemoved(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Installs a core
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.2")
	require.NoError(t, err)

	// Verifies expected tool is installed
	toolPath := cli.DataDir().Join("packages", "arduino", "tools", "avr-gcc", "7.3.0-atmel3.6.1-arduino5")
	require.DirExists(t, toolPath.String())

	// Upgrades everything
	_, _, err = cli.Run("upgrade")
	require.NoError(t, err)

	// Verifies tool is uninstalled since it's not used by newer core version
	require.NoDirExists(t, toolPath.String())
}
