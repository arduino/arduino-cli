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

// ProjectName is the prefix used in the test temp files
var ProjectName = "cli"

// Environment is a test environment for the test suite.
type Environment struct {
	rootDir      *paths.Path
	downloadsDir *SharedDir
	t            *testing.T
	cleanUp      func()
}

// NewEnvironment creates a new test environment.
func NewEnvironment(t *testing.T) *Environment {
	downloadsDir := NewSharedDir(t, "downloads")
	rootDir, err := paths.MkTempDir("", ProjectName)
	require.NoError(t, err)
	return &Environment{
		rootDir:      rootDir,
		downloadsDir: downloadsDir,
		t:            t,
		cleanUp: func() {
			require.NoError(t, rootDir.RemoveAll())
		},
	}
}

// RegisterCleanUpCallback adds a clean up function to the clean up chain
func (e *Environment) RegisterCleanUpCallback(newCleanUp func()) {
	previousCleanUp := e.cleanUp
	e.cleanUp = func() {
		newCleanUp()
		previousCleanUp()
	}
}

// CleanUp removes the test environment.
func (e *Environment) CleanUp() {
	e.cleanUp()
}

// RootDir returns the root dir of the environment.
func (e *Environment) RootDir() *paths.Path {
	return e.rootDir
}

// SharedDownloadsDir return the shared directory for downloads
func (e *Environment) SharedDownloadsDir() *SharedDir {
	return e.downloadsDir
}

// T returns the testing environment
func (e *Environment) T() *testing.T {
	return e.t
}
