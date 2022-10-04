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

package board_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCorrectBoardListOrdering(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("board", "listall", "--format", "json")
	require.NoError(t, err)
	requirejson.Query(t, jsonOut, "[.boards[] | .fqbn]", `[
		"arduino:avr:yun",
		"arduino:avr:uno",
		"arduino:avr:unomini",
		"arduino:avr:diecimila",
		"arduino:avr:nano",
		"arduino:avr:mega",
		"arduino:avr:megaADK",
		"arduino:avr:leonardo",
		"arduino:avr:leonardoeth",
		"arduino:avr:micro",
		"arduino:avr:esplora",
		"arduino:avr:mini",
		"arduino:avr:ethernet",
		"arduino:avr:fio",
		"arduino:avr:bt",
		"arduino:avr:LilyPadUSB",
		"arduino:avr:lilypad",
		"arduino:avr:pro",
		"arduino:avr:atmegang",
		"arduino:avr:robotControl",
		"arduino:avr:robotMotor",
		"arduino:avr:gemma",
		"arduino:avr:circuitplay32u4cat",
		"arduino:avr:yunmini",
		"arduino:avr:chiwawa",
		"arduino:avr:one",
		"arduino:avr:unowifi"
	]`)
}
