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

package librariesmanager

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestParseGitURL(t *testing.T) {
	{
		_, _, _, err := parseGitArgURL("")
		require.EqualError(t, err, "invalid git url")
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("https://github.com/arduino/arduino-lib.git")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "https://github.com/arduino/arduino-lib.git", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("https://github.com/arduino/arduino-lib.git#0.1.2")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "https://github.com/arduino/arduino-lib.git", gitURL)
		require.EqualValues(t, "0.1.2", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("git@github.com:arduino/arduino-lib.git")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "https://github.com/arduino/arduino-lib.git", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("git@bitbucket.org:arduino/arduino-lib.git")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "https://bitbucket.org/arduino/arduino-lib.git", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("git@github.com:arduino/arduino-lib.git#0.1.2")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "https://github.com/arduino/arduino-lib.git", gitURL)
		require.EqualValues(t, "0.1.2", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("file:///path/to/arduino-lib")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "file:///path/to/arduino-lib", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("file:///path/to/arduino-lib.git")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "file:///path/to/arduino-lib.git", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("/path/to/arduino-lib")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "/path/to/arduino-lib", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		libraryName, gitURL, ref, err := parseGitArgURL("/path/to/arduino-lib.git")
		require.NoError(t, err)
		require.Equal(t, "arduino-lib", libraryName)
		require.Equal(t, "/path/to/arduino-lib.git", gitURL)
		require.EqualValues(t, "", ref)
	}
	{
		_, _, _, err := parseGitArgURL("https://arduino.cc")
		require.EqualError(t, err, "invalid git url")
	}
	{
		_, _, _, err := parseGitArgURL("https://arduino.cc/")
		require.EqualError(t, err, "invalid git url")
	}
	{
		_, _, _, err := parseGitArgURL("://not@a@url")
		require.EqualError(t, err, "invalid git url: parse \"://not@a@url\": missing protocol scheme")
	}
}

func TestValidateLibrary(t *testing.T) {
	tmpDir := paths.New(t.TempDir())

	nonExistingDirLib := tmpDir.Join("nonExistingDirLib")
	err := validateLibrary(nonExistingDirLib)
	require.Errorf(t, err, "directory doesn't exist: %s", nonExistingDirLib)

	emptyLib := tmpDir.Join("emptyLib")
	emptyLib.Mkdir()
	err = validateLibrary(emptyLib)
	require.Errorf(t, err, "library not valid")

	onlyPropertiesLib := tmpDir.Join("onlyPropertiesLib")
	onlyPropertiesLib.Mkdir()
	onlyPropertiesLib.Join("library.properties").WriteFile([]byte{})
	err = validateLibrary(onlyPropertiesLib)
	require.Errorf(t, err, "library not valid")

	missingPropertiesLib := tmpDir.Join("missingPropertiesLib")
	missingPropertiesLibSourceDir := missingPropertiesLib.Join("src")
	missingPropertiesLibSourceDir.MkdirAll()
	missingPropertiesLibSourceDir.Join("some_file.hpp").WriteFile([]byte{})
	err = validateLibrary(missingPropertiesLib)
	require.Errorf(t, err, "library not valid")

	validLib := tmpDir.Join("valiLib")
	validLibSourceDir := validLib.Join("src")
	validLibSourceDir.MkdirAll()
	validLibSourceDir.Join("some_file.hpp").WriteFile([]byte{})
	validLib.Join("library.properties").WriteFile([]byte{})
	err = validateLibrary(validLib)
	require.NoError(t, err)

	validLegacyLib := tmpDir.Join("validLegacyLib")
	validLegacyLib.Mkdir()
	validLegacyLib.Join("some_file.hpp").WriteFile([]byte{})
	err = validateLibrary(validLib)
	require.NoError(t, err)
}
