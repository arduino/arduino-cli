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
 */

package builder

import (
	"os/exec"
	"strings"

	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-paths-helper"
)

func GCCPreprocRunner(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) error {
	cmd, err := prepareGCCPreprocRecipeProperties(ctx, sourceFilePath, targetFilePath, includes)
	if err != nil {
		return i18n.WrapError(err)
	}

	_, _, err = utils.ExecCommand(ctx, cmd /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
	if err != nil {
		return i18n.WrapError(err)
	}

	return nil
}

func GCCPreprocRunnerForDiscoveringIncludes(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) ([]byte, error) {
	cmd, err := prepareGCCPreprocRecipeProperties(ctx, sourceFilePath, targetFilePath, includes)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	_, stderr, err := utils.ExecCommand(ctx, cmd /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Capture)
	if err != nil {
		return stderr, i18n.WrapError(err)
	}

	return stderr, nil
}

func prepareGCCPreprocRecipeProperties(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) (*exec.Cmd, error) {
	properties := ctx.BuildProperties.Clone()
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, sourceFilePath)
	properties.SetPath(constants.BUILD_PROPERTIES_PREPROCESSED_FILE_PATH, targetFilePath)

	includesStrings := utils.Map(includes.AsStrings(), utils.WrapWithHyphenI)
	properties.Set(constants.BUILD_PROPERTIES_INCLUDES, strings.Join(includesStrings, constants.SPACE))

	if properties.Get(constants.RECIPE_PREPROC_MACROS) == "" {
		//generate PREPROC_MACROS from RECIPE_CPP_PATTERN
		properties.Set(constants.RECIPE_PREPROC_MACROS, GeneratePreprocPatternFromCompile(properties.Get(constants.RECIPE_CPP_PATTERN)))
	}

	cmd, err := builder_utils.PrepareCommandForRecipe(ctx, properties, constants.RECIPE_PREPROC_MACROS, true)
	if err != nil {
		return nil, i18n.WrapError(err)
	}

	// Remove -MMD argument if present. Leaving it will make gcc try
	// to create a /dev/null.d dependency file, which won't work.
	cmd.Args = utils.Filter(cmd.Args, func(a string) bool { return a != "-MMD" })

	return cmd, nil
}

func GeneratePreprocPatternFromCompile(compilePattern string) string {
	// add {preproc.macros.flags}
	// replace "{object_file}" with "{preprocessed_file_path}"
	returnString := compilePattern
	returnString = strings.Replace(returnString, "{compiler.cpp.flags}", "{compiler.cpp.flags} {preproc.macros.flags}", 1)
	returnString = strings.Replace(returnString, "{object_file}", "{preprocessed_file_path}", 1)
	return returnString
}
