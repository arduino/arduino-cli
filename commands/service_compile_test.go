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

package commands

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/buildcache"
	"github.com/arduino/arduino-cli/internal/inventory"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

func TestMaybePurgeBuildCacheUsesConfiguredPath(t *testing.T) {
	cachePath := paths.New(t.TempDir()).Join("configured-cache")
	for _, entry := range []string{"cores/old-core", "sketches/old-sketch"} {
		_, err := buildcache.New(cachePath.Join(paths.New(entry).Parent().String())).GetOrCreate(paths.New(entry).Base())
		assert.NoError(t, err)
	}

	const countKey = "build_cache.compilation_count_since_last_purge"
	previousCount := inventory.Store.GetUint(countKey)
	inventory.Store.Set(countKey, uint(0))
	defer inventory.Store.Set(countKey, previousCount)

	maybePurgeBuildCache(1, 0, cachePath)

	assert.False(t, cachePath.Join("cores", "old-core").Exist())
	assert.False(t, cachePath.Join("sketches", "old-sketch").Exist())
}

func TestGenBuildPath(t *testing.T) {
	srv := NewArduinoCoreServer().(*arduinoCoreServerImpl)
	want := srv.settings.GetBuildCachePath().Join("sketches", "ACBD18DB4CC2F85CEDEF654FCCC4A4D8")
	act := srv.getDefaultSketchBuildPath(&sketch.Sketch{FullPath: paths.New("foo")}, nil)
	assert.True(t, act.EquivalentTo(want))
	assert.Equal(t, "ACBD18DB4CC2F85CEDEF654FCCC4A4D8", (&sketch.Sketch{FullPath: paths.New("foo")}).Hash())
}
