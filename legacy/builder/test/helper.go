/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 * Copyright 2015 Matthijs Kooijman
 */

package test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
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
		switch err.(type) {
		case *errors.Error:
			fmt.Println(err.(*errors.Error).ErrorStack())
		}
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
	ctx.BuildCachePath = buildCachePath
	return buildCachePath
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
