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
	"os"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	sketchFolderPath := paths.New("testdata", "SketchSimple")
	mainFilePath := sketchFolderPath.Join(fmt.Sprintf("%s.ino", "SketchSimple"))
	otherFile := sketchFolderPath.Join("other.cpp")

	sketch, err := New(nil)
	assert.Nil(t, sketch)
	assert.Error(t, err)

	// Loading using Sketch folder path
	sketch, err = New(sketchFolderPath)
	assert.Nil(t, err)
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.True(t, sketchFolderPath.EquivalentTo(sketch.FullPath))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	assert.Equal(t, sketch.AdditionalFiles.Len(), 1)
	assert.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(otherFile))
	assert.Equal(t, sketch.RootFolderFiles.Len(), 1)
	assert.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(otherFile))

	// Loading using Sketch main file path
	sketch, err = New(mainFilePath)
	assert.Nil(t, err)
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.True(t, sketchFolderPath.EquivalentTo(sketch.FullPath))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	assert.Equal(t, sketch.AdditionalFiles.Len(), 1)
	assert.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(otherFile))
	assert.Equal(t, sketch.RootFolderFiles.Len(), 1)
	assert.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(otherFile))
}

func TestNewSketchPde(t *testing.T) {
	sketchFolderPath := paths.New("testdata", "SketchPde")
	mainFilePath := sketchFolderPath.Join(fmt.Sprintf("%s.pde", "SketchPde"))

	// Loading using Sketch folder path
	sketch, err := New(sketchFolderPath)
	assert.Nil(t, err)
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.True(t, sketchFolderPath.EquivalentTo(sketch.FullPath))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	assert.Equal(t, sketch.AdditionalFiles.Len(), 0)
	assert.Equal(t, sketch.RootFolderFiles.Len(), 0)

	// Loading using Sketch main file path
	sketch, err = New(mainFilePath)
	assert.Nil(t, err)
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.True(t, sketchFolderPath.EquivalentTo(sketch.FullPath))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	assert.Equal(t, sketch.AdditionalFiles.Len(), 0)
	assert.Equal(t, sketch.RootFolderFiles.Len(), 0)
}

func TestNewSketchBothInoAndPde(t *testing.T) {
	sketchName := "SketchBothInoAndPde"
	sketchFolderPath := paths.New("testdata", sketchName)
	sketch, err := New(sketchFolderPath)
	require.Nil(t, sketch)
	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple main sketch files found")
	require.Contains(t, err.Error(), fmt.Sprintf("%s.ino", sketchName))
	require.Contains(t, err.Error(), fmt.Sprintf("%s.pde", sketchName))
}

func TestNewSketchWrongMain(t *testing.T) {
	sketchName := "SketchWithWrongMain"
	sketchFolderPath := paths.New("testdata", sketchName)
	sketch, err := New(sketchFolderPath)
	require.Nil(t, sketch)
	require.Error(t, err)
	sketchFolderPath, _ = sketchFolderPath.Abs()
	expectedMainFile := sketchFolderPath.Join(sketchName)
	expectedError := fmt.Sprintf("main file missing from sketch: %s", expectedMainFile)
	require.Contains(t, err.Error(), expectedError)

	sketchFolderPath = paths.New("testdata", sketchName)
	mainFilePath := sketchFolderPath.Join(fmt.Sprintf("%s.ino", sketchName))
	sketch, err = New(mainFilePath)
	require.Nil(t, sketch)
	require.Error(t, err)
	expectedError = fmt.Sprintf("no such file or directory: %s", expectedMainFile)
	require.Contains(t, err.Error(), expectedError)
}

