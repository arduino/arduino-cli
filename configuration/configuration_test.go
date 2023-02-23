// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func tmpDirOrDie() string {
	dir, err := os.MkdirTemp(os.TempDir(), "cli_test")
	if err != nil {
		panic(fmt.Sprintf("error creating tmp dir: %v", err))
	}
	// Symlinks are evaluated becase the temp folder on Mac OS is inside /var, it's not writable
	// and is a symlink to /private/var, we want the full path so we do this
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		panic(fmt.Sprintf("error evaluating tmp dir symlink: %v", err))
	}
	return dir
}

func TestInit(t *testing.T) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	settings := Init(filepath.Join(tmp, "arduino-cli.yaml"))
	require.NotNil(t, settings)

	require.Equal(t, "info", settings.GetString("logging.level"))
	require.Equal(t, "text", settings.GetString("logging.format"))

	require.Empty(t, settings.GetStringSlice("board_manager.additional_urls"))

	require.NotEmpty(t, settings.GetString("directories.Data"))
	require.NotEmpty(t, settings.GetString("directories.Downloads"))
	require.NotEmpty(t, settings.GetString("directories.User"))

	require.Equal(t, "50051", settings.GetString("daemon.port"))

	require.Equal(t, true, settings.GetBool("metrics.enabled"))
	require.Equal(t, ":9090", settings.GetString("metrics.addr"))
}

func TestFindConfigFile(t *testing.T) {
	configFile := FindConfigFileInArgs([]string{"--config-file"})
	require.Equal(t, "", configFile)

	configFile = FindConfigFileInArgs([]string{"--config-file", "some/path/to/config"})
	require.Equal(t, "some/path/to/config", configFile)

	configFile = FindConfigFileInArgs([]string{"--config-file", "some/path/to/config/arduino-cli.yaml"})
	require.Equal(t, "some/path/to/config/arduino-cli.yaml", configFile)

	configFile = FindConfigFileInArgs([]string{})
	require.Equal(t, "", configFile)
}
