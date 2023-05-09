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
	"encoding/json"
	"os"
	"testing"
	"time"

	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestLibLayoutAndLocationJSONUnMarshaler(t *testing.T) {
	testLayout := func(l LibraryLayout) {
		d, err := json.Marshal(l)
		require.NoError(t, err)
		var m LibraryLayout
		err = json.Unmarshal(d, &m)
		require.NoError(t, err)
		require.Equal(t, l, m)
	}
	testLayout(FlatLayout)
	testLayout(RecursiveLayout)

	testLocation := func(l LibraryLocation) {
		d, err := json.Marshal(l)
		require.NoError(t, err)
		var m LibraryLocation
		err = json.Unmarshal(d, &m)
		require.NoError(t, err)
		require.Equal(t, l, m)
	}
	testLocation(IDEBuiltIn)
	testLocation(PlatformBuiltIn)
	testLocation(ReferencedPlatformBuiltIn)
	testLocation(User)
	testLocation(Unmanaged)
}

func TestLibrariesLoader(t *testing.T) {
	{
		lib, err := Load(paths.New("testdata", "TestLib"), User)
		require.NoError(t, err)
		require.Equal(t, "TestLib", lib.Name)
		require.Equal(t, "1.0.3", lib.Version.String())
		require.False(t, lib.IsLegacy)
		require.False(t, lib.InDevelopment)
	}
	{
		lib, err := Load(paths.New("testdata", "TestLibInDev"), User)
		require.NoError(t, err)
		require.Equal(t, "TestLibInDev", lib.Name)
		require.Equal(t, "1.0.3", lib.Version.String())
		require.False(t, lib.IsLegacy)
		require.True(t, lib.InDevelopment)
	}
	{
		lib, err := Load(paths.New("testdata", "LibWithNonUTF8Properties"), User)
		require.NoError(t, err)
		require.Equal(t, "LibWithNonUTF8Properties", lib.Name)
		require.Equal(t, "àrduìnò", lib.Author)
	}
	{
		lib, err := Load(paths.New("testdata", "EmptyLib"), User)
		require.Error(t, err)
		require.Nil(t, lib)
	}
	{
		lib, err := Load(paths.New("testdata", "LegacyLib"), User)
		require.NoError(t, err)
		require.Equal(t, "LegacyLib", lib.Name)
		require.True(t, lib.IsLegacy)
	}
}

func TestSymlinkLoop(t *testing.T) {
	// Set up directory structure of test library.
	testLib := paths.New("testdata", "TestLib")
	examplesPath := testLib.Join("examples")
	require.NoError(t, examplesPath.Mkdir())
	defer examplesPath.RemoveAll()

	// It's probably most friendly for contributors using Windows to create the symlinks needed for the test on demand.
	err := os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer1").String())
	require.NoError(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")
	// It's necessary to have multiple symlinks to a parent directory to create the loop.
	err = os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer2").String())
	require.NoError(t, err)

	// The failure condition is Load() never returning, testing for which requires setting up a timeout.
	done := make(chan bool)
	go func() {
		_, err = Load(testLib, User)
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		require.FailNow(t, "Load didn't complete in the allocated time.")
	}
	require.Error(t, err)
}

func TestLegacySymlinkLoop(t *testing.T) {
	// Set up directory structure of test library.
	testLib := paths.New("testdata", "LegacyLib")
	examplesPath := testLib.Join("examples")
	require.NoError(t, examplesPath.Mkdir())
	defer examplesPath.RemoveAll()

	// It's probably most friendly for contributors using Windows to create the symlinks needed for the test on demand.
	err := os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer1").String())
	require.NoError(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")
	// It's necessary to have multiple symlinks to a parent directory to create the loop.
	err = os.Symlink(examplesPath.Join("..").String(), examplesPath.Join("UpGoer2").String())
	require.NoError(t, err)

	// The failure condition is Load() never returning, testing for which requires setting up a timeout.
	done := make(chan bool)
	go func() {
		_, err = Load(testLib, User)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		require.FailNow(t, "Load didn't complete in the allocated time.")
	}
	require.Error(t, err)
}

func TestLoadExamples(t *testing.T) {
	example, err := paths.New(".", "testdata", "TestLibExamples", "examples", "simple").Abs()
	require.NoError(t, err)
	lib, err := Load(paths.New("testdata", "TestLibExamples"), User)
	require.NoError(t, err)
	require.Len(t, lib.Examples, 1)
	require.True(t, lib.Examples.Contains(example))
}
