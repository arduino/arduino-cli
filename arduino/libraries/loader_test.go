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

package libraries

import (
	"os"
	"testing"
	"time"

	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkLoop(t *testing.T) {
	// Set up directory structure of test library.
	libraryPath, err := paths.TempDir().MkTempDir("TestSymlinkLoop")
	defer libraryPath.RemoveAll() // Clean up after the test.
	require.Nil(t, err)
	err = libraryPath.Join("TestSymlinkLoop.h").WriteFile([]byte{})
	require.Nil(t, err)
	examplesPath := libraryPath.Join("examples")
	err = examplesPath.Mkdir()
	require.Nil(t, err)

	// It's probably most friendly for contributors using Windows to create the symlinks needed for the test on demand.
	err = os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer1").String())
	require.Nil(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")
	// It's necessary to have multiple symlinks to a parent directory to create the loop.
	err = os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer2").String())
	require.Nil(t, err)

	// The failure condition is Load() never returning, testing for which requires setting up a timeout.
	done := make(chan bool)
	go func() {
		_, err = Load(libraryPath, User)
		done <- true
	}()

	assert.Eventually(
		t,
		func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		},
		20*time.Second,
		10*time.Millisecond,
		"Infinite symlink loop while loading library",
	)
	require.Nil(t, err)
}
