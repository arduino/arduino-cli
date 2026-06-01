// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestYamlUpdate(t *testing.T) {
	{
		sample, err := paths.New("testdata", "SketchWithProfiles", "sketch.yml").ReadFile()
		require.NoError(t, err)
		tmp, err := paths.WriteToTempFile(sample, nil, "")
		require.NoError(t, err)
		defer tmp.Remove()

		err = updateOrAddYamlRootEntry(tmp, "default_fqbn", "arduino:avr:uno")
		require.NoError(t, err)
		err = updateOrAddYamlRootEntry(tmp, "default_port", "/dev/ttyACM0")
		require.NoError(t, err)

		updated, err := tmp.ReadFile()
		require.NoError(t, err)
		expected := string(sample)
		expected += fmt.Sprintln("default_fqbn: arduino:avr:uno")
		expected += fmt.Sprintln("default_port: /dev/ttyACM0")
		require.Equal(t, expected, string(updated))
	}
	{
		sample, err := paths.New("testdata", "SketchWithDefaultFQBNAndPort", "sketch.yml").ReadFile()
		require.NoError(t, err)
		tmp, err := paths.WriteToTempFile(sample, nil, "")
		require.NoError(t, err)
		defer tmp.Remove()

		err = updateOrAddYamlRootEntry(tmp, "default_fqbn", "TEST1")
		require.NoError(t, err)
		err = updateOrAddYamlRootEntry(tmp, "default_port", "TEST2")
		require.NoError(t, err)

		updated, err := tmp.ReadFile()
		fmt.Print(string(updated))
		require.NoError(t, err)
		expected := strings.Replace(string(sample), "arduino:avr:uno", "TEST1", 1)
		expected = strings.Replace(expected, "/dev/ttyACM0", "TEST2", 1)
		require.Equal(t, expected, string(updated))
	}
	{
		tmp, err := paths.WriteToTempFile([]byte{}, nil, "")
		require.NoError(t, err)
		require.NoError(t, tmp.Remove())
		err = updateOrAddYamlRootEntry(tmp, "default_fqbn", "TEST1")
		require.NoError(t, err)

		updated, err := tmp.ReadFile()
		require.NoError(t, err)
		expected := "default_fqbn: TEST1\n"
		require.Equal(t, expected, string(updated))
	}
}
