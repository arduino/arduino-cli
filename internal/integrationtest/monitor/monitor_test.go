// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package monitor_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestMonitorConfigFlags(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install AVR platform (this is required to enable the 'serial' monitor...)
	// TODO: maybe this is worth opening an issue?
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Install mocked discovery and monitor for testing
	require.NoError(t, err)
	cli.InstallMockedSerialDiscovery(t)
	cli.InstallMockedSerialMonitor(t)

	// Test monitor command
	quit := func() io.Reader {
		// tells mocked monitor to exit
		return bytes.NewBufferString("QUIT\n")
	}

	t.Run("NoArgs", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quit(), "monitor", "-p", "/dev/ttyARG", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 9600")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("BaudConfig", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quit(), "monitor", "-p", "/dev/ttyARG", "-c", "baudrate=115200", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 115200")
		require.Contains(t, string(stdout), "Configuration parity = none")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("BaudAndParitfyConfig", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quit(), "monitor", "-p", "/dev/ttyARG",
			"-c", "baudrate=115200", "-c", "parity=even", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 115200")
		require.Contains(t, string(stdout), "Configuration parity = even")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("InvalidConfigKey", func(t *testing.T) {
		_, stderr, err := cli.RunWithCustomInput(quit(), "monitor", "-p", "/dev/ttyARG",
			"-c", "baud=115200", "-c", "parity=even", "--raw")
		require.Error(t, err)
		require.Contains(t, string(stderr), "invalid port configuration: baud=115200")
	})

	t.Run("InvalidConfigValue", func(t *testing.T) {
		_, stderr, err := cli.RunWithCustomInput(quit(), "monitor", "-p", "/dev/ttyARG",
			"-c", "parity=9600", "--raw")
		require.Error(t, err)
		require.Contains(t, string(stderr), "invalid port configuration value for parity: 9600")
	})
}
