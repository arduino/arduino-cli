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

package builder_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/stretchr/testify/require"
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

	builder.SketchSaveItemCpp(sketchName, source, tmp)

	out, err := ioutil.ReadFile(filepath.Join(tmp, outName))
	if err != nil {
		t.Fatalf("unable to read output file %s: %v", outName, err)
	}

	require.Equal(t, source, out)
}

func TestLoadSketchFolder(t *testing.T) {
	// pass the path to the sketch folder
	sketchPath := filepath.Join("testdata", t.Name())
	mainFilePath := filepath.Join(sketchPath, t.Name()+".ino")
	s, err := builder.SketchLoad(sketchPath, "")
	require.Nil(t, err)
	require.NotNil(t, s)
	require.Equal(t, mainFilePath, s.MainFile.Path)
	require.Equal(t, sketchPath, s.LocationPath)
	require.Len(t, s.OtherSketchFiles, 2)
	require.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	require.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	require.Len(t, s.AdditionalFiles, 3)
	require.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	require.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
	require.Len(t, s.RootFolderFiles, 4)
	require.Equal(t, "header.h", filepath.Base(s.RootFolderFiles[0].Path))
	require.Equal(t, "old.pde", filepath.Base(s.RootFolderFiles[1].Path))
	require.Equal(t, "other.ino", filepath.Base(s.RootFolderFiles[2].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.RootFolderFiles[3].Path))

	// pass the path to the main file
	sketchPath = mainFilePath
	s, err = builder.SketchLoad(sketchPath, "")
	require.Nil(t, err)
	require.NotNil(t, s)
	require.Equal(t, mainFilePath, s.MainFile.Path)
	require.Len(t, s.OtherSketchFiles, 2)
	require.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	require.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	require.Len(t, s.AdditionalFiles, 3)
	require.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	require.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
	require.Len(t, s.RootFolderFiles, 4)
	require.Equal(t, "header.h", filepath.Base(s.RootFolderFiles[0].Path))
	require.Equal(t, "old.pde", filepath.Base(s.RootFolderFiles[1].Path))
	require.Equal(t, "other.ino", filepath.Base(s.RootFolderFiles[2].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.RootFolderFiles[3].Path))
}

func TestLoadSketchFolderPde(t *testing.T) {
	// pass the path to the sketch folder
	sketchPath := filepath.Join("testdata", t.Name())
	mainFilePath := filepath.Join(sketchPath, t.Name()+".pde")
	s, err := builder.SketchLoad(sketchPath, "")
	require.Nil(t, err)
	require.NotNil(t, s)
	require.Equal(t, mainFilePath, s.MainFile.Path)
	require.Equal(t, sketchPath, s.LocationPath)
	require.Len(t, s.OtherSketchFiles, 2)
	require.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	require.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	require.Len(t, s.AdditionalFiles, 3)
	require.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	require.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
	require.Len(t, s.RootFolderFiles, 4)
	require.Equal(t, "header.h", filepath.Base(s.RootFolderFiles[0].Path))
	require.Equal(t, "old.pde", filepath.Base(s.RootFolderFiles[1].Path))
	require.Equal(t, "other.ino", filepath.Base(s.RootFolderFiles[2].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.RootFolderFiles[3].Path))
}

