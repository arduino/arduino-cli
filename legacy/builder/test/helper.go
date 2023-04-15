// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
// Copyright 2015 Matthijs Kooijman
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

package test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func LoadAndInterpolate(t *testing.T, filename string, ctx *types.Context) string {
	funcsMap := template.FuncMap{
		"QuoteCppString": utils.QuoteCppPath,
	}

	tpl, err := template.New(filepath.Base(filename)).Funcs(funcsMap).ParseFiles(filename)
	NoError(t, err)

	var buf bytes.Buffer
	data := make(map[string]interface{})
	data["sketch"] = ctx.Sketch
	err = tpl.Execute(&buf, data)
	NoError(t, err)

	return buf.String()
}

func Abs(t *testing.T, rel *paths.Path) *paths.Path {
	absPath, err := rel.Abs()
	NoError(t, err)
	return absPath
}

func NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if !assert.NoError(t, err, msgAndArgs...) {
		fmt.Printf("%+v\n", err) // Outputs stack trace in case of wrapped error
		t.FailNow()
	}
}

func SetupBuildPath(t *testing.T, ctx *types.Context) *paths.Path {
	buildPath, err := paths.MkTempDir("", "test_build_path")
	NoError(t, err)
	ctx.BuildPath = buildPath
	return buildPath
}

func SetupBuildCachePath(t *testing.T, ctx *types.Context) *paths.Path {
	buildCachePath, err := paths.MkTempDir(constants.EMPTY_STRING, "test_build_cache")
	NoError(t, err)
	ctx.CoreBuildCachePath = buildCachePath
	return buildCachePath
}

func parseFQBN(t *testing.T, fqbnIn string) *cores.FQBN {
	fqbn, err := cores.ParseFQBN(fqbnIn)
	require.NoError(t, err)
	return fqbn
}

func OpenSketch(t *testing.T, sketchPath *paths.Path) *sketch.Sketch {
	sketch, err := sketch.New(sketchPath)
	require.NoError(t, err)
	return sketch
}

type ByLibraryName []*libraries.Library

func (s ByLibraryName) Len() int {
	return len(s)
}
func (s ByLibraryName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByLibraryName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
