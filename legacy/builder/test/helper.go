// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
// Copyright 2015 Matthijs Kooijman
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

package test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func SetupBuildPath(t *testing.T) *paths.Path {
	buildPath, err := paths.MkTempDir("", "test_build_path")
	require.NoError(t, err)
	return buildPath
}

func parseFQBN(t *testing.T, fqbnIn string) *cores.FQBN {
	fqbn, err := cores.ParseFQBN(fqbnIn)
	require.NoError(t, err)
	return fqbn
}
