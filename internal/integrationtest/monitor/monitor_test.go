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
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// returns a reader that tells the mocked monitor to exit
func quitMonitor() io.Reader {
	// tells mocked monitor to exit
	return bytes.NewBufferString("QUIT\n")
}

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

	t.Run("NoArgs", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARG", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 9600")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("BaudConfig", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARG", "-c", "baudrate=115200", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 115200")
		require.Contains(t, string(stdout), "Configuration parity = none")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("BaudAndParitfyConfig", func(t *testing.T) {
		stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARG",
			"-c", "baudrate=115200", "-c", "parity=even", "--raw")
		require.NoError(t, err)
		require.Contains(t, string(stdout), "Opened port: /dev/ttyARG")
		require.Contains(t, string(stdout), "Configuration baudrate = 115200")
		require.Contains(t, string(stdout), "Configuration parity = even")
		require.Contains(t, string(stdout), "Configuration rts = on")
		require.Contains(t, string(stdout), "Configuration dtr = on")
	})

	t.Run("InvalidConfigKey", func(t *testing.T) {
		_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARG",
			"-c", "baud=115200", "-c", "parity=even", "--raw")
		require.Error(t, err)
		require.Contains(t, string(stderr), "invalid port configuration: baud=115200")
	})

	t.Run("InvalidConfigValue", func(t *testing.T) {
		_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARG",
			"-c", "parity=9600", "--raw")
		require.Error(t, err)
		require.Contains(t, string(stderr), "invalid port configuration value for parity: 9600")
	})
}

