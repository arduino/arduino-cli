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

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// Environment is a test environment for the test suite.
type Environment struct {
	rootDir      *paths.Path
	downloadsDir *paths.Path
	t            *require.Assertions
}

// SharedDownloadDir returns the shared downloads directory.
func SharedDownloadDir(t *testing.T) *paths.Path {
	downloadsDir := paths.TempDir().Join("arduino-cli-test-suite-staging")
	require.NoError(t, downloadsDir.MkdirAll())
	return downloadsDir
}

// NewEnvironment creates a new test environment.
func NewEnvironment(t *testing.T) *Environment {
	downloadsDir := SharedDownloadDir(t)
	rootDir, err := paths.MkTempDir("", "arduino-cli-test-suite")
	require.NoError(t, err)
	return &Environment{
		rootDir:      rootDir,
		downloadsDir: downloadsDir,
		t:            require.New(t),
	}
}

// CleanUp removes the test environment.
func (e *Environment) CleanUp() {
	e.t.NoError(e.rootDir.RemoveAll())
}

// Root returns the root dir of the environment.
func (e *Environment) Root() *paths.Path {
	return e.rootDir
}
