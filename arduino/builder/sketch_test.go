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

package builder_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/stretchr/testify/assert"
)

func TestSaveSketch(t *testing.T) {
	sketchName := t.Name() + ".ino"
	outName := sketchName + ".cpp"
	sketchFile := filepath.Join("testdata", sketchName)
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	source, err := ioutil.ReadFile(sketchFile)
	if err != nil {
		t.Fatalf("unable to read golden file %s: %v", sketchFile, err)
	}

	builder.SaveSketchItemCpp(&sketch.Item{Path: sketchName, Source: source}, tmp)

	out, err := ioutil.ReadFile(filepath.Join(tmp, outName))
	if err != nil {
		t.Fatalf("unable to read output file %s: %v", outName, err)
	}

	assert.Equal(t, source, out)
}

func TestLoadSketchFolder(t *testing.T) {
	// pass the path to the sketch folder
	sketchPath := filepath.Join("testdata", t.Name())
	mainFilePath := filepath.Join(sketchPath, t.Name()+".ino")
	s, err := builder.LoadSketch(sketchPath, "")
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, mainFilePath, s.MainFile.Path)
	assert.Len(t, s.OtherSketchFiles, 2) // [old.pde, other.ino]
	assert.Len(t, s.AdditionalFiles, 3)  // [header.h, s_file.S, src/helper.h]

	// pass the path to the main file
	sketchPath = mainFilePath
	s, err = builder.LoadSketch(sketchPath, "")
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, mainFilePath, s.MainFile.Path)
	assert.Len(t, s.OtherSketchFiles, 2)
	assert.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	assert.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	assert.Len(t, s.AdditionalFiles, 3)
	assert.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	assert.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	assert.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
}

func TestLoadSketchFolderWrongMain(t *testing.T) {
	sketchPath := filepath.Join("testdata", t.Name())
	_, err := builder.LoadSketch(sketchPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find the main sketch file")

	_, err = builder.LoadSketch("does/not/exist", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