func TestMonitorCommandFlagsAndDefaultPortFQBNSelection(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install AVR platform
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Patch the Yun board to require special RTS/DTR serial configuration
	f, err := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.6", "boards.txt").Append()
	require.NoError(t, err)
	_, err = f.WriteString(`
uno.serial.disableRTS=true
uno.serial.disableDTR=false
yun.serial.disableRTS=true
yun.serial.disableDTR=true
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Install mocked discovery and monitor for testing
	cli.InstallMockedSerialDiscovery(t)
	cli.InstallMockedSerialMonitor(t)

	// Create test sketches
	getSketchPath := func(sketch string) string {
		p, err := paths.New("testdata", sketch).Abs()
		require.NoError(t, err)
		require.True(t, p.IsDir())
		return p.String()
	}
	sketch := getSketchPath("SketchWithNoProfiles")
	sketchWithPort := getSketchPath("SketchWithDefaultPort")
	sketchWithFQBN := getSketchPath("SketchWithDefaultFQBN")
	sketchWithPortAndFQBN := getSketchPath("SketchWithDefaultPortAndFQBN")
	sketchWithPortAndConfig := getSketchPath("SketchWithDefaultPortAndConfig")
	sketchWithPortAndConfigAndProfile := getSketchPath("SketchWithDefaultPortAndConfigAndProfile")

	t.Run("NoFlags", func(t *testing.T) {
		t.Run("NoDefaultPortNoDefaultFQBN", func(t *testing.T) {
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketch)
			require.Error(t, err)
			require.Contains(t, string(stderr), "No monitor available for the port protocol default")
		})

		t.Run("WithDefaultPort", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketchWithPort)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultFQBN", func(t *testing.T) {
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketchWithFQBN)
			require.Error(t, err)
			require.Contains(t, string(stderr), "No monitor available for the port protocol default")
		})

		t.Run("WithDefaultPortAndQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = off")
		})

		t.Run("FQBNFromSpecificProfile", func(t *testing.T) {
			// The only way to assert we're picking up the fqbn specified from the profile is to provide a wrong value
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", "--profile", "profile1", sketchWithPortAndFQBN)
			require.Error(t, err)
			require.Contains(t, string(stderr), "not an FQBN: broken_fqbn")
		})

		t.Run("WithDefaultPortAndConfig", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketchWithPortAndConfig)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("WithDefaultPortAndConfigAndProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})
	})

	t.Run("WithPortFlag", func(t *testing.T) {
		t.Run("NoDefaultPortNoDefaultFQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketch)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultPort", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketchWithPort)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultFQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketchWithFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = off")
		})

		t.Run("WithDefaultPortAndQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = off")
		})

		t.Run("FQBNFromSpecificProfile", func(t *testing.T) {
			// The only way to assert we're picking up the fqbn specified from the profile is to provide a wrong value
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", "--profile", "profile1", sketchWithPortAndFQBN)
			require.Error(t, err)
			require.Contains(t, string(stderr), "not an FQBN: broken_fqbn")
		})

		t.Run("WithDefaultPortAndConfig", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketchWithPortAndConfig)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("WithDefaultPortAndConfigAndProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = on")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})
	})

	t.Run("WithFQBNFlag", func(t *testing.T) {
		t.Run("NoDefaultPortNoDefaultFQBN", func(t *testing.T) {
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", sketch)
			require.Error(t, err)
			require.Contains(t, string(stderr), "No monitor available for the port protocol default")
		})

		t.Run("WithDefaultPort", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", sketchWithPort)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultFQBN", func(t *testing.T) {
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", sketchWithFQBN)
			require.Error(t, err)
			require.Contains(t, string(stderr), "No monitor available for the port protocol default")
		})

		t.Run("WithDefaultPortAndQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("IgnoreProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", "--profile", "profile1", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyDEF")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-downloaded platform that is not patched for test
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultPortAndConfig", func(t *testing.T) {
			_, stderr, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", "--profile", "profile1", sketchWithPortAndConfig)
			require.Error(t, err)
			require.Contains(t, string(stderr), "Profile 'profile1' not found")
			require.Contains(t, string(stderr), "Unknown FQBN: unknown package arduino")
		})

		t.Run("WithDefaultPortAndConfigAndProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:uno", "--raw", "--profile", "uno", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyPROF")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-downloaded platform that is not patched for test
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 19200")
			require.Contains(t, string(stdout), "Configuration bits = 8")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})
	})

	t.Run("WithPortAndFQBNFlags", func(t *testing.T) {
		t.Run("NoDefaultPortNoDefaultFQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketch)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultPort", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketchWithPort)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultFQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketchWithFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultPortAndQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})

		t.Run("WithDefaultPortAndConfig", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketchWithPortAndConfig)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("WithDefaultPortAndConfigAndProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = off")
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 57600")
			require.Contains(t, string(stdout), "Configuration bits = 9")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("IgnoreProfile", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-p", "/dev/ttyARGS", "-b", "arduino:avr:uno", "--raw", "--profile", "profile1", sketchWithPortAndFQBN)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyARGS")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-downloaded platform that is not patched for test
			require.Contains(t, string(stdout), "Configuration dtr = on")
		})
	})

	t.Run("WithProfileFlags", func(t *testing.T) {
		t.Run("NoOtherArgs", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-m", "uno", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyPROF")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-installed AVR core (not patched by this test)
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 19200")
			require.Contains(t, string(stdout), "Configuration bits = 8")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("WithFQBN", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-b", "arduino:avr:yun", "-m", "uno", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyPROF")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-installed AVR core (not patched by this test)
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 19200")
			require.Contains(t, string(stdout), "Configuration bits = 8")
			require.Contains(t, string(stdout), "Configuration parity = none")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})

		t.Run("WithConfigFlag", func(t *testing.T) {
			stdout, _, err := cli.RunWithCustomInput(quitMonitor(), "monitor", "-c", "odd", "-m", "uno", "--raw", sketchWithPortAndConfigAndProfile)
			require.NoError(t, err)
			require.Contains(t, string(stdout), "Opened port: /dev/ttyPROF")
			require.Contains(t, string(stdout), "Configuration rts = on") // This is taken from profile-installed AVR core (not patched by this test)
			require.Contains(t, string(stdout), "Configuration dtr = on")
			require.Contains(t, string(stdout), "Configuration baudrate = 19200")
			require.Contains(t, string(stdout), "Configuration bits = 8")
			require.Contains(t, string(stdout), "Configuration parity = odd")
			require.Contains(t, string(stdout), "Configuration stop_bits = 1")
		})
	})
}
