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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/types"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

func getContext() *types.Context {
	fqbn, _ := cores.ParseFQBN("foo:bar:baz")
	return &types.Context{
		HardwareDirs: paths.PathList{
			paths.New("path1"),
			paths.New("path2"),
		},
		ToolsDirs: paths.PathList{
			paths.New("path3"),
			paths.New("path4"),
		},
		BuiltInLibrariesDirs: paths.PathList{
			paths.New("path5"),
			paths.New("path6"),
		},
		OtherLibrariesDirs: paths.PathList{
			paths.New("path7"),
			paths.New("path8"),
		},
		SketchLocation: paths.New("/foo/bar/sketch"),
		Sketch: &types.Sketch{
			AdditionalFiles: []types.SketchFile{
				types.SketchFile{
					Name: paths.New("/foo/bar/baz"),
				},
			},
		},
		FQBN:                  fqbn,
		ArduinoAPIVersion:     "0.99",
		CustomBuildProperties: []string{"a", "b"},
	}
}

func TestContextExtractBuildOptions(t *testing.T) {
	c := getContext()
	options := c.ExtractBuildOptions()

	var cases = []struct {
		key  string
		want string
	}{
		{"hardwareFolders", "path1,path2"},
		{"toolsFolders", "path3,path4"},
		{"builtInLibrariesFolders", "path5,path6"},
		{"otherLibrariesFolders", "path7,path8"},
		{"sketchLocation", "/foo/bar/sketch"},
		{"additionalFiles", ".."},
		{"fqbn", "foo:bar:baz"},
		{"runtime.ide.version", "0.99"},
		{"customBuildProperties", "a,b"},
	}

	for _, cs := range cases {
		assert.Equal(t, cs.want, options.Get(cs.key))
	}

}

func TestInjectBuildOptions(t *testing.T) {
	c := getContext()
	options := c.ExtractBuildOptions()

	var cases = []struct {
		key  string
		want string
	}{
		{"hardwareFolders", "foo,bar"},
		{"toolsFolders", "snafu,fud"},
		{"builtInLibrariesFolders", "baz,bar"},
		{"otherLibrariesFolders", "bar,foo"},
		{"sketchLocation", "/path/to/sketch"},
		{"fqbn", "bar:foo:baz"},
		{"runtime.ide.version", "99.99"},
		{"customBuildProperties", "a,b"},
	}

	// alter options
	for _, cs := range cases {
		options.Set(cs.key, cs.want)
	}

	// inject altered options
	c.InjectBuildOptions(options)

	// ensure results are what we want
	optionsAfter := c.ExtractBuildOptions()
	for _, cs := range cases {
		assert.Equal(t, cs.want, optionsAfter.Get(cs.key))
	}
}

func TestGetLogger(t *testing.T) {
	c := getContext()
	assert.IsType(t, new(i18n.HumanLogger), c.GetLogger())
}

func TestSetLogger(t *testing.T) {
	c := getContext()
	l := new(i18n.LoggerToCustomStreams)
	c.SetLogger(l)
	assert.Equal(t, l, c.GetLogger())
}
