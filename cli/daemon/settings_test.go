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

package daemon

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestLoadWithoutConfigs(t *testing.T) {
	cmd := NewCommand()
	configFile := "unexisting-config.yml"

	s, err := load(cmd, configFile)
	require.NoError(t, err)
	require.NotNil(t, s)

	// Verify default settings
	require.Equal(t, "127.0.0.1", s.IP)
	require.Equal(t, "50051", s.Port)
	require.Equal(t, false, s.Daemonize)
	require.Equal(t, false, s.Debug)
	require.Equal(t, []string{}, s.DebugFilter)
	require.Equal(t, false, s.MetricsEnabled)
	require.Equal(t, ":9090", s.MetricsAddress)
}

func TestLoadWithConfigs(t *testing.T) {
	cmd := NewCommand()
	configFile := paths.New("testdata", "daemon-config.yaml")
	require.NoError(t, configFile.ToAbs())

	s, err := load(cmd, configFile.String())
	require.NoError(t, err)
	require.NotNil(t, s)

	// Verify settings ar correctly read from config file
	require.Equal(t, "127.0.0.1", s.IP)
	require.Equal(t, "0", s.Port)
	require.Equal(t, false, s.Daemonize)
	require.Equal(t, false, s.Debug)
	require.Equal(t, []string{}, s.DebugFilter)
	require.Equal(t, false, s.MetricsEnabled)
	require.Equal(t, ":9090", s.MetricsAddress)
}

func TestLoadWithConfigsAndFlags(t *testing.T) {
	cmd := NewCommand()
	configFile := paths.New("testdata", "daemon-config.yaml")
	require.NoError(t, configFile.ToAbs())

	cmd.Flags().Set("ip", "0.0.0.0")
	cmd.Flags().Set("port", "12345")

	s, err := load(cmd, configFile.String())
	require.NoError(t, err)
	require.NotNil(t, s)

	// Verify settings read from config file are override by those
	// set via flags
	require.Equal(t, "0.0.0.0", s.IP)
	require.Equal(t, "12345", s.Port)
	require.Equal(t, false, s.Daemonize)
	require.Equal(t, false, s.Debug)
	require.Equal(t, []string{}, s.DebugFilter)
	require.Equal(t, false, s.MetricsEnabled)
	require.Equal(t, ":9090", s.MetricsAddress)
}
