// This file is part of arduino-cli.
//
// Copyright 2022-2025 ARDUINO SA (http://www.arduino.cc/)
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

package sketch_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestSketchProfileDump(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Prepare the sketch with libraries
	tmpDir, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = tmpDir.RemoveAll })

	sketchTemplate, err := paths.New("testdata", "SketchWithLibrary").Abs()
	require.NoError(t, err)

	sketch := tmpDir.Join("SketchWithLibrary")
	libInside := sketch.Join("libraries", "MyLib")
	err = sketchTemplate.CopyDirTo(sketch)
	require.NoError(t, err)

	libOutsideTemplate := sketchTemplate.Join("..", "MyLibOutside")
	libOutside := sketch.Join("..", "MyLibOutside")
	err = libOutsideTemplate.CopyDirTo(libOutside)
	require.NoError(t, err)

	// Install the required core and libraries
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit BusIO@1.17.1", "--no-overwrite")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit GFX Library@1.12.1", "--no-overwrite")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit SSD1306@2.5.14", "--no-overwrite")
	require.NoError(t, err)

	// Check if the profile dump:
	// - keeps libraries in the sketch with a relative path
	// - keeps libraries outside the sketch with an absolute path
	// - keeps libraries installed in the system with just the name and version
	out, _, err := cli.Run("compile", "-b", "arduino:avr:uno",
		"--library", libInside.String(),
		"--library", libOutside.String(),
		"--dump-profile",
		sketch.String())
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(`
profiles:
  uno:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - dir: libraries/MyLib
      - dir: `+libOutside.String()+`
      - Adafruit SSD1306 (2.5.14)
      - Adafruit GFX Library (1.12.1)
      - Adafruit BusIO (1.17.1)
`), strings.TrimSpace(string(out)))

	// Dump the profile in the sketch directory and compile with it again
	err = sketch.Join("sketch.yaml").WriteFile(out)
	require.NoError(t, err)
	out, _, err = cli.Run("compile", "-m", "uno", "--json", sketch.String())
	require.NoError(t, err)
	// Check if local libraries are picked up correctly
	libInsideJson, _ := json.Marshal(libInside.String())
	libOutsideJson, _ := json.Marshal(libOutside.String())
	j := requirejson.Parse(t, out).Query(".builder_result.used_libraries")
	j.MustContain(`
		[
			{"name": "MyLib", "install_dir": ` + string(libInsideJson) + `},
			{"name": "MyLibOutside", "install_dir": ` + string(libOutsideJson) + `}
		]`)
}

func TestRelativeLocalLib(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Prepare the sketch with libraries
	tmpDir, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = tmpDir.RemoveAll })

	sketchTemplate, err := paths.New("testdata", "SketchWithLibrary").Abs()
	require.NoError(t, err)

	sketch := tmpDir.Join("SketchWithLibrary")
	libInside := sketch.Join("libraries", "MyLib")
	err = sketchTemplate.CopyDirTo(sketch)
	require.NoError(t, err)

	libOutsideTemplate := sketchTemplate.Join("..", "MyLibOutside")
	libOutside := sketch.Join("..", "MyLibOutside")
	err = libOutsideTemplate.CopyDirTo(libOutside)
	require.NoError(t, err)

	// Install the required core and libraries
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit BusIO@1.17.1", "--no-overwrite")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit GFX Library@1.12.1", "--no-overwrite")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "Adafruit SSD1306@2.5.14", "--no-overwrite")
	require.NoError(t, err)

	// Check if the profile dump:
	// - keeps libraries in the sketch with a relative path to the folder, not the .ino
	out, _, err := cli.Run("compile", "-b", "arduino:avr:uno",
		"--library", libInside.String(),
		"--library", libOutside.String(),
		"--dump-profile",
		sketch.Join("SketchWithLibrary.ino").String())
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(`
profiles:
  uno:
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.6)
    libraries:
      - dir: libraries/MyLib
      - dir: `+libOutside.String()+`
      - Adafruit SSD1306 (2.5.14)
      - Adafruit GFX Library (1.12.1)
      - Adafruit BusIO (1.17.1)
`), strings.TrimSpace(string(out)))

	outRelative := strings.Replace(string(out), libOutside.String(), "../MyLibOutside", 1)

	// Dump the profile in the sketch directory and compile with it again
	err = sketch.Join("sketch.yaml").WriteFile([]byte(outRelative))
	require.NoError(t, err)
	out, _, err = cli.Run("compile", "-m", "uno", "--json", sketch.Join("SketchWithLibrary.ino").String())
	require.NoError(t, err)
	// Check if local libraries are picked up correctly
	libInsideJson, _ := json.Marshal(libInside.String())
	libOutsideJson, _ := json.Marshal(libOutside.String())
	j := requirejson.Parse(t, out).Query(".builder_result.used_libraries")
	j.MustContain(`
		[
			{"name": "MyLib", "install_dir": ` + string(libInsideJson) + `},
			{"name": "MyLibOutside", "install_dir": ` + string(libOutsideJson) + `}
		]`)
}
