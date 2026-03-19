// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestProjectFileLoading(t *testing.T) {
	{
		sketchProj := paths.New("testdata", "SketchWithProfiles", "sketch.yml")
		proj, err := LoadProjectFile(sketchProj)
		require.NoError(t, err)
		golden, err := sketchProj.ReadFile()
		require.NoError(t, err)
		require.Equal(t, proj.AsYaml(), string(golden))
	}
	{
		sketchProj := paths.New("testdata", "SketchWithDefaultFQBNAndPort", "sketch.yml")
		proj, err := LoadProjectFile(sketchProj)
		require.NoError(t, err)
		golden, err := sketchProj.ReadFile()
		require.NoError(t, err)
		require.Equal(t, proj.AsYaml(), string(golden))
	}
	{
		sketchProj := paths.New("testdata", "profiles", "profile_1.yml")
		proj, err := LoadProjectFile(sketchProj)
		require.NoError(t, err)
		golden, err := sketchProj.ReadFile()
		require.NoError(t, err)
		require.Equal(t, string(golden), proj.AsYaml())
	}
	{
		sketchProj := paths.New("testdata", "profiles", "bad_profile_1.yml")
		_, err := LoadProjectFile(sketchProj)
		require.Error(t, err)
	}
}

func TestProjectFileLibraries(t *testing.T) {
	sketchProj := paths.New("testdata", "profiles", "profile_with_libraries.yml")
	proj, err := LoadProjectFile(sketchProj)
	require.NoError(t, err)
	require.Len(t, proj.Profiles, 1)
	prof := proj.Profiles[0]
	require.Len(t, prof.Libraries, 6)
	require.Equal(t, "FlashStorage@1.2.3", prof.Libraries[0].String())
	require.Equal(t, "@dir:/path/to/system/lib", prof.Libraries[1].String())
	require.Equal(t, "@dir:path/to/sketch/lib", prof.Libraries[2].String())
	require.Equal(t, "DependencyLib@2.3.4 (dep)", prof.Libraries[3].String())
	require.Equal(t, "@git:https://github.com/username/HelloWorld.git#v2.13", prof.Libraries[4].String())
	require.Equal(t, "@git:https://github.com/username/HelloWorld.git#v2.14", prof.Libraries[5].String())
	require.Equal(t, "FlashStorage_1.2.3_e525d7c96b27788f", prof.Libraries[0].InternalUniqueIdentifier())
	require.Panics(t, func() { prof.Libraries[1].InternalUniqueIdentifier() })
	require.Panics(t, func() { prof.Libraries[2].InternalUniqueIdentifier() })
	require.Equal(t, "DependencyLib_2.3.4_ecde631facb47ae5", prof.Libraries[3].InternalUniqueIdentifier())
	require.Equal(t, "git-github.com_username_HelloWorld.git_v2.13-0c146203", prof.Libraries[4].InternalUniqueIdentifier())
	require.Equal(t, "git-github.com_username_HelloWorld.git_v2.14-49f5df7f", prof.Libraries[5].InternalUniqueIdentifier())

	orig, err := sketchProj.ReadFile()
	require.NoError(t, err)
	require.Equal(t, string(orig), proj.AsYaml())
}
