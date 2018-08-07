/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package packagemanager_test

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/stretchr/testify/require"
)

func TestFindBoardWithFQBN(t *testing.T) {
	pm := packagemanager.NewPackageManager(
		paths.New("testdata"),
		paths.New("testdata"),
		paths.New("testdata"),
		paths.New("testdata"))
	pm.LoadHardwareFromDirectory(paths.New("testdata"))

	board, err := pm.FindBoardWithFQBN("arduino:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Uno")

	board, err = pm.FindBoardWithFQBN("arduino:avr:mega")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Mega or Mega 2560")
}
