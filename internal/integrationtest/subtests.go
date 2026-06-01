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

package integrationtest

import "testing"

// CLISubtests is a suite of tests to run under the same integreationtest.Environment
type CLISubtests []struct {
	Name     string
	Function func(*testing.T, *Environment, *ArduinoCLI)
}

// Run runs the test suite as subtests of the current test
func (testSuite CLISubtests) Run(t *testing.T, env *Environment, cli *ArduinoCLI) {
	for _, test := range testSuite {
		t.Run(test.Name, func(t *testing.T) {
			test.Function(t, env, cli)
		})
	}
}