func TestNewSketchCasingWrong(t *testing.T) {
	{
		sketchPath := paths.New("testdata", "SketchWithWrongMain")
		sketch, err := New(sketchPath)
		assert.Nil(t, sketch)
		assert.Error(t, err)
		_, ok := err.(*InvalidSketchFolderNameError)
		assert.False(t, ok)
		sketchPath, _ = sketchPath.Abs()
		expectedError := fmt.Sprintf("main file missing from sketch: %s", sketchPath.Join(sketchPath.Base()+".ino"))
		assert.EqualError(t, err, expectedError)
	}
	{
		sketchPath := paths.New("testdata", "SketchWithWrongMain", "main.ino")
		sketch, err := New(sketchPath)
		assert.Nil(t, sketch)
		assert.Error(t, err)
		_, ok := err.(*InvalidSketchFolderNameError)
		assert.False(t, ok)
		sketchPath, _ = sketchPath.Parent().Abs()
		expectedError := fmt.Sprintf("main file missing from sketch: %s", sketchPath.Join(sketchPath.Base()+".ino"))
		assert.EqualError(t, err, expectedError)
	}
	{
		sketchPath := paths.New("testdata", "non-existent")
		sketch, skerr := New(sketchPath)
		require.Nil(t, sketch)
		require.Error(t, skerr)
		_, ok := skerr.(*InvalidSketchFolderNameError)
		assert.False(t, ok)
		sketchPath, _ = sketchPath.Abs()
		expectedError := fmt.Sprintf("no such file or directory: %s", sketchPath)
		require.EqualError(t, skerr, expectedError)
	}
}

func TestNewSketchCasingCorrect(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchCasingCorrect")
	mainFilePath := sketchPath.Join("SketchCasingCorrect.ino")
	sketch, err := New(sketchPath)
	assert.NotNil(t, sketch)
	assert.NoError(t, err)
	assert.True(t, sketchPath.EquivalentTo(sketch.FullPath))
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	assert.Equal(t, sketch.AdditionalFiles.Len(), 0)
	assert.Equal(t, sketch.RootFolderFiles.Len(), 0)
}

func TestSketchWithMarkdownAsciidocJson(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchWithMarkdownAsciidocJson")
	mainFilePath := sketchPath.Join("SketchWithMarkdownAsciidocJson.ino")
	adocFilePath := sketchPath.Join("foo.adoc")
	jsonFilePath := sketchPath.Join("foo.json")
	mdFilePath := sketchPath.Join("foo.md")

	sketch, err := New(sketchPath)
	assert.NotNil(t, sketch)
	assert.NoError(t, err)
	assert.True(t, sketchPath.EquivalentTo(sketch.FullPath))
	assert.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	assert.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	require.Equal(t, sketch.AdditionalFiles.Len(), 3)
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(adocFilePath))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(jsonFilePath))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(mdFilePath))
	assert.Equal(t, sketch.RootFolderFiles.Len(), 3)
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(adocFilePath))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(jsonFilePath))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(mdFilePath))
}

func TestSketchWithTppFile(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchWithTppFile")
	mainFilePath := sketchPath.Join("SketchWithTppFile.ino")
	templateFile := sketchPath.Join("template.tpp")

	sketch, err := New(sketchPath)
	require.NotNil(t, sketch)
	require.NoError(t, err)
	require.True(t, sketchPath.EquivalentTo(sketch.FullPath))
	require.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	require.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	require.Equal(t, sketch.AdditionalFiles.Len(), 1)
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(templateFile))
	require.Equal(t, sketch.RootFolderFiles.Len(), 1)
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(templateFile))
}

func TestSketchWithIppFile(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchWithIppFile")
	mainFilePath := sketchPath.Join("SketchWithIppFile.ino")
	templateFile := sketchPath.Join("template.ipp")

	sketch, err := New(sketchPath)
	require.NotNil(t, sketch)
	require.NoError(t, err)
	require.True(t, sketchPath.EquivalentTo(sketch.FullPath))
	require.True(t, mainFilePath.EquivalentTo(sketch.MainFile))
	require.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	require.Equal(t, sketch.AdditionalFiles.Len(), 1)
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(templateFile))
	require.Equal(t, sketch.RootFolderFiles.Len(), 1)
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(templateFile))
}

