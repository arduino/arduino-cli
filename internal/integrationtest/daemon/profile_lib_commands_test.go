// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package daemon

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func indexLibArray(l ...*commands.ProfileLibraryReference) []*commands.ProfileLibraryReference {
	return l
}

func indexLib(name, version string, isdep ...bool) *commands.ProfileLibraryReference {
	return &commands.ProfileLibraryReference{
		Library: &commands.ProfileLibraryReference_IndexLibrary_{
			IndexLibrary: &commands.ProfileLibraryReference_IndexLibrary{
				Name:         name,
				Version:      version,
				IsDependency: len(isdep) > 0 && isdep[0],
			},
		},
	}
}

func TestProfileLibAddListAndRemov(t *testing.T) {
	env, cli := integrationtest.CreateEnvForDaemon(t)
	t.Cleanup(func() { env.CleanUp() })

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	t.Cleanup(func() { tmp.RemoveAll() })
	sk := tmp.Join("sketch")

	// Create a new sketch
	_, _, err = cli.Run("sketch", "new", sk.String())
	require.NoError(t, err)

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Create a new profile
	resp, err := grpcInst.ProfileCreate(t.Context(), "test", sk.String(), "arduino:avr:uno", true)
	require.NoError(t, err)
	projectFile := paths.New(resp.GetProjectFilePath())

	expect := func(expected string) {
		p, _ := projectFile.ReadFile()
		require.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(p)))
	}
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	_, err = grpcInst.ProfileCreate(t.Context(), "test2", sk.String(), "arduino:avr:mini", false)
	require.NoError(t, err)
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	// Add a library to the profile
	{
		addresp, err := grpcInst.ProfileLibAdd(t.Context(), sk.String(), "test", indexLib("ArduinoJson", "6.18.5"), true, false)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(indexLib("ArduinoJson", "6.18.5")), addresp.GetAddedLibraries())
	}
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	// Add a library with deps to the profile
	{
		addresp, err := grpcInst.ProfileLibAdd(t.Context(), sk.String(), "test", indexLib("Adafruit 9DOF", "1.1.4"), true, false)
		require.NoError(t, err)
		expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit 9DOF (1.1.4)
      - dependency: Adafruit L3GD20 U (2.0.3)
      - dependency: Adafruit LSM303DLHC (1.0.4)
      - dependency: Adafruit Unified Sensor (1.1.15)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)
		require.Equal(t, indexLibArray(
			indexLib("Adafruit 9DOF", "1.1.4"),
			indexLib("Adafruit L3GD20 U", "2.0.3", true),
			indexLib("Adafruit LSM303DLHC", "1.0.4", true),
			indexLib("Adafruit Unified Sensor", "1.1.15", true),
		), addresp.GetAddedLibraries())
	}

	{
		// Add a library with deps to the profile
		addresp, err := grpcInst.ProfileLibAdd(t.Context(), sk.String(), "test", indexLib("Adafruit ADG72x", "1.0.0"), true, false)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(
			indexLib("Adafruit ADG72x", "1.0.0"),
			indexLib("Adafruit BusIO", "1.17.4", true),
		), addresp.GetAddedLibraries())
	}
	{
		// Add a library with deps to the profile
		addresp, err := grpcInst.ProfileLibAdd(t.Context(), sk.String(), "test", indexLib("Adafruit ADS1X15", "2.6.0"), true, false)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(indexLib("Adafruit ADS1X15", "2.6.0")), addresp.GetAddedLibraries())
	}
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit 9DOF (1.1.4)
      - dependency: Adafruit L3GD20 U (2.0.3)
      - dependency: Adafruit LSM303DLHC (1.0.4)
      - dependency: Adafruit Unified Sensor (1.1.15)
      - Adafruit ADG72x (1.0.0)
      - dependency: Adafruit BusIO (1.17.4)
      - Adafruit ADS1X15 (2.6.0)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	// Remove a library with deps from the profile
	{
		remresp, err := grpcInst.ProfileLibRemove(t.Context(), sk.String(), "test", indexLib("Adafruit ADG72x", "1.0.0"), true)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(indexLib("Adafruit ADG72x", "1.0.0")), remresp.RemovedLibraries)
	}
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit 9DOF (1.1.4)
      - dependency: Adafruit L3GD20 U (2.0.3)
      - dependency: Adafruit LSM303DLHC (1.0.4)
      - dependency: Adafruit Unified Sensor (1.1.15)
      - dependency: Adafruit BusIO (1.17.4)
      - Adafruit ADS1X15 (2.6.0)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	// Remove another library with deps from the profile that will also remove some shared dependencies
	{
		remresp, err := grpcInst.ProfileLibRemove(t.Context(), sk.String(), "test", indexLib("Adafruit ADS1X15", "2.6.0"), true)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(
			indexLib("Adafruit ADS1X15", "2.6.0"),
			indexLib("Adafruit BusIO", "1.17.4", true),
		), remresp.RemovedLibraries)
	}
	expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit 9DOF (1.1.4)
      - dependency: Adafruit L3GD20 U (2.0.3)
      - dependency: Adafruit LSM303DLHC (1.0.4)
      - dependency: Adafruit Unified Sensor (1.1.15)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)

	// Now explicitly add a dependency making it no longer a (removable) dependency
	{
		addresp, err := grpcInst.ProfileLibAdd(t.Context(), sk.String(), "test", indexLib("Adafruit Unified Sensor", "1.1.15"), true, false)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(indexLib("Adafruit Unified Sensor", "1.1.15")), addresp.GetSkippedLibraries())
		expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit 9DOF (1.1.4)
      - dependency: Adafruit L3GD20 U (2.0.3)
      - dependency: Adafruit LSM303DLHC (1.0.4)
      - Adafruit Unified Sensor (1.1.15)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)
	}

	// Try to remove the main library again, the explicitly added dependency should remain
	{
		remresp, err := grpcInst.ProfileLibRemove(t.Context(), sk.String(), "test", indexLib("Adafruit 9DOF", "1.1.4"), true)
		require.NoError(t, err)
		require.Equal(t, indexLibArray(
			indexLib("Adafruit 9DOF", "1.1.4"),
			indexLib("Adafruit L3GD20 U", "2.0.3", true),
			indexLib("Adafruit LSM303DLHC", "1.0.4", true),
		), remresp.RemovedLibraries)
		expect(`
profiles:
  test:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - ArduinoJson (6.18.5)
      - Adafruit Unified Sensor (1.1.15)

  test2:
    fqbn: arduino:avr:mini
    platforms:
      - platform: arduino:avr (1.8.6)

default_profile: test
`)
	}
}