func TestLoadSketchFolderBothInoAndPde(t *testing.T) {
	// pass the path to the sketch folder containing two main sketches, .ino and .pde
	sketchPath := filepath.Join("testdata", t.Name())
	_, err := builder.SketchLoad(sketchPath, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple main sketch files found")
	require.Contains(t, err.Error(), t.Name()+".ino")
	require.Contains(t, err.Error(), t.Name()+".pde")
}

func TestLoadSketchFolderSymlink(t *testing.T) {
	// pass the path to the sketch folder
	symlinkSketchPath := filepath.Join("testdata", t.Name())
	srcSketchPath := t.Name() + "Src"
	os.Symlink(srcSketchPath, symlinkSketchPath)
	defer os.Remove(symlinkSketchPath)
	mainFilePath := filepath.Join(symlinkSketchPath, t.Name()+".ino")
	s, err := builder.SketchLoad(symlinkSketchPath, "")
	require.Nil(t, err)
	require.NotNil(t, s)
	require.Equal(t, mainFilePath, s.MainFile.Path)
	require.Equal(t, symlinkSketchPath, s.LocationPath)
	require.Len(t, s.OtherSketchFiles, 2)
	require.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	require.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	require.Len(t, s.AdditionalFiles, 3)
	require.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	require.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
	require.Len(t, s.RootFolderFiles, 4)
	require.Equal(t, "header.h", filepath.Base(s.RootFolderFiles[0].Path))
	require.Equal(t, "old.pde", filepath.Base(s.RootFolderFiles[1].Path))
	require.Equal(t, "other.ino", filepath.Base(s.RootFolderFiles[2].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.RootFolderFiles[3].Path))

	// pass the path to the main file
	symlinkSketchPath = mainFilePath
	s, err = builder.SketchLoad(symlinkSketchPath, "")
	require.Nil(t, err)
	require.NotNil(t, s)
	require.Equal(t, mainFilePath, s.MainFile.Path)
	require.Len(t, s.OtherSketchFiles, 2)
	require.Equal(t, "old.pde", filepath.Base(s.OtherSketchFiles[0].Path))
	require.Equal(t, "other.ino", filepath.Base(s.OtherSketchFiles[1].Path))
	require.Len(t, s.AdditionalFiles, 3)
	require.Equal(t, "header.h", filepath.Base(s.AdditionalFiles[0].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.AdditionalFiles[1].Path))
	require.Equal(t, "helper.h", filepath.Base(s.AdditionalFiles[2].Path))
	require.Len(t, s.RootFolderFiles, 4)
	require.Equal(t, "header.h", filepath.Base(s.RootFolderFiles[0].Path))
	require.Equal(t, "old.pde", filepath.Base(s.RootFolderFiles[1].Path))
	require.Equal(t, "other.ino", filepath.Base(s.RootFolderFiles[2].Path))
	require.Equal(t, "s_file.S", filepath.Base(s.RootFolderFiles[3].Path))
}

func TestLoadSketchFolderIno(t *testing.T) {
	// pass the path to the sketch folder
	sketchPath := filepath.Join("testdata", t.Name())
	_, err := builder.SketchLoad(sketchPath, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "sketch must not be a directory")
}

func TestLoadSketchFolderWrongMain(t *testing.T) {
	sketchPath := filepath.Join("testdata", t.Name())
	_, err := builder.SketchLoad(sketchPath, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to find a sketch file in directory testdata")

	_, err = builder.SketchLoad("does/not/exist", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "does/not/exist")
}

func TestMergeSketchSources(t *testing.T) {
	// borrow the sketch from TestLoadSketchFolder to avoid boilerplate
	s, err := builder.SketchLoad(filepath.Join("testdata", "TestLoadSketchFolder"), "")
	require.Nil(t, err)
	require.NotNil(t, s)

	// load expected result
	suffix := ".txt"
	if runtime.GOOS == "windows" {
		suffix = "_win.txt"
	}
	mergedPath := filepath.Join("testdata", t.Name()+suffix)
	mergedBytes, err := ioutil.ReadFile(mergedPath)
	if err != nil {
		t.Fatalf("unable to read golden file %s: %v", mergedPath, err)
	}

	offset, source, err := builder.SketchMergeSources(s, nil)
	require.Nil(t, err)
	require.Equal(t, 2, offset)
	require.Equal(t, string(mergedBytes), source)
}

func TestMergeSketchSourcesArduinoIncluded(t *testing.T) {
	s, err := builder.SketchLoad(filepath.Join("testdata", t.Name()), "")
	require.Nil(t, err)
	require.NotNil(t, s)

	// ensure not to include Arduino.h when it's already there
	_, source, err := builder.SketchMergeSources(s, nil)
	require.Nil(t, err)
	require.Equal(t, 1, strings.Count(source, "<Arduino.h>"))
}

func TestCopyAdditionalFiles(t *testing.T) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)

	// load the golden sketch
	s1, err := builder.SketchLoad(filepath.Join("testdata", t.Name()), "")
	require.Nil(t, err)
	require.Len(t, s1.AdditionalFiles, 1)

	// copy the sketch over, create a fake main file we don't care about it
	// but we need it for `SketchLoad` to succeed later
	err = builder.SketchCopyAdditionalFiles(s1, tmp, nil)
	require.Nil(t, err)
	fakeIno := filepath.Join(tmp, fmt.Sprintf("%s.ino", filepath.Base(tmp)))
	require.Nil(t, ioutil.WriteFile(fakeIno, []byte{}, os.FileMode(0644)))

	// compare
	s2, err := builder.SketchLoad(tmp, "")
	require.Nil(t, err)
	require.Len(t, s2.AdditionalFiles, 1)

	// save file info
	info1, err := os.Stat(s2.AdditionalFiles[0].Path)
	require.Nil(t, err)

	// copy again
	err = builder.SketchCopyAdditionalFiles(s1, tmp, nil)
	require.Nil(t, err)

	// verify file hasn't changed
	info2, err := os.Stat(s2.AdditionalFiles[0].Path)
	require.Equal(t, info1.ModTime(), info2.ModTime())
}

func TestLoadSketchCaseMismatch(t *testing.T) {
	// pass the path to the sketch folder
	sketchPath := filepath.Join("testdata", t.Name())
	mainFilePath := filepath.Join(sketchPath, t.Name()+".ino")
	s, err := builder.SketchLoad(sketchPath, "")
	require.Nil(t, s)
	require.Error(t, err)

	// pass the path to the main file
	s, err = builder.SketchLoad(mainFilePath, "")
	require.Nil(t, s)
	require.Error(t, err)
}
