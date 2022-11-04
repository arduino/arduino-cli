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
	"testing"

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
}
