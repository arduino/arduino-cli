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

package sketches

import (
	"fmt"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestSketchLoadingFromFolderOrMainFile(t *testing.T) {
	skFolder := paths.New("testdata/Sketch1")
	skMainIno := skFolder.Join("Sketch1.ino")

	{
		sk, err := NewSketchFromPath(skFolder)
		require.NoError(t, err)
		require.Equal(t, sk.Name, "Sketch1")
		fmt.Println(sk.FullPath.String(), "==", skFolder.String())
		require.True(t, sk.FullPath.EquivalentTo(skFolder))
	}

	{
		sk, err := NewSketchFromPath(skMainIno)
		require.NoError(t, err)
		require.Equal(t, sk.Name, "Sketch1")
		fmt.Println(sk.FullPath.String(), "==", skFolder.String())
		require.True(t, sk.FullPath.EquivalentTo(skFolder))
	}
}

func TestSketchBuildPath(t *testing.T) {
	// Verifies build path is returned if sketch path is set
	sketchPath := paths.New("testdata/Sketch1")
	sketch, err := NewSketchFromPath(sketchPath)
	require.NoError(t, err)
	buildPath, err := sketch.BuildPath()
	require.NoError(t, err)
	require.Contains(t, buildPath.String(), "arduino-sketch-")

	// Verifies sketch path is returned if sketch has .pde extension
	sketchPath = paths.New("testdata", "SketchPde")
	sketch, err = NewSketchFromPath(sketchPath)
	require.NoError(t, err)
	require.NotNil(t, sketch)
	buildPath, err = sketch.BuildPath()
	require.NoError(t, err)
	require.Contains(t, buildPath.String(), "arduino-sketch-")

	// Verifies error is returned if there are multiple main files
	sketchPath = paths.New("testdata", "SketchMultipleMainFiles")
	sketch, err = NewSketchFromPath(sketchPath)
	require.Nil(t, sketch)
	require.Error(t, err, "multiple main sketch files found")

	// Verifies error is returned if sketch path is not set
	sketch = &Sketch{}
	buildPath, err = sketch.BuildPath()
	require.Nil(t, buildPath)
	require.Error(t, err, "sketch path is empty")
}

func TestCheckForPdeFiles(t *testing.T) {
	sketchPath := paths.New("testdata", "Sketch1")
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

	sketchPath = paths.New("testdata", "Sketch1", "Sketch1.ino")
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

func TestSketchLoadWithCasing(t *testing.T) {
	sketchFolder := paths.New("testdata", "SketchCasingWrong")

	sketch, err := NewSketchFromPath(sketchFolder)
	require.Nil(t, sketch)

	sketchFolderAbs, _ := sketchFolder.Abs()
	sketchMainFileAbs := sketchFolderAbs.Join("SketchCasingWrong.ino")
	expectedError := fmt.Sprintf("no valid sketch found in %s: missing %s", sketchFolderAbs, sketchMainFileAbs)
	require.EqualError(t, err, expectedError)
}

func TestSketchLoadingCorrectCasing(t *testing.T) {
	sketchFolder := paths.New("testdata", "SketchCasingCorrect")
	sketch, err := NewSketchFromPath(sketchFolder)
	require.NotNil(t, sketch)
	require.NoError(t, err)
	require.Equal(t, sketch.Name, "SketchCasingCorrect")
	require.True(t, sketch.FullPath.EquivalentTo(sketchFolder))
}
