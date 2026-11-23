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

package detector

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDetectorCacheKeepsSketchDirectives(t *testing.T) {
	cacheFile := paths.New(t.TempDir(), "includes.cache")

	saved := newDetectorCache()
	saved.ExpectAddedIncludePath(paths.New("/core"))
	saved.SetSketchDirectives([]string{"#include <Wire.h>"})
	require.NoError(t, saved.Save(cacheFile))

	loaded := newDetectorCache()
	require.NoError(t, loaded.Load(cacheFile))
	require.Equal(t, saved.entries[0].String(), loaded.Peek().String())
	require.True(t, loaded.SketchDirectivesUnchanged([]string{"#include <Wire.h>"}))
	require.False(t, loaded.SketchDirectivesUnchanged([]string{"#include <SPI.h>"}))
}

func TestDetectorCacheKeepsSketchDirectivesThatAreNotText(t *testing.T) {
	cacheFile := paths.New(t.TempDir(), "includes.cache")
	directives := directiveLines([]byte("#define NAME \"M\xfc\xfcller\"\n"))

	saved := newDetectorCache()
	saved.SetSketchDirectives(directives)
	require.NoError(t, saved.Save(cacheFile))

	loaded := newDetectorCache()
	require.NoError(t, loaded.Load(cacheFile))
	require.True(t, loaded.SketchDirectivesUnchanged(directives))
}

func TestDetectorCacheLoadsEntriesSavedWithoutDirectives(t *testing.T) {
	cacheFile := paths.New(t.TempDir(), "includes.cache")
	require.NoError(t, cacheFile.WriteFile([]byte(`[{"added_include_path": "/core"}]`)))

	loaded := newDetectorCache()
	require.NoError(t, loaded.Load(cacheFile))
	require.Equal(t, "Added include path: /core", loaded.Peek().String())
	require.False(t, loaded.SketchDirectivesUnchanged([]string{}))
}
