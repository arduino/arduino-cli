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
	_, _, err := cli.Run("core", "update-idex")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "update-idex")
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
	require.Equal(t, string(stdout), "\n")
}
