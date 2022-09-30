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

package compile_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestCompileWithLibrary(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	sketchName := "CompileSketchWithWiFi101Dependency"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino:avr:uno"
	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	data, err := sketchFile.ReadFile()
	require.NoError(t, err)
	data = append([]byte("#include <WiFi101.h>\n"), data...)
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	// Manually installs a library
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	libPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")
	_, err = git.PlainClone(libPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.16.1"),
	})
	require.NoError(t, err)

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--library", libPath.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "WiFi101")
}