func TestNewSketchFolderSymlink(t *testing.T) {
	// pass the path to the sketch folder
	sketchName := "SketchSymlink"
	sketchPath, _ := paths.New("testdata", fmt.Sprintf("%sSrc", sketchName)).Abs()
	sketchPathSymlink, _ := paths.New("testdata", sketchName).Abs()
	os.Symlink(sketchPath.String(), sketchPathSymlink.String())
	defer sketchPathSymlink.Remove()

	mainFilePath := sketchPathSymlink.Join(fmt.Sprintf("%sSrc.ino", sketchName))
	sketch, err := New(sketchPathSymlink)
	require.Nil(t, err)
	require.NotNil(t, sketch)
	require.True(t, sketch.MainFile.EquivalentTo(mainFilePath))
	require.True(t, sketch.FullPath.EquivalentTo(sketchPath))
	require.True(t, sketch.FullPath.EquivalentTo(sketchPathSymlink))
	require.Equal(t, sketch.OtherSketchFiles.Len(), 2)
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPath.Join("old.pde")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPath.Join("other.ino")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPathSymlink.Join("old.pde")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPathSymlink.Join("other.ino")))
	require.Equal(t, sketch.AdditionalFiles.Len(), 3)
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("header.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("s_file.S")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("src", "helper.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("header.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("s_file.S")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("src", "helper.h")))
	require.Equal(t, sketch.RootFolderFiles.Len(), 4)
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("header.h")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("old.pde")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("other.ino")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("s_file.S")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("header.h")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("old.pde")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("other.ino")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("s_file.S")))

	// pass the path to the main file
	sketch, err = New(mainFilePath)
	require.Nil(t, err)
	require.NotNil(t, sketch)
	require.True(t, sketch.MainFile.EquivalentTo(mainFilePath))
	require.True(t, sketch.FullPath.EquivalentTo(sketchPath))
	require.True(t, sketch.FullPath.EquivalentTo(sketchPathSymlink))
	require.Equal(t, sketch.OtherSketchFiles.Len(), 2)
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPath.Join("old.pde")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPath.Join("other.ino")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPathSymlink.Join("old.pde")))
	require.True(t, sketch.OtherSketchFiles.ContainsEquivalentTo(sketchPathSymlink.Join("other.ino")))
	require.Equal(t, sketch.AdditionalFiles.Len(), 3)
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("header.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("s_file.S")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPath.Join("src", "helper.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("header.h")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("s_file.S")))
	require.True(t, sketch.AdditionalFiles.ContainsEquivalentTo(sketchPathSymlink.Join("src", "helper.h")))
	require.Equal(t, sketch.RootFolderFiles.Len(), 4)
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("header.h")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("old.pde")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("other.ino")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPath.Join("s_file.S")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("header.h")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("old.pde")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("other.ino")))
	require.True(t, sketch.RootFolderFiles.ContainsEquivalentTo(sketchPathSymlink.Join("s_file.S")))
}

func TestGenBuildPath(t *testing.T) {
	want := paths.TempDir().Join("arduino/arduino-sketch-ACBD18DB4CC2F85CEDEF654FCCC4A4D8")
	assert.True(t, GenBuildPath(paths.New("foo")).EquivalentTo(want))

	want = paths.TempDir().Join("arduino/arduino-sketch-D41D8CD98F00B204E9800998ECF8427E")
	assert.True(t, GenBuildPath(nil).EquivalentTo(want))
}

func TestCheckForPdeFiles(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchSimple")
	files := CheckForPdeFiles(sketchPath)
	require.Empty(t, files)

	sketchPath = paths.New("testdata", "SketchPde")
	files = CheckForPdeFiles(sketchPath)
	require.Len(t, files, 1)
	require.Equal(t, sketchPath.Join("SketchPde.pde"), files[0])

	sketchPath = paths.New("testdata", "SketchMultipleMainFiles")
	files = CheckForPdeFiles(sketchPath)
	require.Len(t, files, 1)
	require.Equal(t, sketchPath.Join("SketchMultipleMainFiles.pde"), files[0])

	sketchPath = paths.New("testdata", "SketchSimple", "SketchSimple.ino")
	files = CheckForPdeFiles(sketchPath)
	require.Empty(t, files)

	sketchPath = paths.New("testdata", "SketchPde", "SketchPde.pde")
	files = CheckForPdeFiles(sketchPath)
	require.Len(t, files, 1)
	require.Equal(t, sketchPath.Parent().Join("SketchPde.pde"), files[0])

	sketchPath = paths.New("testdata", "SketchMultipleMainFiles", "SketchMultipleMainFiles.ino")
	files = CheckForPdeFiles(sketchPath)
	require.Len(t, files, 1)
	require.Equal(t, sketchPath.Parent().Join("SketchMultipleMainFiles.pde"), files[0])

	sketchPath = paths.New("testdata", "SketchMultipleMainFiles", "SketchMultipleMainFiles.pde")
	files = CheckForPdeFiles(sketchPath)
	require.Len(t, files, 1)
	require.Equal(t, sketchPath.Parent().Join("SketchMultipleMainFiles.pde"), files[0])
}

