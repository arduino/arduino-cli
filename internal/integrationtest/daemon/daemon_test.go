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

package daemon_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testsuite"
)

// createEnvForDaemon performs the minimum required operations to start the arduino-cli daemon.
// It returns a testsuite.Environment and an ArduinoCLI client to perform the integration tests.
// The Environment must be disposed by calling the CleanUp method via defer.
func createEnvForDaemon(t *testing.T) (*testsuite.Environment, *integrationtest.ArduinoCLI) {
	env := testsuite.NewEnvironment(t)

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         paths.New("..", "..", "..", "arduino-cli"),
		UseSharedStagingFolder: true,
	})

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_ = cli.StartDaemon(false)
	return env, cli
}
