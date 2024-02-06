// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package diagnostics

import (
	"github.com/sirupsen/logrus"
)

type Store struct {
	results Diagnostics
}

func NewStore() *Store {
	return &Store{}
}

func (m *Store) Parse(cmdline []string, out []byte) {
	compiler := DetectCompilerFromCommandLine(
		cmdline,
		false, // at the moment compiler-probing is not required
	)
	if compiler == nil {
		logrus.Warnf("Could not detect compiler from: %s", cmdline)
		return
	}
	diags, err := ParseCompilerOutput(compiler, out)
	if err != nil {
		logrus.Warnf("Error parsing compiler output: %s", err)
		return
	}
	m.results = append(m.results, diags...)
}

func (m *Store) Diagnostics() Diagnostics {
	return m.results
}