func TestNewSketchWithSymlink(t *testing.T) {
	sketchPath, _ := paths.New("testdata", "SketchWithSymlink").Abs()
	mainFilePath := sketchPath.Join("SketchWithSymlink.ino")
	helperFilePath := sketchPath.Join("some_folder", "helper.h")
	helperFileSymlinkPath := sketchPath.Join("src", "helper.h")
	srcPath := sketchPath.Join("src")

	// Create a symlink in the Sketch folder
	os.Symlink(sketchPath.Join("some_folder").String(), srcPath.String())
	defer srcPath.Remove()

	sketch, err := New(sketchPath)
	require.NoError(t, err)
	require.NotNil(t, sketch)
	require.True(t, sketch.MainFile.EquivalentTo(mainFilePath))
	require.True(t, sketch.FullPath.EquivalentTo(sketchPath))
	require.Equal(t, sketch.OtherSketchFiles.Len(), 0)
	require.Equal(t, sketch.AdditionalFiles.Len(), 2)
	require.True(t, sketch.AdditionalFiles.Contains(helperFilePath))
	require.True(t, sketch.AdditionalFiles.Contains(helperFileSymlinkPath))
	require.Equal(t, sketch.RootFolderFiles.Len(), 0)
}

func TestNewSketchWithSymlinkLoop(t *testing.T) {
	sketchPath, _ := paths.New("testdata", "SketchWithSymlinkLoop").Abs()
	someSymlinkPath := sketchPath.Join("some_folder", "some_symlink")

	// Create a recursive Sketch symlink
	err := os.Symlink(sketchPath.String(), someSymlinkPath.String())
	require.NoErrorf(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")
	defer someSymlinkPath.Remove()

	// The failure condition is New() never returning, testing for which requires setting up a timeout.
	done := make(chan bool)
	var sketch *Sketch
	go func() {
		sketch, err = New(sketchPath)
		done <- true
	}()

	assert.Eventually(
		t,
		func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		},
		20*time.Second,
		10*time.Millisecond,
		"Infinite symlink loop while loading sketch",
	)
	require.Error(t, err)
	require.Nil(t, sketch)
}

func TestSketchWithMultipleSymlinkLoops(t *testing.T) {
	sketchPath := paths.New("testdata", "SketchWithMultipleSymlinkLoops")
	srcPath := sketchPath.Join("src")
	srcPath.Mkdir()
	defer srcPath.RemoveAll()

	firstSymlinkPath := srcPath.Join("UpGoer1")
	secondSymlinkPath := srcPath.Join("UpGoer2")
	err := os.Symlink("..", firstSymlinkPath.String())
	require.NoErrorf(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")
	err = os.Symlink("..", secondSymlinkPath.String())
	require.NoErrorf(t, err, "This test must be run as administrator on Windows to have symlink creation privilege.")

	// The failure condition is New() never returning, testing for which requires setting up a timeout.
	done := make(chan bool)
	var sketch *Sketch
	go func() {
		sketch, err = New(sketchPath)
		done <- true
	}()

	assert.Eventually(
		t,
		func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		},
		20*time.Second,
		10*time.Millisecond,
		"Infinite symlink loop while loading sketch",
	)
	require.Error(t, err)
	require.Nil(t, sketch)
}
