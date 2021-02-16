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

package sketch

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItem(t *testing.T) {
	sketchItem := filepath.Join("testdata", t.Name()+".ino")
	item := NewItem(sketchItem)
	assert.Equal(t, sketchItem, item.Path)
	sourceBytes, err := item.GetSourceBytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`#include <testlib.h>`), sourceBytes)
	sourceStr, err := item.GetSourceStr()
	assert.Nil(t, err)
	assert.Equal(t, "#include <testlib.h>", sourceStr)

	item = NewItem("doesnt/exist")
	sourceBytes, err = item.GetSourceBytes()
	assert.Nil(t, sourceBytes)
	assert.NotNil(t, err)
}

func TestSort(t *testing.T) {
	items := []*Item{
		{"foo"},
		{"baz"},
		{"bar"},
	}

	sort.Sort(ItemByPath(items))

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

	sketch, err := New(sketchFolderPath, mainFilePath, "", allFilesPaths)
	assert.Nil(t, err)
	assert.Equal(t, mainFilePath, sketch.MainFile.Path)
	assert.Equal(t, sketchFolderPath, sketch.LocationPath)
	assert.Len(t, sketch.OtherSketchFiles, 0)
	assert.Len(t, sketch.AdditionalFiles, 1)
	assert.Equal(t, sketch.AdditionalFiles[0].Path, paths.New(sketchFolderPath).Join("other.cpp").String())
	assert.Len(t, sketch.RootFolderFiles, 1)
	assert.Equal(t, sketch.RootFolderFiles[0].Path, paths.New(sketchFolderPath).Join("other.cpp").String())
}

func TestNewSketchCasingWrong(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchCasingWrong")
	mainFilePath := paths.New("testadata", "sketchcasingwrong.ino").String()
	sketch, err := New(sketchPath.String(), mainFilePath, "", []string{mainFilePath})
	assert.Nil(t, sketch)
	assert.Error(t, err)
	assert.IsType(t, &InvalidSketchFoldernameError{}, err)
	e := err.(*InvalidSketchFoldernameError)
	assert.NotNil(t, e.Sketch)
	expectedError := fmt.Sprintf("no valid sketch found in %s: missing %s", sketchPath.String(), sketchPath.Join(sketchPath.Base()+".ino"))
	assert.EqualError(t, err, expectedError)
}

func TestNewSketchCasingCorrect(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchCasingCorrect").String()
	mainFilePath := paths.New("testadata", "SketchCasingCorrect.ino").String()
	sketch, err := New(sketchPath, mainFilePath, "", []string{mainFilePath})
	assert.NotNil(t, sketch)
	assert.NoError(t, err)
	assert.Equal(t, sketchPath, sketch.LocationPath)
	assert.Equal(t, mainFilePath, sketch.MainFile.Path)
	assert.Len(t, sketch.OtherSketchFiles, 0)
	assert.Len(t, sketch.AdditionalFiles, 0)
	assert.Len(t, sketch.RootFolderFiles, 0)
}

func TestCheckSketchCasingWrong(t *testing.T) {
	sketchFolder := paths.New("testdata", "SketchCasingWrong")
	err := CheckSketchCasing(sketchFolder.String())
	expectedError := fmt.Sprintf("no valid sketch found in %s: missing %s", sketchFolder, sketchFolder.Join(sketchFolder.Base()+".ino"))
	assert.EqualError(t, err, expectedError)
}

func TestCheckSketchCasingCorrect(t *testing.T) {
	sketchFolder := paths.New("testdata", "SketchCasingCorrect").String()
	err := CheckSketchCasing(sketchFolder)
	require.NoError(t, err)
}
