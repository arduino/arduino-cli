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

package update_test

import (
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, err := cli.Run("update")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Downloading index: package_index.tar.bz2 downloaded")
	require.Contains(t, string(stdout), "Downloading index: library_index.tar.bz2 downloaded")
}

func TestUpdateShowingOutdated(t *testing.T) {
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

	// Installs latest version of a core and a library
	_, _, err = cli.Run("core", "install", "arduino:samd")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "ArduinoJson")
	require.NoError(t, err)

	// Verifies outdated cores and libraries are printed after updating indexes
	stdout, _, err := cli.Run("update", "--show-outdated")
	require.NoError(t, err)
	lines := strings.Split(string(stdout), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	require.Contains(t, lines[0], "Downloading index: library_index.tar.bz2 downloaded")
	require.Contains(t, lines[1], "Downloading index: package_index.tar.bz2 downloaded")
	require.Contains(t, lines[3], "Arduino AVR Boards")
	require.Contains(t, lines[4], "USBHost")
}

func TestUpdateWithUrlNotFound(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Brings up a local server to fake a failure
	url := env.HTTPServeFileError(8000, paths.New("test_index.json"), 404)

	stdout, _, err := cli.Run("update", "--additional-urls="+url.String())
	require.Error(t, err)
	require.Contains(t, string(stdout), "Downloading index: test_index.json Server responded with: 404 Not Found")
}

func TestUpdateWithUrlInternalServerError(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Brings up a local server to fake a failure
	url := env.HTTPServeFileError(8000, paths.New("test_index.json"), 500)

	stdout, _, err := cli.Run("update", "--additional-urls="+url.String())
	require.Error(t, err)
	require.Contains(t, string(stdout), "Downloading index: test_index.json Server responded with: 500 Internal Server Error")
}

func TestUpdateShowingOutdatedUsingLibraryWithInvalidVersion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Install latest version of a library
	_, _, err = cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)

	// Verifies library doesn't get updated
	stdout, _, err := cli.Run("update", "--show-outdated")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "WiFi101")

	// Changes the version of the currently installed library so that it's
	// invalid
	libPath := cli.SketchbookDir().Join("libraries", "WiFi101", "library.properties")
	require.NoError(t, libPath.WriteFile([]byte("name=WiFi101\nversion=1.0001")))

	// Verifies library gets updated
	stdout, _, err = cli.Run("update", "--show-outdated")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "WiFi101")
}
