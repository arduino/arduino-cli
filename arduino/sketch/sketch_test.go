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

package sketch_test

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/stretchr/testify/assert"
)

func TestNewItem(t *testing.T) {
	sketchItem := filepath.Join("testdata", t.Name()+".ino")
	item, err := sketch.NewItem(sketchItem)
	assert.Nil(t, err)
	assert.Equal(t, sketchItem, item.Path)
	assert.Equal(t, []byte(`#include <testlib.h>`), item.Source)
	assert.Equal(t, "#include <testlib.h>", item.GetSourceStr())

	item, err = sketch.NewItem("doesnt/exist")
	assert.Nil(t, item)
	assert.NotNil(t, err)
}

func TestSort(t *testing.T) {
	items := []*sketch.Item{
		&sketch.Item{"foo", nil},
		&sketch.Item{"baz", nil},
		&sketch.Item{"bar", nil},
	}

	sort.Sort(sketch.ItemByPath(items))

	assert.Equal(t, "bar", items[0].Path)
	assert.Equal(t, "baz", items[1].Path)
	assert.Equal(t, "foo", items[2].Path)
}

func TestNew(t *testing.T) {
	sketchFolderPath := filepath.Join("testdata", t.Name())
	mainFilePath := filepath.Join(sketchFolderPath, t.Name()+".ino")
	otherFile := filepath.Join(sketchFolderPath, "other.cpp")
	allFilesPaths := []string{
		mainFilePath,
		otherFile,
	}

	sketch, err := sketch.New(sketchFolderPath, mainFilePath, "", allFilesPaths)
	assert.Nil(t, err)
	assert.Equal(t, mainFilePath, sketch.MainFile.Path)
	assert.Equal(t, sketchFolderPath, sketch.LocationPath)
	assert.Len(t, sketch.OtherSketchFiles, 0)
	assert.Len(t, sketch.AdditionalFiles, 1)
}
