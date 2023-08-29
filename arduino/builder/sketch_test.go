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

package builder

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestMergeSketchSources(t *testing.T) {
	// borrow the sketch from TestLoadSketchFolder to avoid boilerplate
	sk, err := sketch.New(paths.New("testdata", "TestLoadSketchFolder"))
	require.Nil(t, err)
	require.NotNil(t, sk)

	// load expected result
	suffix := ".txt"
	if runtime.GOOS == "windows" {
		suffix = "_win.txt"
	}
	mergedPath := paths.New("testdata", t.Name()+suffix)
	require.NoError(t, mergedPath.ToAbs())
	mergedBytes, err := mergedPath.ReadFile()
	require.NoError(t, err, "reading golden file %s: %v", mergedPath, err)

	pathToGoldenSource := mergedPath.Parent().Parent().String()
	if runtime.GOOS == "windows" {
		pathToGoldenSource = strings.ReplaceAll(pathToGoldenSource, `\`, `\\`)
	}
	mergedSources := strings.ReplaceAll(string(mergedBytes), "%s", pathToGoldenSource)

	b := NewBuilder(sk)
	offset, source, err := b.sketchMergeSources(nil)
	require.Nil(t, err)
	require.Equal(t, 2, offset)
	require.Equal(t, mergedSources, source)
}

func TestMergeSketchSourcesArduinoIncluded(t *testing.T) {
	sk, err := sketch.New(paths.New("testdata", t.Name()))
	require.Nil(t, err)
	require.NotNil(t, sk)

	// ensure not to include Arduino.h when it's already there
	b := NewBuilder(sk)
	_, source, err := b.sketchMergeSources(nil)
	require.Nil(t, err)
	require.Equal(t, 1, strings.Count(source, "<Arduino.h>"))
}

func TestCopyAdditionalFiles(t *testing.T) {
	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()

	// load the golden sketch
	sk1, err := sketch.New(paths.New("testdata", t.Name()))
	require.Nil(t, err)
	require.Equal(t, sk1.AdditionalFiles.Len(), 1)

	// copy the sketch over, create a fake main file we don't care about it
	// but we need it for `SketchLoad` to succeed later
	err = sketchCopyAdditionalFiles(sk1, tmp, nil)
	require.Nil(t, err)
	fakeIno := tmp.Join(fmt.Sprintf("%s.ino", tmp.Base()))
	require.Nil(t, fakeIno.WriteFile([]byte{}))

	// compare
	sk2, err := sketch.New(tmp)
	require.Nil(t, err)
	require.Equal(t, sk2.AdditionalFiles.Len(), 1)

	// save file info
	info1, err := sk2.AdditionalFiles[0].Stat()
	require.Nil(t, err)

	// copy again
	err = sketchCopyAdditionalFiles(sk1, tmp, nil)
	require.Nil(t, err)

	// verify file hasn't changed
	info2, err := sk2.AdditionalFiles[0].Stat()
	require.NoError(t, err)
	require.Equal(t, info1.ModTime(), info2.ModTime())
}
