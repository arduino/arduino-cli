// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package types_test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

func TestNewSourceFileFromSketch(t *testing.T) {
	c := getContext()
	c.SketchBuildPath = paths.New("/absolute")
	origin := &types.Sketch{
		MainFile: types.SketchFile{Name: paths.New("/absolute/path/to/foo")},
	}

	path := paths.New("relative")
	sf, err := types.NewSourceFile(c, origin, path)
	assert.Nil(t, err)
	assert.Equal(t, "relative", sf.RelativePath.String())

	path = paths.New("/absolute")
	c.SketchBuildPath = paths.New("relative")
	sf, err = types.NewSourceFile(c, origin, path)
	assert.NotNil(t, err)
}

func TestNewSourceFileFromLibrary(t *testing.T) {
	c := getContext()
	origin := &libraries.Library{
		SourceDir: paths.New("/absolute/path/to/foo"),
	}

	path := paths.New("relative")
	sf, err := types.NewSourceFile(c, origin, path)
	assert.Nil(t, err)
	assert.Equal(t, "relative", sf.RelativePath.String())

	path = paths.New("/absolute")
	origin.SourceDir = paths.New("relative")
	sf, err = types.NewSourceFile(c, origin, path)
	assert.NotNil(t, err)
}

func TestGettersFromSketch(t *testing.T) {
	c := getContext()
	c.SketchBuildPath = paths.New("/absolute")
	origin := &types.Sketch{
		MainFile: types.SketchFile{Name: paths.New("/absolute/path/to/foo")},
	}
	path := paths.New("relative")
	sf, _ := types.NewSourceFile(c, origin, path)
	assert.Equal(t, "/absolute/relative", sf.SourcePath(c).String())
	assert.Equal(t, "/absolute/relative.o", sf.ObjectPath(c).String())
	assert.Equal(t, "/absolute/relative.d", sf.DepfilePath(c).String())
}

func TestGettersFromLibrary(t *testing.T) {
	c := getContext()
	c.LibrariesBuildPath = paths.New("/absolute/path/to/foo")
	origin := &libraries.Library{
		SourceDir: paths.New("/absolute/path/to/foo"),
	}
	path := paths.New("relative")
	sf, _ := types.NewSourceFile(c, origin, path)
	assert.Equal(t, "/absolute/path/to/foo/relative", sf.SourcePath(c).String())
	assert.Equal(t, "/absolute/path/to/foo/relative.o", sf.ObjectPath(c).String())
	assert.Equal(t, "/absolute/path/to/foo/relative.d", sf.DepfilePath(c).String())
}

func TestUniqueSourceFileQueue(t *testing.T) {
	c := getContext()
	origin := &libraries.Library{
		SourceDir: paths.New("/absolute/path/to/foo"),
	}

	path1 := paths.New("path1")
	sf1, _ := types.NewSourceFile(c, origin, path1)

	path2 := paths.New("path2")
	sf2, _ := types.NewSourceFile(c, origin, path2)

	q := types.UniqueSourceFileQueue{sf1, sf2}

	assert.Len(t, q, 2)
	assert.False(t, q.Less(0, 1)) // Less always returns false
	assert.False(t, q.Less(1, 0))
	assert.Panics(t, func() { q.Swap(0, 1) })

	// push an item that's already there and ensure it's discarded
	q.Push(sf1)
	assert.Len(t, q, 2)

	// push a new item
	path3 := paths.New("path3")
	sf3, _ := types.NewSourceFile(c, origin, path3)
	q.Push(sf3)
	assert.Len(t, q, 3)

	// Pop the items (from the head)
	assert.Equal(t, sf1, q.Pop())
	assert.Equal(t, sf2, q.Pop())
	assert.Equal(t, sf3, q.Pop())

	// assert the queue is empty
	assert.Len(t, q, 0)
	assert.True(t, q.Empty())
}
